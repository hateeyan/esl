package esl

import (
	"bytes"
	"strconv"
)

type arg struct {
	key   []byte
	value []byte
}

type args []arg

func (a *args) Add(key, value []byte) {
	kvs := *a
	if cap(kvs) > len(kvs) {
		kvs = kvs[:len(kvs)+1]
	} else {
		kvs = append(kvs, arg{})
	}
	kv := &kvs[len(kvs)-1]
	kv.key = append(kv.key[:0], key...)
	kv.value = append(kv.value[:0], value...)
	*a = kvs
}

func (a args) get(key []byte) []byte {
	for _, kv := range a {
		if bytes.Compare(kv.key, key) == 0 {
			return kv.value
		}
	}
	return nil
}

func (a args) GetInt(key []byte) (int, error) {
	v := a.get(key)
	if v == nil {
		return 0, nil
	}
	return strconv.Atoi(string(v))
}

type Header struct {
	kvs args
}

func (h *Header) Add(key, value []byte) {
	h.kvs.Add(key, value)
}

func (h *Header) GetInt(key string) (int, error) {
	return h.kvs.GetInt([]byte(key))
}

func (h *Header) Get(key string) string {
	return string(h.kvs.get([]byte(key)))
}

func (h *Header) reset() {
	h.kvs = h.kvs[:0]
}
