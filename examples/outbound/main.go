package main

import (
	"fmt"
	"github.com/hateeyan/esl"
)

func main() {
	outbound := esl.Outbound{
		Handler: func(conn *esl.Connection) {
			fmt.Println("new connection")
			fmt.Println(string(conn.Info().Bytes()))
			reply := conn.Execute("answer", "")
			if err := reply.Err(); err != nil {
				panic(err)
			}
			reply = conn.Execute("info", "")
			if err := reply.Err(); err != nil {
				panic(err)
			}
			reply = conn.Hangup("NORMAL_CLEARING")
			if err := reply.Err(); err != nil {
				panic(err)
			}
		},
	}
	if err := outbound.Serve(); err != nil {
		panic(err)
	}
}
