# esl

Freeswitch Event Socket Protocol implementation for Go.

Still in progress.

### Usage

- handle event and log

  ```go
  import (
  	"fmt"
  	"github.com/hateeyan/esl"
  )
  
  func main() {
  	client, err := esl.Dial("192.168.40.192:8014", "NewVois001")
  	if err != nil {
  		panic(err)
  	}
  	defer client.Close()
  
  	if err := client.Send("log 7"); err != nil {
  		panic(err)
  	}
  	if err := client.Send("event all"); err != nil {
  		panic(err)
  	}
  
  	client.Run(func(msg *esl.Message) {
  		fmt.Println(msg.ContentType())
  		fmt.Println(string(msg.Body()))
  	})
  }
  ```


