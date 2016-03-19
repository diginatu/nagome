package nicolive

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	numCommentConnectionGoRoutines = 2
)

// CommentConnection is a struct to manage sending/receiving comments.
// This struct automatically submits NULL character to reserve connection and
// get the PostKey, which is necessary for sending comments.
// liveWaku should have connection information which is able to get by fetchInformation()
type CommentConnection struct {
	isConnected bool
	liveWaku    *LiveWaku
	socket      net.Conn

	ReconnectTimes    uint
	ReconnectWaitTime time.Duration
	reconnectN        uint

	commReadWriter bufio.ReadWriter

	connectReadMutex  sync.Mutex
	connectWriteMutex sync.Mutex
	retryMutex        sync.Mutex
	termSig           chan bool
}

// NewCommentConnection returns a pointer to new CommentConnection
func NewCommentConnection(l *LiveWaku) *CommentConnection {
	return &CommentConnection{
		liveWaku:          l,
		ReconnectTimes:    3,
		ReconnectWaitTime: time.Second,
		termSig:           make(chan bool),
	}
}

func (cc *CommentConnection) open() {
	var err error
	Logger.Println("CommentConnection opening")

	addrport := fmt.Sprintf("%s:%s",
		cc.liveWaku.CommentServer.Addr,
		cc.liveWaku.CommentServer.Port)

	cc.socket, err = net.Dial("tcp", addrport)
	if err != nil {
		Logger.Println(NicoErrFromStdErr(err))
		return
	}

	cc.commReadWriter = bufio.ReadWriter{
		Reader: bufio.NewReader(cc.socket),
		Writer: bufio.NewWriter(cc.socket),
	}

	fmt.Fprintf(cc.commReadWriter,
		"<thread thread=\"%s\" res_from=\"-1000\" version=\"20061206\" />\x00",
		cc.liveWaku.CommentServer.Thread)
	err = cc.commReadWriter.Flush()
	if err != nil {
		Logger.Println(NicoErrFromStdErr(err))
		return
	}

	cc.reconnectN = 0
}

// Connect Connect to nicolive and start receiving comment
func (cc *CommentConnection) Connect() NicoError {
	if cc.isConnected {
		return NicoErr(NicoErrOther, "already connected", "")
	}
	cc.isConnected = true

	cc.retryMutex.Lock()
	cc.connectWriteMutex.Lock()
	cc.connectReadMutex.Lock()

	go func() {
		cc.open()

		cc.retryMutex.Unlock()
		cc.connectWriteMutex.Unlock()
		cc.connectReadMutex.Unlock()
	}()

	go cc.receiveStream()
	go cc.keepAlive()

	return nil
}

// retryConnect try to retry connecting comment server
// returns ok or not
func (cc *CommentConnection) retryConnect() bool {
	cc.retryMutex.Lock()
	defer cc.retryMutex.Unlock()
	cc.connectWriteMutex.Lock()
	defer cc.connectWriteMutex.Unlock()
	cc.connectReadMutex.Lock()
	defer cc.connectReadMutex.Unlock()

	Logger.Println("CommentConnection reconnect")

	cc.close()
	cc.reconnectN++
	if cc.reconnectN <= cc.ReconnectTimes {
		select {
		case <-cc.termSig:
			Logger.Println(NicoErr(NicoErrOther, "comment connection terminated",
				"connection was closed and canceled to reconnect"))
			return false
		case <-time.After(cc.ReconnectWaitTime):
			cc.open()
			return true
		}
	} else {
		Logger.Println(NicoErr(NicoErrOther, "comment connection error",
			"retry time reached reconnectTimes"))
		go cc.Disconnect()
		<-cc.termSig
		return false
	}
}

func (cc *CommentConnection) receiveStream() {
	for {
		select {
		case <-cc.termSig:
			return
		default:
			cc.connectReadMutex.Lock()
			commxml, err := cc.commReadWriter.ReadString('\x00')
			cc.connectReadMutex.Unlock()
			if err != nil {
				Logger.Println(NicoErrFromStdErr(err))
				if ok := cc.retryConnect(); ok {
					continue
				}
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
}

func (cc *CommentConnection) keepAlive() {
	tick := time.Tick(time.Minute)
	for {
		select {
		case <-cc.termSig:
			return
		case <-tick:
			cc.connectWriteMutex.Lock()
			err := cc.commReadWriter.WriteByte(0)
			if err == nil {
				err = cc.commReadWriter.Flush()
			}
			cc.connectWriteMutex.Unlock()
			if err != nil {
				Logger.Println(NicoErrFromStdErr(err))
				continue
			}
		}
	}
}

// Close closes the connection
// do not end go routines
func (cc *CommentConnection) close() {
	cc.socket.Close()
}

// Disconnect close and disconnect
// terminate all goroutines and wait to exit
func (cc *CommentConnection) Disconnect() NicoError {
	if !cc.isConnected {
		return NicoErr(NicoErrOther, "not connected yet", "")
	}

	cc.close()
	for i := 0; i < numCommentConnectionGoRoutines; i++ {
		cc.termSig <- true
	}

	cc.isConnected = false
	Logger.Println("CommentConnection disconnected")

	return nil
}
