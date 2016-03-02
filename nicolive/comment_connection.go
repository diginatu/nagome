package nicolive

import (
	"bufio"
	"fmt"
	"net"
)

// CommentConnection is a struct to manage sending/receiving comments.
// This struct automatically submits NULL character to reserve connection and
// get the PostKey, which is necessary for sending comments.
type CommentConnection struct {
	liveWaku *LiveWaku
}

// Connect Connect to nicolive and start receiving comment
func (cc CommentConnection) Connect() NicoError {
	conn, err := net.Dial("tcp", "google.com:80")
	if err != nil {
		return NicoErrFromStdErr(err)
	}

	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return NicoErrFromStdErr(err)
	}

	fmt.Println(status)

	return nil
}

// NewCommentConnection returns a pointer to new CommentConnection
func NewCommentConnection(l *LiveWaku) *CommentConnection {
	return &CommentConnection{
		liveWaku: l,
	}
}
