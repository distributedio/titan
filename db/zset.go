package db

import (
    "strconv"
    "encoding/binary"
)

// ZSetMeta is the meta data of the sorted set
type ZSetMeta struct {
    Object
    Len int64
}

// ZSet implements the the sorted set
type ZSet struct {
    meta ZSetMeta
    key  []byte
    txn  *Transaction
}

type MemberScore struct {
    Member string
    Score float64
}

// GetZSet returns a sorted set, create new one if don't exists
func GetZSet(txn *Transaction, key []byte) (*ZSet, error) {
    zset := &ZSet{txn: txn, key: key}

    mkey := MetaKey(txn.db, key)
    meta, err := txn.t.Get(mkey)
    if err != nil {
        if IsErrNotFound(err) {
            now := Now()
            zset.meta.CreatedAt = now
            zset.meta.UpdatedAt = now
            zset.meta.ExpireAt = 0
            zset.meta.ID = UUID()
            zset.meta.Type = ObjectZSet
            zset.meta.Encoding = ObjectEncodingHT
            zset.meta.Len = 0
            return zset, nil
        }
        return nil, err
    }
    if err := zset.decodeMeta(meta); err != nil {
        return nil, err
    }
    if zset.meta.Type != ObjectZSet {
        return nil, ErrTypeMismatch
    }
    return zset, nil
}

func (zset *ZSet) ZAdd(members [][]byte, scores []float64) (int64, error) {
    added := int64(0)

    oldValues := make([][]byte, len(members))
    var err error
    if zset.meta.Len > 0 {
        oldValues, err = zset.MGet(members)
        if err != nil {
            return 0, err
        }
    }

    dkey := DataKey(zset.txn.db, zset.meta.ID)
    scorePrefix := ZSetScorePrefix(zset.txn.db, zset.meta.ID)
    var found bool
    for i := range members {
        found = false
        if oldValues[i] != nil {
            found = true
            oldScore := DecodeFloat64(oldValues[i])
            if scores[i] == oldScore {
                continue
            }
            oldScoreKey := zsetScoreKey(scorePrefix, oldValues[i], members[i])
            if err = zset.txn.t.Delete(oldScoreKey); err != nil {
                return added, err
            }
        }
        memberKey := zsetMemberKey(dkey, members[i])
        bytesScore := EncodeFloat64(scores[i])
        if err = zset.txn.t.Set(memberKey, bytesScore); err != nil {
            return added, err
        }

        scoreKey := zsetScoreKey(scorePrefix, bytesScore, members[i])
        if err = zset.txn.t.Set(scoreKey, NilValue); err != nil {
            return added, err
        }

        if !found {
            added += 1
        }
    }

    zset.meta.Len += added
    if err = zset.updateMeta(); err != nil {
        return 0, err
    }
    return added, nil
}

func (zset *ZSet) MGet(members [][]byte) ([][]byte, error) {
    ikeys := make([][]byte, len(members))
    dkey := DataKey(zset.txn.db, zset.meta.ID)
    for i := range members {
        ikeys[i] = zsetMemberKey(dkey, members[i])
    }

    return BatchGetValues(zset.txn, ikeys)
}

func (zset *ZSet) updateMeta() error {
    meta := zset.encodeMeta(zset.meta)
    return zset.txn.t.Set(MetaKey(zset.txn.db, zset.key), meta)
}

func (zset *ZSet) encodeMeta(meta ZSetMeta) []byte {
    b := EncodeObject(&meta.Object)
    m := make([]byte, 8)
    binary.BigEndian.PutUint64(m[:8], uint64(meta.Len))
    return append(b, m...)
}

//decodeMeta if obj has been existed , stop parse
func (zset *ZSet) decodeMeta(b []byte) error {
    obj, err := DecodeObject(b)
    if err != nil {
        return err
    }
    zset.meta.Object = *obj

    m := b[ObjectEncodingLength:]
    if len(m) != 8 {
        return ErrInvalidLength
    }
    zset.meta.Len = int64(binary.BigEndian.Uint64(m[:8]))
    return nil
}

func (zset *ZSet) Exist() bool {
    if zset.meta.Len == 0 {
        return false
    }
    return true
}

