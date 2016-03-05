package nicolive

import (
	"bufio"
	"fmt"
	"net"
)

const ()

// CommentConnection is a struct to manage sending/receiving comments.
// This struct automatically submits NULL character to reserve connection and
// get the PostKey, which is necessary for sending comments.
type CommentConnection struct {
	liveWaku *LiveWaku
	socket   net.Conn
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
	defer cc.socket.Close()

	fmt.Fprintf(cc.socket, "GET / HTTP/1.0\r\n\r\n")
	status, err := bufio.NewReader(cc.socket).ReadString('\n')
	if err != nil {
		return NicoErrFromStdErr(err)
	}

	fmt.Println(status)

	return nil
}

// Close closes connection
func (cc CommentConnection) Close() {
	cc.socket.Close()
}

// NewCommentConnection returns a pointer to new CommentConnection
func NewCommentConnection(l *LiveWaku) *CommentConnection {
	return &CommentConnection{
		liveWaku: l,
	}
}
