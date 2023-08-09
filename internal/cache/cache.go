package cache

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
)

type Cache interface {
	Get(namespace, key string) (string, error)
	Set(namespace, key string, value string) error
	Has(namespace, key string) (bool, error)
	Delete(namespace, key string) error
	Close() error
}

// Db is a wrapper around badger.DB
type Db struct {
	db *badger.DB
}

// New returns a new initialized BadgerDB database implementing the DB
// interface. If the database cannot be initialized, an error will be returned.
func New(path string) (Cache, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}
	c := &Db{
		db: db,
	}
	return c, nil
}

// namespaceKey returns a composite key used for lookup and storage for a given namespace and key.
func namespaceKey(namespace, key string) []byte {
	prefix := []byte(fmt.Sprintf("%s/", namespace))
	return append(prefix, key...)
}

// Get implements the DB interface. It attempts to get a value for a given key
// and namespace. If the key does not exist in the provided namespace, an error
// is returned, otherwise the retrieved value.
func (c *Db) Get(namespace, key string) (string, error) {
	var value []byte
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(namespaceKey(namespace, key))
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(value)
		return err
	})
	return string(value), err
}

// Set implements the DB interface. It attempts to store a value for a given key
// and namespace. If the key/value pair cannot be saved, an error is returned.
func (c *Db) Set(namespace, key, value string) error {
	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Set(namespaceKey(namespace, key), []byte(value))
	})
}

// Has implements the DB interface. It returns a boolean reflecting if the
// database has a given key for a namespace or not. An error is only returned if
// an error to Get would be returned that is not of type badger.ErrKeyNotFound.
func (c *Db) Has(namespace, key string) (bool, error) {
	var has bool
	err := c.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(namespaceKey(namespace, key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				has = false
				return nil
			}
			return err
		}
		has = true
		return nil
	})
	return has, err
}

// Delete implements the DB interface. It attempts to delete a value for a given key
// and namespace. If the key/value pair cannot be deleted, an error is returned.
func (c *Db) Delete(namespace, key string) error {
	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(namespaceKey(namespace, key))
	})
}

// Close implements the DB interface. It closes the connection to the underlying
// BadgerDB database as well as invoking the context's cancel function.
func (c *Db) Close() error {
	return c.db.Close()
}
