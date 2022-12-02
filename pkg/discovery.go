package pkg

import (
	"log"
	"net"
	"strings"
	"time"
)

type Discoveryer interface {
	Discovery() []string
}

type SimpleDiscovery struct {
	Conn           *net.UDPConn
	Service        []string
	LastUpdateTime time.Time
	Channel        string
}

func NewSimpleDiscovery(addr string, channel string) *SimpleDiscovery {
	uAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Panic(err)
	}
	conn, err := net.DialUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0}, uAddr)
	if err != nil {
		log.Panic(err)
	}
	return &SimpleDiscovery{
		Conn:           conn,
		Channel:        channel,
		LastUpdateTime: time.Now(),
	}
}

func NewUdpDiscovery(addr string, channel string) Discoveryer {
	return NewSimpleDiscovery(addr, channel)
}

func (u *SimpleDiscovery) Query(channel string) {
	data := make([]byte, len(channel)+1)
	data[0] = 2
	copy(data[1:], []byte(channel))

	u.Conn.Write(data)

	buf := make([]byte, 2048)
	n, err := u.Conn.Read(buf)
	if err != nil {
		return
	}
	log.Printf("get servers %s", buf[:n])
	u.Service = strings.Split(string(buf[:n]), ",")
}

func (u *SimpleDiscovery) Discovery() []string {
	if len(u.Service) == 0 || time.Since(u.LastUpdateTime) > 10*time.Second {
		u.Query(u.Channel)
		u.LastUpdateTime = time.Now()
	}
	return u.Service
}

func (u *SimpleDiscovery) Register(port string) {
	data := make([]byte, len(u.Channel)+2+len(port))
	data[0] = 1
	copy(data[1:], []byte(u.Channel+"|"+port))
	go func() {
		for {
			u.Conn.Write(data)
			time.Sleep(3 * time.Second)
		}
	}()
}
