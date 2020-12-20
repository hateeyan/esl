package main

import (
	"fmt"
	"github.com/hateeyan/esl"
	"time"
)

func main() {
	client, err := esl.Dial("192.168.40.192:8014", "NewVois001", func(msg *esl.Message) {
		fmt.Println(msg.ContentType())
		fmt.Println(string(msg.Body()))
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()
	client.OnReconnect = func(c *esl.Client) {
		c.Command("event all")
	}

	reply := client.Command("event all")
	if err := reply.Err(); err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Minute)
}
