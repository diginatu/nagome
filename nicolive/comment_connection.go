package nicolive

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const ()

// CommentConnection is a struct to manage sending/receiving comments.
// This struct automatically submits NULL character to reserve connection and
// get the PostKey, which is necessary for sending comments.
// liveWaku should have connection information which can get by fetchInformation()
type CommentConnection struct {
	liveWaku *LiveWaku
	socket   net.Conn

	reconnectTimes    int
	reconnectN        int
	reconnectWaitMsec int

	commReadWriter bufio.ReadWriter
	writeMutex     sync.Mutex
	waitGroup      sync.WaitGroup
	termSig        chan struct{}
}

// NewCommentConnection returns a pointer to new CommentConnection
func NewCommentConnection(l *LiveWaku) *CommentConnection {
	return &CommentConnection{
		liveWaku:          l,
		reconnectTimes:    3,
		reconnectWaitMsec: 1000,
		reconnectN:        0,
	}
}

// Connect Connect to nicolive and start receiving comment
func (cc CommentConnection) Connect() NicoError {
	var err error

	addrport := fmt.Sprintf("%s:%s",
		cc.liveWaku.CommentServer.Addr,
		cc.liveWaku.CommentServer.Port)

	cc.socket, err = net.Dial("tcp", addrport)
	if err != nil {
		return NicoErrFromStdErr(err)
	}

	cc.commReadWriter.Reader = bufio.NewReader(cc.socket)
	cc.commReadWriter.Writer = bufio.NewWriter(cc.socket)

	fmt.Fprintf(cc.socket,
		"<thread thread=\"%s\" res_from=\"-1000\" version=\"20061206\" />\x00",
		cc.liveWaku.CommentServer.Thread)

	cc.waitGroup.Add(2)
	go cc.receiveStream()
	go cc.keepAlive()

	return nil
}

func (cc *CommentConnection) receiveStream() {
	defer cc.waitGroup.Done()

	for {
		commxml, err := cc.commReadWriter.ReadString('\x00')
		if err != nil {
			fmt.Println("receive err")
			return
		}
		fmt.Println(commxml)

		if strings.HasPrefix(commxml, "<thread ") {
			fmt.Println("thread")
			continue
		}
		if strings.HasPrefix(commxml, "<chat ") {
			fmt.Println("chat")
			continue
		}
	}
}

func (cc *CommentConnection) keepAlive() {
	defer cc.waitGroup.Done()

	tick := time.Tick(time.Minute)
	for {
		select {
		case <-tick:
			cc.writeMutex.Lock()
			err := cc.commReadWriter.WriteByte(0)
			if err == nil {
				err = cc.commReadWriter.Flush()
			}
			cc.writeMutex.Unlock()
			if err != nil {
				return
			}
		case <-cc.termSig:
			return
		}
	}
}

// Close closes connection
func (cc CommentConnection) Close() {
	cc.socket.Close()
	cc.waitGroup.Wait()
	close(cc.termSig)
}
