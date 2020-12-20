package esl

import (
	"testing"
)

func TestMessage_payload(t *testing.T) {
	type fields struct {
		Header Header
		bs     int
		body   []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   Message
	}{
		{
			name: "parse event payload",
			fields: fields{
				Header: Header{contentLength: 516},
				body:   []byte("Event-Name: API\nCore-UUID: aa3be358-32b9-11eb-b2df-77cae58380cf\nFreeSWITCH-Hostname: localhost.localdomain\nFreeSWITCH-Switchname: localhost.localdomain\nFreeSWITCH-IPv4: 192.168.40.192\nFreeSWITCH-IPv6: %3A%3A1\nEvent-Date-Local: 2020-12-20%2012%3A58%3A17\nEvent-Date-GMT: Sun,%2020%20Dec%202020%2004%3A58%3A17%20GMT\nEvent-Date-Timestamp: 1608440297753799\nEvent-Calling-File: switch_loadable_module.c\nEvent-Calling-Function: switch_api_execute\nEvent-Calling-Line-Number: 2424\nEvent-Sequence: 624574\nAPI-Command: status\n\n"),
			},
			want: Message{Header: Header{contentLength: 516, kvs: []arg{
				{key: []byte("Event-Name"), value: []byte("API")},
				{key: []byte("Core-UUID"), value: []byte("aa3be358-32b9-11eb-b2df-77cae58380cf")},
				{key: []byte("FreeSWITCH-Hostname"), value: []byte("localhost.localdomain")},
				{key: []byte("FreeSWITCH-Switchname"), value: []byte("localhost.localdomain")},
				{key: []byte("FreeSWITCH-IPv4"), value: []byte("192.168.40.192")},
				{key: []byte("FreeSWITCH-IPv6"), value: []byte("%3A%3A1")},
				{key: []byte("Event-Date-Local"), value: []byte("2020-12-20%2012%3A58%3A17")},
				{key: []byte("Event-Date-GMT"), value: []byte("Sun,%2020%20Dec%202020%2004%3A58%3A17%20GMT")},
				{key: []byte("Event-Date-Timestamp"), value: []byte("1608440297753799")},
				{key: []byte("Event-Calling-File"), value: []byte("switch_loadable_module.c")},
				{key: []byte("Event-Calling-Function"), value: []byte("switch_api_execute")},
				{key: []byte("Event-Calling-Line-Number"), value: []byte("2424")},
				{key: []byte("Event-Sequence"), value: []byte("624574")},
				{key: []byte("API-Command"), value: []byte("status")},
			}}, bs: 516, body: []byte("Event-Name: API\nCore-UUID: aa3be358-32b9-11eb-b2df-77cae58380cf\nFreeSWITCH-Hostname: localhost.localdomain\nFreeSWITCH-Switchname: localhost.localdomain\nFreeSWITCH-IPv4: 192.168.40.192\nFreeSWITCH-IPv6: %3A%3A1\nEvent-Date-Local: 2020-12-20%2012%3A58%3A17\nEvent-Date-GMT: Sun,%2020%20Dec%202020%2004%3A58%3A17%20GMT\nEvent-Date-Timestamp: 1608440297753799\nEvent-Calling-File: switch_loadable_module.c\nEvent-Calling-Function: switch_api_execute\nEvent-Calling-Line-Number: 2424\nEvent-Sequence: 624574\nAPI-Command: status\n\n")},
		},
		{
			name: "parse event with body",
			fields: fields{
				Header: Header{contentLength: 93},
				body:   []byte("Event-Name: CUSTOM\nCall-ID: 40ac06d4-429f-11eb-af7b-77cae58380cf\nContent-Length: 9\n\nbody test"),
			},
			want: Message{Header: Header{contentLength: 93, kvs: []arg{
				{key: []byte("Event-Name"), value: []byte("CUSTOM")},
				{key: []byte("Call-ID"), value: []byte("40ac06d4-429f-11eb-af7b-77cae58380cf")},
				{key: []byte("Content-Length"), value: []byte("9")},
			}},
				bs:   84,
				body: []byte("Event-Name: CUSTOM\nCall-ID: 40ac06d4-429f-11eb-af7b-77cae58380cf\nContent-Length: 9\n\nbody test"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Message{
				Header: tt.fields.Header,
				bs:     tt.fields.bs,
				body:   tt.fields.body,
			}
			if got := e.payload(); !messageEqual(*got, tt.want) {
				t.Errorf("payload() = %v, want %v", got, tt.want)
			}
		})
	}
}
