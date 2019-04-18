package idobot

/*
DB周りの操作を集めたファイル
*/

import (
	bolt "go.etcd.io/bbolt"
)

const (
	systemBucket = "System"
)

func (bot *botImpl) DB() *bolt.DB {
	return bot.db
}

func (bot *botImpl) Save(key, value string) error {
	return bot.SaveBucket(key, value, bot.bucketName)
}

func (bot *botImpl) SaveSystemBucket(key, value string) error {
	return bot.SaveBucket(key, value, systemBucket)
}

func (bot *botImpl) SaveBucket(key, value, bucket string) error {
	return bot.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(key), []byte(value))
		return err
	})
}

func (bot *botImpl) Read(key string) (string, error) {
	return bot.ReadBucket(key, bot.bucketName)
}

func (bot *botImpl) ReadSystemBucket(key string) (string, error) {
	return bot.ReadBucket(key, systemBucket)
}

func (bot *botImpl) ReadBucket(key, bucket string) (string, error) {
	var v string
	err := bot.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		v = string(b.Get([]byte(key)))
		return nil
	})
	return v, err
}
