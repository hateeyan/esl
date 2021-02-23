package esl

import (
	"reflect"
	"testing"
)

func TestCommand_Message(t *testing.T) {
	type fields struct {
		uuid []byte
		kvs  Args
	}
	type args struct {
		outbound bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			name: "outbound",
			fields: fields{
				kvs: Args{kvs: []arg{
					{key: []byte("call-command"), value: []byte("execute")},
					{key: []byte("execute-app-name"), value: []byte("playback")},
					{key: []byte("execute-app-arg"), value: []byte("foo.wav")},
				}},
			},
			args: args{outbound: true},
			want: []byte("sendmsg\ncall-command: execute\nexecute-app-name: playback\nexecute-app-arg: foo.wav\n\n"),
		},
		{
			name: "inbound",
			fields: fields{
				uuid: []byte("46ca9b34-2bd2-464f-ad0c-082914d264a8"),
				kvs: Args{kvs: []arg{
					{key: []byte("call-command"), value: []byte("execute")},
					{key: []byte("execute-app-name"), value: []byte("playback")},
					{key: []byte("execute-app-arg"), value: []byte("foo.wav")},
				}},
			},
			args: args{outbound: false},
			want: []byte("sendmsg 46ca9b34-2bd2-464f-ad0c-082914d264a8\ncall-command: execute\nexecute-app-name: playback\nexecute-app-arg: foo.wav\n\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Command{
				uuid: tt.fields.uuid,
				kvs:  tt.fields.kvs,
			}
			if got := c.message(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Message() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommand_Api(t *testing.T) {
	type fields struct {
		kvs Args
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "api",
			fields: fields{kvs: Args{kvs: []arg{
				{key: []byte("execute-app-name"), value: []byte("originate")},
				{key: []byte("execute-app-arg"), value: []byte("sofia/mydomain.com/ext@yourvsp.com 1000")},
			}}},
			want: []byte("api originate sofia/mydomain.com/ext@yourvsp.com 1000\n\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Command{kvs: tt.fields.kvs}
			if got := c.api(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Api() = %v, want %v", got, tt.want)
			}
		})
	}
}
