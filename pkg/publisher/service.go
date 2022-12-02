package publisher

import (
	"context"
	"log"
	"sync"

	"github.com/widaT/livewin-chat/pkg"
	grpcproto "github.com/widaT/livewin-chat/pkg/proto/broker"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var opts = grpc.WithTransportCredentials(insecure.NewCredentials())

type Service struct {
	*sync.Mutex
	brokers     map[string]grpcproto.BrokerClient
	discoveryer pkg.Discoveryer
}

func (s *Service) getBrokers() map[string]grpcproto.BrokerClient {
	addrs := s.discoveryer.Discovery()
	log.Printf("addrs %s", addrs)
	var brokers = make(map[string]grpcproto.BrokerClient)
	s.Lock()
	brokers_temp := maps.Clone(s.brokers)
	s.Unlock()

	for _, addr := range addrs {
		if c, found := brokers_temp[addr]; found {
			brokers[addr] = c
		} else {
			conn, err := grpc.Dial(addr, opts)
			if err != nil {
				log.Fatal(err)
			}
			brokers[addr] = grpcproto.NewBrokerClient(conn)
		}
	}
	s.Lock()
	s.brokers = brokers
	s.Unlock()
	return brokers
}

func NewService(discoveryer pkg.Discoveryer) (*Service, error) {
	var brokers = make(map[string]grpcproto.BrokerClient)
	return &Service{
		brokers:     brokers,
		discoveryer: discoveryer,
		Mutex:       &sync.Mutex{},
	}, nil
}

func (s *Service) Broadcast(message []byte) error {
	for _, client := range s.getBrokers() {
		_, err := client.Broadcast(context.Background(), &grpcproto.BroadcastReq{
			Proto: &grpcproto.Proto{
				Body: message,
			},
		})
		if err != nil {
			log.Printf("[err]%v", err)
			return err
		}
	}
	return nil
}

func (s *Service) BroadcastInGroup(channel string, message []byte) error {
	for _, client := range s.getBrokers() {
		_, err := client.BroadcastInGroup(context.Background(), &grpcproto.BroadcastChannelReq{
			ChannelID: channel,
			Proto: &grpcproto.Proto{
				Body: []byte(message),
			},
		})
		if err != nil {
			log.Printf("[err]%v", err)
			return err
		}
	}
	return nil
}
func (s *Service) Channels() (map[string][]string, error) {
	ret := make(map[string][]string)
	for key, client := range s.getBrokers() {
		channels, err := client.Channels(context.Background(), &grpcproto.ChannelsReq{})
		if err != nil {
			log.Printf("[err]%v", err)
			return nil, err
		}
		ret[key] = channels.Channels
	}
	return ret, nil
}
