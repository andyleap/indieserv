package main

import (
	"encoding/binary"
	"encoding/json"

	"github.com/andyleap/gopub"
	"github.com/boltdb/bolt"
)

type SubStorage struct {
	b *Blog
}

func (ss SubStorage) AddSub(id int64, sub *gopub.Subscription) {
	ss.b.db.Update(func(tx *bolt.Tx) error {
		subs := tx.Bucket([]byte("subs"))
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(id))
		data, _ := json.Marshal(sub)
		subs.Put(b, data)
		return nil
	})
}

func (ss SubStorage) RemoveSub(id int64) {
	ss.b.db.Update(func(tx *bolt.Tx) error {
		subs := tx.Bucket([]byte("subs"))
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(id))
		subs.Delete(b)
		return nil
	})
}

func (ss SubStorage) GetSubs() map[int64]*gopub.Subscription {
	subs := make(map[int64]*gopub.Subscription)
	ss.b.db.View(func(tx *bolt.Tx) error {
		subbucket := tx.Bucket([]byte("subs"))
		subbucket.ForEach(func(k, v []byte) error {
			var sub *gopub.Subscription
			json.Unmarshal(v, &sub)
			subs[sub.ID] = sub
			return nil
		})
		return nil
	})
	return subs
}
