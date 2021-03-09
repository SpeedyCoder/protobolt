package protobolt

import (
	"errors"

	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var ErrNotFound = errors.New("object not found")

type Entity interface {
	proto.Message
	GetProtoBoltPK() []byte
}

type DB struct {
	BoltDB *bbolt.DB
}

func (db DB) Init(entities ...Entity) error {
	for _, e := range entities {
		if err := db.BoltDB.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(e.ProtoReflect().Descriptor().FullName()))
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

func (db DB) Get(e Entity) error {
	return db.BoltDB.View(func(tx *bbolt.Tx) error {
		data := tx.Bucket([]byte(e.ProtoReflect().Descriptor().FullName())).Get(e.GetProtoBoltPK())
		if data == nil {
			return ErrNotFound
		}
		return proto.Unmarshal(data, e)
	})
}

func (db DB) Save(e Entity) error {
	data, err := proto.Marshal(e)
	if err != nil {
		return err
	}
	return db.BoltDB.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(e.ProtoReflect().Descriptor().FullName())).Put(
			e.GetProtoBoltPK(),
			data,
		)
	})
}

func (db DB) Delete(e Entity) error {
	return db.BoltDB.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(e.ProtoReflect().Descriptor().FullName())).Delete(
			e.GetProtoBoltPK(),
		)
	})
}

func (db DB) ForEach(e Entity, cb func(Entity) error) error {
	return db.BoltDB.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(e.ProtoReflect().Descriptor().FullName())).ForEach(func(k, v []byte) error {
			if err := proto.Unmarshal(v, e); err != nil {
				return err
			}
			return cb(e)
		})
	})
}
