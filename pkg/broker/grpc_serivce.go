package broker

import (
	"context"
	"log"
	"net"

	"github.com/gobwas/ws"
	grpcproto "github.com/widaT/livewin-chat/pkg/proto/broker"
	"google.golang.org/grpc"
)

type BrokerServiceImpl struct {
	Server *Server
}

func (b *BrokerServiceImpl) Broadcast(ctx context.Context, req *grpcproto.BroadcastReq) (*grpcproto.BroadcastReply, error) {
	if err := b.BroadcastMessage(req.Proto.Body); err != nil {
		return nil, err
	}
	return &grpcproto.BroadcastReply{}, nil
}

func (b *BrokerServiceImpl) BroadcastInGroup(ctx context.Context, req *grpcproto.BroadcastChannelReq) (*grpcproto.BroadcastChannelReply, error) {
	if err := b.BroadcastChannelMessage(req.ChannelID, req.Proto.Body); err != nil {
		return nil, err
	}
	return &grpcproto.BroadcastChannelReply{}, nil
}

func (b *BrokerServiceImpl) Channels(ctx context.Context, req *grpcproto.ChannelsReq) (*grpcproto.ChannelsReply, error) {

	channels := b.Server.Getchannels()

	return &grpcproto.ChannelsReply{
		Channels: channels,
	}, nil
}

func (b *BrokerServiceImpl) BroadcastMessage(message []byte) error {
	frame := ws.NewTextFrame(message)
	b.Server.BroacastConns(func(c *Conn) {
		ws.WriteFrame(c.C, frame)
	})
	return nil
}

func (b *BrokerServiceImpl) BroadcastChannelMessage(channel string, message []byte) error {
	frame := ws.NewTextFrame(message)
	b.Server.BroacastChannelConns(channel, func(c *Conn) {
		ws.WriteFrame(c.C, frame)
	})
	return nil
}

func GrpcService(server *Server, addr string) {
	grpcServer := grpc.NewServer()
	grpcproto.RegisterBrokerServer(grpcServer, &BrokerServiceImpl{Server: server})
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	grpcServer.Serve(lis)
}
