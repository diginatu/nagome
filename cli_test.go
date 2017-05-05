package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestTCPAPI(t *testing.T) {
	var err error
	cli := NewCLI("test")
	cli.SavePath, err = ioutil.TempDir("", "nagome")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(cli.SavePath, pluginDirName), 0777); err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.RemoveAll(cli.SavePath)
		if err != nil {
			t.Fatal(err)
		}
	}()
	cv := NewCommentViewer("", cli)

	plug := newPlugin(cv)
	plug.Name = "main"
	plug.Description = "main plugin"
	plug.Version = "0.0"
	plug.Method = "tcp"
	plug.Subscribe = []string{DomainNagome, DomainComment, DomainUI}
	cv.AddPlugin(plug)

	cv.TCPPort = "0"

	cv.Start()

	conn, err := net.Dial("tcp", ":"+cv.TCPPort)
	if err != nil {
		t.Fatal(err)
	}

	// Connect as a main plugin
	fmt.Fprintf(conn, "{ \"domain\": \"nagome_direct\", \"command\": \"No\", \"content\": { \"no\": 0 } }\n")

	dec := json.NewDecoder(conn)
	m := new(Message)
	for {
		err := dec.Decode(m)
		if err != nil {
			t.Fatal("Should be accepted : ", err)
		}
		if m.Domain == DomainDirectngm && m.Command == CommDirectngmPlugEnabled {
			break
		}
	}

	err = conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	cv.Wait()
	// shold quit because main plugin was closed
}

func setLogForTest() {
	if testing.Verbose() {
		log.SetFlags(log.Lshortfile)
		log.SetPrefix("        ")
	} else {
		log.SetOutput(ioutil.Discard)
	}
}
func TestMain(m *testing.M) {
	flag.Parse()
	setLogForTest()
	code := m.Run()
	os.Exit(code)
}
