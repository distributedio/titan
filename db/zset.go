package db

import (
    "encoding/json"
    "strconv"
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
            zset.meta.Type = ObjectZset
            zset.meta.Encoding = ObjectEncodingHT
            zset.meta.Len = 0
            return zset, nil
        }
        return nil, err
    }
    if err := json.Unmarshal(meta, &zset.meta); err != nil {
        return nil, err
    }
    if zset.meta.Type != ObjectZset {
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
    meta, err := json.Marshal(zset.meta)
    if err != nil {
        return err
    }
    return zset.txn.t.Set(MetaKey(zset.txn.db, zset.key), meta)
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
        iter, err = zset.txn.t.SeekReverse(scorePrefix)
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
            }

        }
        i++
    }

    return items, nil
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
