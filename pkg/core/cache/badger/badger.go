package badger

import (
	"encoding/json"
	"log"

	"github.com/dgraph-io/badger"
	"github.com/golang/snappy"
)

type Badger struct {
	Client *badger.DB
}

func New(basePath string) *Badger {
	client, err := badger.Open(badger.DefaultOptions(basePath))
	if err != nil {
		log.Fatal(err)
	}
	return &Badger{Client: client}
}

// NewWithBadger returns a new Cache using the provided Diskv as underlying storage.
func NewWithBadger(client *badger.DB) *Badger {
	return &Badger{Client: client}
}

func (b *Badger) Get(key string) (string, error) {
	var resp []byte
	var err error
	err = b.Client.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) (err error) {
			resp, err = decompress(val)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return string(resp), err
}

func (b *Badger) Unmarshal(key string, object interface{}) error {
	item, err := b.Get(key)
	if err == nil {
		err = json.Unmarshal([]byte(item), object)
	}
	return err
}

func (b *Badger) Set(key string, value interface{}) error {
	err := b.Client.Update(func(txn *badger.Txn) error {
		cnt, err := compress(convertToBytes(value))
		if err != nil {
			return err
		}
		err = txn.Set([]byte(key), cnt)
		return err
	})
	return err
}

func (b *Badger) Fetch(key string, fc func() interface{}) (string, error) {
	var resp []byte
	var err error
	err = b.Client.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) (err error) {
			resp, err = decompress(val)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return string(resp), err
}

func (b *Badger) Delete(key string) error {
	return b.Client.View(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		if err != nil {
			return err
		}
		return nil
	})
}

func convertToBytes(value interface{}) []byte {
	switch result := value.(type) {
	case string:
		return []byte(result)
	case []byte:
		return result
	default:
		bytes, _ := json.Marshal(value)
		return bytes
	}
}

func compress(data []byte) ([]byte, error) {
	return snappy.Encode([]byte{}, data), nil
}

func decompress(data []byte) ([]byte, error) {
	return snappy.Decode([]byte{}, data)
}
