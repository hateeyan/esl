package esl

import (
	"bytes"
	"strconv"
)

var strColonSpace = []byte(": ")

type arg struct {
	key   []byte
	value []byte
}

type Args struct {
	kvs []arg
	buf []byte
}

func (a *Args) Add(key, value string) {
	var kv *arg
	a.kvs, kv = allocArg(a.kvs)
	kv.key = append(kv.key[:0], key...)
	kv.value = append(kv.value[:0], value...)
}

func (a *Args) AddBytes(key, value []byte) {
	var kv *arg
	a.kvs, kv = allocArg(a.kvs)
	kv.key = append(kv.key[:0], key...)
	kv.value = append(kv.value[:0], value...)
}

func allocArg(h []arg) ([]arg, *arg) {
	n := len(h)
	if cap(h) > n {
		h = h[:n+1]
	} else {
		h = append(h, arg{})
	}
	return h, &h[n]
}

func (a *Args) GetBytes(key []byte) []byte {
	for _, kv := range a.kvs {
		if bytes.Compare(kv.key, key) == 0 {
			return kv.value
		}
	}
	return nil
}

func (a *Args) GetInt(key []byte) (int, error) {
	v := a.GetBytes(key)
	if v == nil {
		return 0, nil
	}
	return strconv.Atoi(string(v))
}

// HeaderBytes return header string
func (a *Args) HeaderBytes() []byte {
	a.buf = a.AppendBytes(a.buf[:0])
	return a.buf
}

// AppendBytes appends header string to dst
func (a *Args) AppendBytes(dst []byte) []byte {
	for i, n := 0, len(a.kvs); i < n; i++ {
		kv := &a.kvs[i]
		dst = append(dst, kv.key...)
		dst = append(dst, strColonSpace...)
		dst = append(dst, kv.value...)
		dst = append(dst, '\n')
	}
	dst = append(dst, '\n')
	return dst
}

func (a *Args) reset() {
	a.kvs = a.kvs[:0]
}
