package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/diginatu/nagome/nicolive"
)

func setLogForTest() {
	if testing.Verbose() {
		log.SetFlags(log.Lshortfile | log.Ltime)
	} else {
		log.SetOutput(ioutil.Discard)
	}
}

func TestMainTCP(t *testing.T) {
	setLogForTest()
	App.SavePath = filepath.Join(os.TempDir(), "nagome_test")
	if err := os.MkdirAll(filepath.Join(App.SavePath, pluginDirName), 0777); err != nil {
		t.Fatal(err)
	}
	cv := NewCommentViewer(new(nicolive.Account), "")

	plug := newPlugin(cv)
	plug.Name = "main"
	plug.Description = "main plugin"
	plug.Version = "0.0"
	plug.Method = "tcp"
	plug.Depends = []string{DomainNagome, DomainComment, DomainUI}
	cv.AddPlugin(plug)

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
