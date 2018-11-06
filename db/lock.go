package db

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"log"
	"time"
)

const lockKeyPrefix = "$sys:lock:"

// lock a key with lease
type lock struct {
	id    uint64
	key   []byte // key can be nil, then we will lock on "$sys:lock:"
	s     *RedisStore
	lease time.Duration
}

func lockKey(key []byte) []byte {
	var buf []byte
	buf = append(buf, []byte(lockKeyPrefix)...)
	buf = append(buf, key...)
	return buf
}

func acquire(s *RedisStore, key []byte, lease time.Duration) (*lock, error) {
	l := &lock{key: key, s: s, lease: lease}
	l.id = binary.LittleEndian.Uint64(UUID())

	lkey := lockKey(l.key)
	for {
		log.Println("acquire lock ", string(key))
		now := time.Now()
		lockValue := struct {
			LockID   uint64
			ExpireAt time.Time
		}{LockID: l.id, ExpireAt: now.Add(l.lease)}
		lval, err := json.Marshal(&lockValue)
		if err != nil {
			return nil, err
		}

		txn, err := l.s.Begin()
		if err != nil {
			return nil, err
		}
		val, err := txn.Get(lkey)
		if err != nil {
			if !IsErrNotFound(err) {
				txn.Rollback()
				return nil, err
			}
			if err := txn.Set(lkey, lval); err != nil {
				txn.Rollback()
				return nil, err
			}
			if err := txn.Commit(context.Background()); err != nil {
				txn.Rollback()
				return nil, err
			}
			return l, nil
		}

		if err := json.Unmarshal(val, &lockValue); err != nil {
			txn.Rollback()
			return nil, err
		}

		// lock has been held by others but exired, or lock is myself
		if (lockValue.LockID != l.id &&
			lockValue.ExpireAt.Sub(now) < 0) || lockValue.LockID == l.id {
			// lock has expired, grab it
			if err := txn.Set(lkey, lval); err != nil {
				txn.Rollback()
				return nil, err
			}
			if err := txn.Commit(context.Background()); err != nil {
				txn.Rollback()
				time.Sleep(2 * time.Second)
				continue
			}
			return l, nil
		}
		txn.Rollback()
		time.Sleep(2 * time.Second)
		continue
	}
}

// renew the lease
func (l *lock) renew() error {
	lkey := lockKey(l.key)
	for {
		log.Println("renew lock ", string(l.key))
		now := time.Now()
		lockValue := struct {
			LockID   uint64
			ExpireAt time.Time
		}{LockID: l.id, ExpireAt: now.Add(l.lease)}

		lval, err := json.Marshal(&lockValue)
		if err != nil {
			return err
		}

		txn, err := l.s.Begin()
		if err != nil {
			return err
		}
		val, err := txn.Get(lkey)
		if err != nil {
			txn.Rollback()
			if !IsErrNotFound(err) {
				return err
			}
			if err := txn.Set(lkey, lval); err != nil {
				txn.Rollback()
				return err
			}
			if err := txn.Commit(context.Background()); err != nil {
				txn.Rollback()
				return err
			}
		}

		if err := json.Unmarshal(val, &lockValue); err != nil {
			txn.Rollback()
			return err
		}
		if lockValue.LockID == l.id {
			if err := txn.Set(lkey, lval); err != nil {
				txn.Rollback()
				return err
			}
			if err := txn.Commit(context.Background()); err != nil {
				txn.Rollback()
				time.Sleep(2 * time.Second)
				continue
			}
			return nil
		}
		txn.Rollback()
		time.Sleep(2 * time.Second)
	}
	return nil
}

func (l *lock) release() error {
	lockValue := struct {
		LockID   uint64
		ExpireAt time.Time
	}{}

	lkey := lockKey(l.key)

	txn, err := l.s.Begin()
	if err != nil {
		return err
	}
	val, err := txn.Get(lkey)
	if err != nil {
		if !IsErrNotFound(err) {
			txn.Rollback()
			return nil
		}
		return err
	}
	if err := json.Unmarshal(val, &lockValue); err != nil {
		txn.Rollback()
		return err
	}
	if lockValue.LockID == l.id {
		if err := txn.Delete(lkey); err != nil {
			txn.Rollback()
			return err
		}
		if err := txn.Commit(context.Background()); err != nil {
			txn.Rollback()
			return err
		}
	}
	txn.Rollback()
	return nil
}
