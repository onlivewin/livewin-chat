package broker

import (
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gobwas/ws"
	"golang.org/x/sys/unix"
)

type Server struct {
	Handler  func(c *Conn, in []byte) error
	Epoller  *epoll
	channels sync.Map
}

func NewServer(handler func(c *Conn, in []byte) error) (server *Server, err error) {
	epoller, err := MkEpoll()
	if err != nil {
		panic(err)
	}

	server = &Server{
		Handler:  handler,
		Epoller:  epoller,
		channels: sync.Map{},
	}
	return server, nil
}

func (s *Server) StoreChannel(channel string, conn *Conn) {
	fd := websocketFD(conn.C)
	ro, _ := s.channels.LoadOrStore(channel, make(map[int]*Conn))
	ro.(map[int]*Conn)[fd] = conn
}

func (s *Server) BroacastConns(fn func(*Conn)) []*Conn {
	s.Epoller.SendMsg(fn)
	return nil
}

func (s *Server) BroacastChannelConns(channel string, fn func(*Conn)) []*Conn {
	if r, found := s.channels.Load(channel); found {
		if data, ok := r.(map[int]*Conn); ok {
			for _, c := range data {
				fn(c)
			}
		}
	}
	return nil
}

func (s *Server) Getchannels() (channels []string) {
	s.channels.Range(func(key, value any) bool {
		channels = append(channels, key.(string))
		return true
	})
	return
}

func (s *Server) Run(addr string) {
	go s.Start()
	http.HandleFunc("/", s.wsHandler)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) RemoveChannelWatcher(channel string, fd int) {
	if r, found := s.channels.Load(channel); found {
		delete(r.(map[int]*Conn), fd)
	}
}

func (s *Server) Start() error {
	for {
		connections, err := s.Epoller.Wait()
		if err != nil && err != unix.EINTR {
			log.Printf("Failed to epoll wait %v", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			header, err := ws.ReadHeader(conn.C)
			if err != nil {
				conn.Close()
				continue
			}

			switch header.OpCode {
			case ws.OpPing:
				ws.WriteFrame(conn.C, ws.NewPongFrame([]byte("")))
				continue
			case ws.OpContinuation | ws.OpPong:
				continue
			case ws.OpClose:
				log.Printf("got close message")
				conn.Close() //这边不需要收动close epool会自动处理
				continue
			default:
			}

			payload := make([]byte, header.Length)
			_, err = io.ReadFull(conn.C, payload)
			if err != nil {
				return err
			}
			if header.Masked {
				ws.Cipher(payload, header.Mask, 0)
			}
			err = s.Handler(conn, payload)
			if err != nil {
				log.Printf("[err] handler error %s", err)
				conn.Close()
			}
		}
	}
}

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	c, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}

	// 认证
	// if r.URL.Query().Get("token") != "abcd" {
	// 	w.Write([]byte("forbidden"))
	// 	return
	// }

	conn := &Conn{
		C: c,
		S: s,
	}
	if err := s.Epoller.Add(conn); err != nil {
		log.Printf("Failed to add connection %v", err)
		conn.Close()
	}
}
