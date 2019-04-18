package idobot

/*
DB周りの操作を集めたファイル
*/

import (
	bolt "go.etcd.io/bbolt"
)

func (bot *botImpl) DB() *bolt.DB {
	return bot.db
}

func (bot *botImpl) PutDB(key, value string) error {
	return bot.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("User"))
		err := b.Put([]byte(key), []byte(value))
		return err
	})
}

func (bot *botImpl) GetDB(key string) (string, error) {
	var v string
	err := bot.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("User"))
		v = string(b.Get([]byte(key)))
		return nil
	})
	return v, err
}
