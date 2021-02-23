package esl

import (
	"reflect"
	"testing"
)

func TestArgs_Set(t *testing.T) {
	type fields struct {
		kvs []arg
		buf []byte
	}
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Args
	}{
		{
			name: "key exist",
			fields: fields{kvs: []arg{
				{key: []byte("key1"), value: []byte("value1")},
				{key: []byte("key2"), value: []byte("value2")},
			}},
			args: args{key: "key1", value: "want1"},
			want: Args{kvs: []arg{
				{key: []byte("key1"), value: []byte("want1")},
				{key: []byte("key2"), value: []byte("value2")},
			}},
		},
		{
			name: "key not exist",
			fields: fields{kvs: []arg{
				{key: []byte("key1"), value: []byte("value1")},
				{key: []byte("key2"), value: []byte("value2")},
			}},
			args: args{key: "key3", value: "want1"},
			want: Args{kvs: []arg{
				{key: []byte("key1"), value: []byte("value1")},
				{key: []byte("key2"), value: []byte("value2")},
				{key: []byte("key3"), value: []byte("want1")},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Args{
				kvs: tt.fields.kvs,
				buf: tt.fields.buf,
			}
			a.Set(tt.args.key, tt.args.value)
			if !reflect.DeepEqual(a, tt.want) {
				t.Errorf("Set() = %v, want %v", a, tt.want)
			}
		})
	}
}

func TestArgs_HeaderBytes(t *testing.T) {
	type fields struct {
		kvs []arg
		buf []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "message header",
			fields: fields{kvs: []arg{
				{key: []byte("call-command"), value: []byte("execute")},
				{key: []byte("execute-app-name"), value: []byte("playback")},
				{key: []byte("execute-app-arg"), value: []byte("foo.wav")},
			}},
			want: []byte("call-command: execute\nexecute-app-name: playback\nexecute-app-arg: foo.wav\n\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Args{
				kvs: tt.fields.kvs,
				buf: tt.fields.buf,
			}
			if got := a.HeaderBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HeaderBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
