package esl

import (
	"bufio"
	"bytes"
	"testing"
)

func messageEqual(m1, m2 Message) bool {
	if string(m1.body) != string(m2.body) {
		return false
	}
loop:
	for _, kv1 := range m1.Header.kvs {
		for _, kv2 := range m2.Header.kvs {
			if string(kv1.key) == string(kv2.key) && string(kv1.value) == string(kv2.value) {
				continue loop
			}
		}
		return false
	}
	return true
}

func Test_parseMessage(t *testing.T) {
	type args struct {
		r *bufio.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    Message
		wantErr bool
	}{
		{
			name: "parse incoming message",
			args: args{r: bufio.NewReader(bytes.NewReader([]byte("Content-Type: log/data\nContent-Length: 57\nLog-Level: 7\nUser-Data: \n\n2020-11-08 09:57:16.712466 [DEBUG] mod_commands.c:6391 2\n")))},
			want: Message{
				Header: Header{kvs: []arg{
					{key: []byte("Content-Type"), value: []byte("log/data")},
					{key: []byte("Content-Length"), value: []byte("57")},
					{key: []byte("Log-Level"), value: []byte("7")},
					{key: []byte("User-Data"), value: []byte("")},
				}},
				body: []byte("2020-11-08 09:57:16.712466 [DEBUG] mod_commands.c:6391 2\n"),
			},
		},
		{
			name: "parse reply message",
			args: args{r: bufio.NewReader(bytes.NewReader([]byte("Content-Type: command/reply\nReply-Text: +OK accepted\n\n")))},
			want: Message{
				Header: Header{kvs: []arg{
					{key: []byte("Content-Type"), value: []byte("command/reply")},
					{key: []byte("Reply-Text"), value: []byte("+OK accepted")},
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMessage(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !messageEqual(*got, tt.want) {
				t.Errorf("parseMessage() got = %v, want %v", got, tt.want)
			}
		})
	}
}
