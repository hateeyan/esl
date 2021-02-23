package esl

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
)

func messageEqual(m1, m2 Message) bool {
	if string(m1.body) != string(m2.body) {
		return false
	}
	if m1.bs != m2.bs {
		return false
	}
loop:
	for _, kv1 := range m1.Header.args.kvs {
		for _, kv2 := range m2.Header.args.kvs {
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
			name: "parse log data",
			args: args{r: bufio.NewReader(bytes.NewReader([]byte("Content-Type: log/data\nContent-Length: 57\nLog-Level: 7\nUser-Data: \n\n2020-11-08 09:57:16.712466 [DEBUG] mod_commands.c:6391 2\n")))},
			want: Message{
				Header: Header{args: Args{kvs: []arg{
					{key: []byte("Content-Type"), value: []byte("log/data")},
					{key: []byte("Content-Length"), value: []byte("57")},
					{key: []byte("Log-Level"), value: []byte("7")},
					{key: []byte("User-Data"), value: []byte("")},
				}}},
				body: []byte("2020-11-08 09:57:16.712466 [DEBUG] mod_commands.c:6391 2\n"),
			},
		},
		{
			name: "parse reply message",
			args: args{r: bufio.NewReader(bytes.NewReader([]byte("Content-Type: command/reply\nReply-Text: +OK accepted\n\n")))},
			want: Message{
				Header: Header{args: Args{kvs: []arg{
					{key: []byte("Content-Type"), value: []byte("command/reply")},
					{key: []byte("Reply-Text"), value: []byte("+OK accepted")},
				}}},
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

func TestClient_waitMessage(t *testing.T) {
	type fields struct {
		r *bufio.Reader
	}
	type args struct {
		fn Handler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "wait for messages",
			fields: fields{r: bufio.NewReader(bytes.NewReader([]byte("Content-Length: 880\nContent-Type: text/event-plain\n\nEvent-Name: HEARTBEAT\nCore-UUID: 4e779f3e-8b37-4b39-9dec-0fc35be35b65\nFreeSWITCH-Hostname: node1\nFreeSWITCH-Switchname: node1\nFreeSWITCH-IPv4: 192.168.40.249\nFreeSWITCH-IPv6: %3A%3A1\nEvent-Date-Local: 2020-12-24%2018%3A49%3A10\nEvent-Date-GMT: Thu,%2024%20Dec%202020%2010%3A49%3A10%20GMT\nEvent-Date-Timestamp: 1608806950484747\nEvent-Calling-File: switch_core.c\nEvent-Calling-Function: send_heartbeat\nEvent-Calling-Line-Number: 81\nEvent-Sequence: 2137\nEvent-Info: System%20Ready\nUp-Time: 0%20years,%200%20days,%202%20hours,%2045%20minutes,%2059%20seconds,%20629%20milliseconds,%20798%20microseconds\nFreeSWITCH-Version: 1.10.5-release~64bit\nUptime-msec: 9959629\nSession-Count: 0\nMax-Sessions: 1000\nSession-Per-Sec: 30\nSession-Per-Sec-Last: 0\nSession-Per-Sec-Max: 4\nSession-Per-Sec-FiveMin: 0\nSession-Since-Startup: 16\nSession-Peak-Max: 8\nSession-Peak-FiveMin: 0\nIdle-CPU: 99.200000\n\nContent-Length: 879\nContent-Type: text/event-plain\n\nEvent-Name: HEARTBEAT\nCore-UUID: 4e779f3e-8b37-4b39-9dec-0fc35be35b65\nFreeSWITCH-Hostname: node1\nFreeSWITCH-Switchname: node1\nFreeSWITCH-IPv4: 192.168.40.249\nFreeSWITCH-IPv6: %3A%3A1\nEvent-Date-Local: 2020-12-24%2018%3A49%3A30\nEvent-Date-GMT: Thu,%2024%20Dec%202020%2010%3A49%3A30%20GMT\nEvent-Date-Timestamp: 1608806970484747\nEvent-Calling-File: switch_core.c\nEvent-Calling-Function: send_heartbeat\nEvent-Calling-Line-Number: 81\nEvent-Sequence: 2140\nEvent-Info: System%20Ready\nUp-Time: 0%20years,%200%20days,%202%20hours,%2046%20minutes,%2019%20seconds,%20634%20milliseconds,%2062%20microseconds\nFreeSWITCH-Version: 1.10.5-release~64bit\nUptime-msec: 9979634\nSession-Count: 0\nMax-Sessions: 1000\nSession-Per-Sec: 30\nSession-Per-Sec-Last: 0\nSession-Per-Sec-Max: 4\nSession-Per-Sec-FiveMin: 0\nSession-Since-Startup: 16\nSession-Peak-Max: 8\nSession-Peak-FiveMin: 0\nIdle-CPU: 99.200000\n\nContent-Length: 880\nContent-Type: text/event-plain\n\nEvent-Name: HEARTBEAT\nCore-UUID: 4e779f3e-8b37-4b39-9dec-0fc35be35b65\nFreeSWITCH-Hostname: node1\nFreeSWITCH-Switchname: node1\nFreeSWITCH-IPv4: 192.168.40.249\nFreeSWITCH-IPv6: %3A%3A1\nEvent-Date-Local: 2020-12-24%2018%3A49%3A50\nEvent-Date-GMT: Thu,%2024%20Dec%202020%2010%3A49%3A50%20GMT\nEvent-Date-Timestamp: 1608806990484730\nEvent-Calling-File: switch_core.c\nEvent-Calling-Function: send_heartbeat\nEvent-Calling-Line-Number: 81\nEvent-Sequence: 2142\nEvent-Info: System%20Ready\nUp-Time: 0%20years,%200%20days,%202%20hours,%2046%20minutes,%2039%20seconds,%20637%20milliseconds,%20730%20microseconds\nFreeSWITCH-Version: 1.10.5-release~64bit\nUptime-msec: 9999637\nSession-Count: 0\nMax-Sessions: 1000\nSession-Per-Sec: 30\nSession-Per-Sec-Last: 0\nSession-Per-Sec-Max: 4\nSession-Per-Sec-FiveMin: 0\nSession-Since-Startup: 16\nSession-Peak-Max: 8\nSession-Peak-FiveMin: 0\nIdle-CPU: 99.200000\n\nContent-Length: 880\nContent-Type: text/event-plain\n\nEvent-Name: HEARTBEAT\nCore-UUID: 4e779f3e-8b37-4b39-9dec-0fc35be35b65\nFreeSWITCH-Hostname: node1\nFreeSWITCH-Switchname: node1\nFreeSWITCH-IPv4: 192.168.40.249\nFreeSWITCH-IPv6: %3A%3A1\nEvent-Date-Local: 2020-12-24%2018%3A48%3A30\nEvent-Date-GMT: Thu,%2024%20Dec%202020%2010%3A48%3A30%20GMT\nEvent-Date-Timestamp: 1608806910464743\nEvent-Calling-File: switch_core.c\nEvent-Calling-Function: send_heartbeat\nEvent-Calling-Line-Number: 81\nEvent-Sequence: 2133\nEvent-Info: System%20Ready\nUp-Time: 0%20years,%200%20days,%202%20hours,%2045%20minutes,%2019%20seconds,%20622%20milliseconds,%20932%20microseconds\nFreeSWITCH-Version: 1.10.5-release~64bit\nUptime-msec: 9919622\nSession-Count: 0\nMax-Sessions: 1000\nSession-Per-Sec: 30\nSession-Per-Sec-Last: 0\nSession-Per-Sec-Max: 4\nSession-Per-Sec-FiveMin: 0\nSession-Since-Startup: 16\nSession-Peak-Max: 8\nSession-Peak-FiveMin: 0\nIdle-CPU: 99.300000\n\n")))},
			args: args{fn: func(msg *Message) {
				event := msg.Header.Get("Event-Name")
				fmt.Println("event:", event, len(msg.Bytes()))
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Inbound{
				Connection: &Connection{
					apps: &Applications{OnEvent: tt.args.fn},
					r:    tt.fields.r,
				},
			}
			c.waitMessage()
		})
	}
}
