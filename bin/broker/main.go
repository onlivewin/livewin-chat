package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"syscall"

	"github.com/widaT/livewin-chat/pkg"
	"github.com/widaT/livewin-chat/pkg/broker"
)

var server *broker.Server
var grpcPort = 6365

func HandleMsg(c *broker.Conn, in []byte) error {
	log.Printf("got messge %q", in)

	var message = broker.Message{}
	err := json.Unmarshal(in, &message)
	if err != nil {
		return err
	}
	switch message.Type {
	case broker.TJoin:
		if len(message.Channel) == 0 {
			c.Close()
			return nil
		}
		if c.Channel == "" {
			c.Channel = message.Channel
			c.S.StoreChannel(c.Channel, c)
		}
	default:
		return errors.New("unreachable")
	}
	return nil
}

func main() {
	var isTLS bool
	var wsPort string
	var registerAddr string
	//var
	flag.IntVar(&grpcPort, "g", 6365, "grpc service port")
	flag.BoolVar(&isTLS, "tls", false, "enable tls")
	flag.StringVar(&wsPort, "p", "8888", "websocket service port")
	flag.StringVar(&registerAddr, "r", "localhost:9655", "register addr")
	flag.Parse()

	// ulimit放开限制
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	var err error
	server, err = broker.NewServer(HandleMsg)
	if err != nil {
		panic(err)
	}

	//grpc 服务
	go broker.GrpcService(server, fmt.Sprintf(":%d", grpcPort))

	//register 服务
	discovery := pkg.NewSimpleDiscovery(registerAddr, "hairy_crab")

	discovery.Register(fmt.Sprintf("%d", grpcPort))

	server.Run(":" + wsPort)
}
