package idobot

/*
Store周りの操作を集めたファイル

必要なバケット

- content
- tag
- time
*/

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type Store interface {
	Save(name, content string) error
	Read(name string) (string, error)
	Close() error
}

func NewStore(path string) (Store, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}

	// バケットを作る
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("content"))
		return err
	})

	if err != nil {
		db.Close()
		return nil, err
	}

	return &StoreImpl{
		db: db,
	}, nil
}

type StoreImpl struct {
	db *bolt.DB
}

func (st *StoreImpl) Save(name, content string) error {
	return st.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("content"))
		if bkt == nil {
			return fmt.Errorf("content bucket not found")
		}
		return bkt.Put([]byte(name), []byte(content))
	})
}
func (st *StoreImpl) Read(name string) (string, error) {
	var content string
	err := st.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("content"))
		if bkt == nil {
			return fmt.Errorf("content bucket not found")
		}
		b := bkt.Get([]byte(name))
		if b == nil {
			return fmt.Errorf("%s didn't exist in %s bucket.", name, "content")
		}
		content = string(b)
		return nil
	})

	return content, err
}

func (st *StoreImpl) Close() error {
	return st.db.Close()
}
