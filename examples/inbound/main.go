package main

import (
	"fmt"
	"github.com/hateeyan/esl"
	"time"
)

func main() {
	inbound := esl.Inbound{
		Address:  "192.168.40.249:8021",
		Password: "ClueCon",
		Apps: esl.Applications{
			OnReconnect: func(c *esl.Inbound) {
				reply := c.Event("CHANNEL_HANGUP_COMPLETE HEARTBEAT")
				if err := reply.Err(); err != nil {
					panic(err)
				}
			},
			OnEvent: func(msg *esl.Message) {
				//fmt.Println(msg.Header.Get("Event-Sequence"))
				event := msg.Header.Get("Event-Name")
				fmt.Println("receive new event:", event, len(msg.Bytes()))
				switch event {
				case "HEARTBEAT":
					fmt.Println(string(msg.Bytes()))
				case "CHANNEL_HANGUP_COMPLETE":
					fmt.Println("hangup")
				default:
					fmt.Println("unexpected event type:", event)
					fmt.Println(string(msg.Bytes()))
				}
			},
		},
	}
	if err := inbound.Run(); err != nil {
		panic(err)
	}
	defer inbound.Close()

	reply := inbound.Event("CHANNEL_HANGUP_COMPLETE HEARTBEAT")
	if err := reply.Err(); err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Minute)
}
