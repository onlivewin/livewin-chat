package broker

import (
	"log"
	"net"
)

type Conn struct {
	Channel string
	C       net.Conn
	S       *Server
}

func (c *Conn) Close() error {
	fd := websocketFD(c.C)
	c.S.Epoller.Remove(c)

	if len(c.Channel) > 0 {

		c.S.RemoveChannelWatcher(c.Channel, fd)
		log.Printf("remove fd:%d from channel %s ", fd, c.Channel)
	}

	if c.C != nil {
		c.C.Close()
	}
	return nil
}
