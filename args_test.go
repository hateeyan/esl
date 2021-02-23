package esl

import (
	"reflect"
	"testing"
)

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