func (zset *ZSet) ZAnyOrderRange(start int64, stop int64, withScore bool, positiveOrder bool) ([][]byte, error) {
    if stop < 0 {
        if stop = zset.meta.Len + stop; stop < 0 {
            return [][]byte{}, nil
        }
    }else if stop >= zset.meta.Len{
        stop = zset.meta.Len -1
    }
    if start < 0 {
        if start = zset.meta.Len + start; start < 0 {
            start = 0
        }
    }
    // return 0 elements
    if start > stop || start >= zset.meta.Len{
        return [][]byte{}, nil
    }

    scorePrefix := ZSetScorePrefix(zset.txn.db, zset.meta.ID)
    var iter Iterator
    var err error
    if positiveOrder {
        iter, err = zset.txn.t.Seek(scorePrefix)
    } else {
        //tikv sdk didn't implement SeekReverse(), so we just use seek() to implement zrevrange now
        //because tikv sdk scan 256 keys in next(), for the zset which have <=256 members,
        // its performance should be similar with seekReverse, for >256 zset, it should be bad
        //iter, err = zset.txn.t.SeekReverse(scorePrefix)
        iter, err = zset.txn.t.Seek(scorePrefix)
        tmp := start
        start = zset.meta.Len-1-stop
        stop = zset.meta.Len-1-tmp
    }

    if err != nil {
        return nil, err
    }

    var items [][]byte
    var scoreAndMember []byte
    var member, score []byte
    for i := int64(0); i <= stop && err == nil && iter.Valid() && iter.Key().HasPrefix(scorePrefix); err = iter.Next()  {
        if i >= start {
            score = nil
            member = nil
            scoreAndMember = iter.Key()[len(scorePrefix)+1:]
            for pos, c := range scoreAndMember {
                if c == ':' {
                    score = scoreAndMember[0:pos]
                    member = scoreAndMember[pos+1:]
                    break
                }
            }
            items = append(items, member)
            if withScore {
                val := []byte(strconv.FormatFloat(DecodeFloat64(score), 'f', -1, 64))
                items = append(items, val)
                if !positiveOrder  {
                    items[len(items)-1], items[len(items)-2] = items[len(items)-2], items[len(items)-1]
                }
            }
        }
        i++
    }
    if !positiveOrder {
        for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
            items[i], items[j] = items[j], items[i]
        }
    }

    return items, nil
}
func (zset *ZSet) ZRem(members [][]byte) (int64, error) {
    deleted := int64(0)

    scores, err := zset.MGet(members)
    if err != nil {
        return 0, err
    }

    dkey := DataKey(zset.txn.db, zset.meta.ID)
    scorePrefix := ZSetScorePrefix(zset.txn.db, zset.meta.ID)
    for i := range members{
        if scores[i] == nil {
            continue
        }

        scoreKey := zsetScoreKey(scorePrefix, scores[i], members[i])
        if err = zset.txn.t.Delete(scoreKey); err != nil {
            return deleted, err
        }

        memberKey := zsetMemberKey(dkey, members[i])
        if err = zset.txn.t.Delete(memberKey); err != nil {
            return deleted, err
        }

        deleted += 1
    }
    zset.meta.Len -= deleted

    if zset.meta.Len == 0 {
        mkey := MetaKey(zset.txn.db, zset.key)
        if err = zset.txn.t.Delete(mkey); err != nil {
            return deleted, err
        }
        if zset.meta.Object.ExpireAt > 0 {
            if err := unExpireAt(zset.txn.t, mkey, zset.meta.Object.ExpireAt); err != nil {
                return deleted, err
            }
        }
        return deleted, nil
    }

    return deleted, zset.updateMeta()
}

func zsetMemberKey(dkey []byte, member []byte) []byte {
    dkey = append(dkey, ':')
    return append(dkey, member...)
}

// ZSetScorePrefix builds a score key prefix from a redis key
func ZSetScorePrefix(db *DB, key []byte) []byte {
    var sPrefix []byte
    sPrefix = append(sPrefix, []byte(db.Namespace)...)
    sPrefix = append(sPrefix, ':')
    sPrefix = append(sPrefix, db.ID.Bytes()...)
    sPrefix = append(sPrefix, ':', 'S', ':')
    sPrefix = append(sPrefix, key...)
    return sPrefix
}

func zsetScoreKey(scorePrefix []byte, score []byte, member []byte) []byte {
    scoreKey := append(scorePrefix, ':')
    scoreKey = append(scoreKey, score...)
    scoreKey = append(scoreKey, ':')
    scoreKey = append(scoreKey, member...)
    return scoreKey
}
