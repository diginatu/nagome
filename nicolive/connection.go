package nicolive

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	keepAliveDuration       = time.Minute
	connectionWriteDeadline = 3 * time.Second
	connectionReadDeadline  = 3 * time.Second
)

type proceedConnMes func(m string)

// connection is an abstract struct to manage connection for comment and antenna.
type connection struct {
	Wg     sync.WaitGroup
	Ctx    context.Context
	Cancal context.CancelFunc
	Ev     EventReceiver

	addrPort string

	conn           net.Conn
	rw             bufio.ReadWriter
	wmu            sync.Mutex
	disconnecting  bool
	proceedMessage proceedConnMes
}

func newConnection(addrPort string, proceedMessage proceedConnMes, ev EventReceiver) *connection {
	if ev == nil {
		ev = &defaultEventReceiver{}
	}

	c := &connection{
		addrPort:       addrPort,
		Ev:             ev,
		proceedMessage: proceedMessage,
	}
	c.Ctx, c.Cancal = context.WithCancel(context.Background())

	return c
}

func (c *connection) Connect(ctx context.Context) error {
	nerr := c.open(ctx)
	if nerr != nil {
		// No need to disconnect.
		return nerr
	}

	c.Wg.Add(1)
	go c.receiveStream()

	return nil
}

func (c *connection) open(ctx context.Context) error {
	var err error

	d := &net.Dialer{
		KeepAlive: keepAliveDuration,
	}

	c.conn, err = d.DialContext(ctx, "tcp", c.addrPort)
	if err != nil {
		return ErrFromStdErr(err)
	}

	c.rw = bufio.ReadWriter{
		Reader: bufio.NewReader(c.conn),
		Writer: bufio.NewWriter(c.conn),
	}

	return nil
}

func (c *connection) Send(m string) error {
	c.wmu.Lock()
	defer c.wmu.Unlock()

	c.conn.SetWriteDeadline(time.Now().Add(connectionWriteDeadline))

	fmt.Fprint(c.rw, m)
	err := c.rw.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (c *connection) receiveStream() {
	defer c.Wg.Done()
	for {
		select {
		case <-c.Ctx.Done():
			return
		default:
			bd, err := c.rw.ReadString('\x00')
			if err != nil {
				nerr, ok := err.(net.Error)
				if ok && nerr.Temporary() {
					c.Ev.ProceedNicoEvent(&Event{
						Type:    EventTypeErr,
						Content: ErrFromStdErr(err),
					})
					continue
				}

				if c.disconnecting {
					return
				}

				c.Ev.ProceedNicoEvent(&Event{
					Type:    EventTypeErr,
					Content: ErrFromStdErr(err),
				})
				go c.Disconnect()
				return
			}

			// strip null char and proceed
			c.proceedMessage(bd[:len(bd)-1])
		}
	}
}

// Disconnect close and disconnect
// terminate all goroutines and wait to exit
func (c *connection) Disconnect() error {
	fmt.Println(2)
	if c.disconnecting {
		return MakeError(ErrOther, "already disconnecting")
	}
	c.disconnecting = true
	defer func() { c.disconnecting = false }()

	c.Cancal()
	c.conn.Close()

	c.Wg.Wait()
	fmt.Println(3)

	return nil
}
