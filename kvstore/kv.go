package kvstore

import "fmt"

type KVStore[K comparable,V string|int] struct {
	Store map[K]V
}

func InitKVStore[K comparable,V string|int]() *KVStore[K,V]{
	return &KVStore[K,V]{
		Store: make(map[K]V),
	}
}

func (kv *KVStore[K,V]) Get(key K) V {
	if val ,ok := kv.Store[key]; !ok {
		var z V
		return z// Return the zero value of type V
	}else{
		return val
	}
}

func (kv *KVStore[K,V]) Set(key K, value V) {
	kv.Store[key] = value
}

func (kv *KVStore[K,V]) Delete(key K) {
	fmt.Println("Deleting key")
	fmt.Println(key)
	delete(kv.Store,key)
}
