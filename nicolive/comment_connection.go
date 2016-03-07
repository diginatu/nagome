package nicolive

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// CommentConnection is a struct to manage sending/receiving comments.
// This struct automatically submits NULL character to reserve connection and
// get the PostKey, which is necessary for sending comments.
// liveWaku should have connection information which is able to get by fetchInformation()
type CommentConnection struct {
	liveWaku *LiveWaku
	socket   net.Conn

	reconnectTimes    int
	reconnectWaitTime time.Duration
	reconnectN        int

	commReadWriter bufio.ReadWriter
	connectMutex   sync.RWMutex
	retryMutex     sync.Mutex
	waitGroup      sync.WaitGroup
	termSig        chan struct{}
}

// NewCommentConnection returns a pointer to new CommentConnection
func NewCommentConnection(l *LiveWaku) *CommentConnection {
	return &CommentConnection{
		liveWaku:          l,
		reconnectTimes:    3,
		reconnectWaitTime: time.Second,
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
		return cc.RetryConnect()
		//return NicoErrFromStdErr(err)
	}

	cc.commReadWriter = bufio.ReadWriter{
		Reader: bufio.NewReader(cc.socket),
		Writer: bufio.NewWriter(cc.socket),
	}

	cc.connectMutex.Lock()
	fmt.Fprintf(cc.commReadWriter,
		"<thread thread=\"%s\" res_from=\"-1000\" version=\"20061206\" />\x00",
		cc.liveWaku.CommentServer.Thread)
	err = cc.commReadWriter.Flush()
	cc.connectMutex.Unlock()
	if err != nil {
		return cc.RetryConnect()
		//return NicoErrFromStdErr(err)
	}

	cc.waitGroup.Add(2)
	go cc.receiveStream()
	go cc.keepAlive()
	cc.reconnectN = 0

	return nil
}

// RetryConnect try to retry connecting comment server
func (cc CommentConnection) RetryConnect() NicoError {
	cc.retryMutex.Lock()
	defer cc.retryMutex.Unlock()

	cc.Close()
	cc.reconnectN++
	if cc.reconnectN < cc.reconnectTimes {
		select {
		case <-cc.termSig:
			return NicoErr(NicoErrOther, "comment connection terminated",
				"connection was closed and canceled to reconnect")
		case <-time.After(cc.reconnectWaitTime):
			cc.Connect()
		}
	} else {
		return NicoErr(NicoErrOther, "comment connection error",
			"retry time reached reconnectTimes")
	}
	return nil
}

func (cc *CommentConnection) receiveStream() {
	defer cc.waitGroup.Done()

	for {
		cc.connectMutex.RLock()
		commxml, err := cc.commReadWriter.ReadString('\x00')
		cc.connectMutex.RUnlock()
		if err != nil {
			go cc.RetryConnect()
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
			cc.connectMutex.Lock()
			err := cc.commReadWriter.WriteByte(0)
			if err == nil {
				err = cc.commReadWriter.Flush()
			}
			cc.connectMutex.Unlock()
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
	close(cc.termSig)
	cc.waitGroup.Wait()
}
