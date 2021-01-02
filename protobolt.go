package protobolt

import (
	"errors"

	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var ErrNotFound = errors.New("object not found")

type Object interface {
	proto.Message
	GetProtoBoltPK() []byte
}

type DB struct {
	BoltDB *bbolt.DB
}

func (db *DB) InitBuckets(os ...Object) error {
	for _, o := range os {
		if err := db.BoltDB.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(o.ProtoReflect().Descriptor().FullName()))
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) Get(o Object) error {
	return db.BoltDB.View(func(tx *bbolt.Tx) error {
		data := tx.Bucket([]byte(o.ProtoReflect().Descriptor().FullName())).Get(o.GetProtoBoltPK())
		if data == nil {
			return ErrNotFound
		}
		return proto.Unmarshal(data, o)
	})
}

func (db *DB) Save(o Object) error {
	data, err := proto.Marshal(o)
	if err != nil {
		return err
	}
	return db.BoltDB.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(o.ProtoReflect().Descriptor().FullName())).Put(
			o.GetProtoBoltPK(),
			data,
		)
	})
}

func (db *DB) Delete(o Object) error {
	return db.BoltDB.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(o.ProtoReflect().Descriptor().FullName())).Delete(
			o.GetProtoBoltPK(),
		)
	})
}

func (db *DB) ForEach(o Object, cb func(object Object) error) error {
	return db.BoltDB.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(o.ProtoReflect().Descriptor().FullName())).ForEach(func(k, v []byte) error {
			if err := proto.Unmarshal(v, o); err != nil {
				return err
			}
			return cb(o)
		})
	})
}
