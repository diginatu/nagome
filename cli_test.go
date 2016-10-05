package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/diginatu/nagome/nicolive"
)

func TestMainTCP(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.Ltime)
	var cv = CommentViewer{
		Ac:      new(nicolive.Account),
		TCPPort: "",
		Evch:    make(chan *Message, eventBufferSize),
		Quit:    make(chan struct{}),
	}
	cv.Cmm = nicolive.NewCommentConnection(&cv)

	plug := &plugin{
		Name:        "main",
		Description: "main plugin",
		Version:     "0.0",
		Method:      "tcp",
		Depends:     []string{DomainNagome, DomainComment, DomainUI},
	}
	plug.Init(1)
	cv.Pgns = append(cv.Pgns, plug)
	cv.TCPPort = "0"

	cv.Start()

	conn, err := net.Dial("tcp", ":"+cv.TCPPort)
	if err != nil {
		t.Fatal(err)
	}

	// Connect as a main plugin
	fmt.Fprintf(conn, "{ \"domain\": \"nagome_direct\", \"command\": \"No\", \"content\": { \"no\": 1 } }\n")

	dec := json.NewDecoder(conn)
	m := new(Message)
	for {
		err := dec.Decode(m)
		if err != nil {
			t.Fatal(err)
		}
		if m.Domain == DomainDirectngm && m.Command == CommDirectngmPlugEnabled {
			break
		}
	}

	conn.Close()
	cv.Wait()
	// shold quit because main plugin was closed
}
