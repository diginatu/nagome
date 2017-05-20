package viewer

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
)

const (
	DefaultAppName = "nagome"
)

func makeTestCLI(savePath string) *CLI {
	c := &CLI{
		AppName:   DefaultAppName,
		SavePath:  savePath,
		OutStream: ioutil.Discard,
	}

	logFlags := log.Lshortfile
	if testing.Verbose() {
		c.ErrStream = os.Stderr
	} else {
		c.ErrStream = ioutil.Discard
	}
	c.log = log.New(c.ErrStream, "        ", logFlags)
	return c
}

func TestCLIVersion(t *testing.T) {
	savepath, err := ioutil.TempDir("", DefaultAppName)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.RemoveAll(savepath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	cli := makeTestCLI(savepath)

	rt := cli.RunCli([]string{DefaultAppName, "-v", "-dbgtostd"})
	if rt != 0 {
		t.Fatalf("Return value should be %v but %v", 0, rt)
	}
}

func TestCLIQuit(t *testing.T) {
	savepath, err := ioutil.TempDir("", DefaultAppName)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.RemoveAll(savepath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	cli := makeTestCLI(savepath)
	//cli.InStream = strings.NewReader("")
	r, _ := io.Pipe()
	cli.InStream = r
	//r.Close()

	rt := cli.RunCli([]string{DefaultAppName, "-savepath", savepath, "-dbgtostd"})
	if rt != 0 {
		t.Fatalf("Return value should be %v but %v", 0, rt)
	}
}

func TestTCPAPI(t *testing.T) {
	var err error
	cli := NewCLI("test", "nagome")
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
