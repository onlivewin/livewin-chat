package main

import (
	"log"
	"net"
	"strings"
	"time"
)

var maxLife = 10 * time.Second
var bigMap = make(map[string]map[string]time.Time)

func main() {
	listen, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 9655,
	})
	if err != nil {
		log.Fatal("Listen failed, err: ", err)
		return
	}
	defer listen.Close()
	for {
		var data [1024]byte
		n, addr, err := listen.ReadFromUDP(data[:]) // 接收数据
		if err != nil {
			log.Println("read udp failed, err: ", err)
			continue
		}

		rawData := data[:n]
		switch rawData[0] {
		case 1: //set
			payload := string(rawData[1:])
			parts := strings.Split(payload, "|")
			if len(parts) != 2 {
				continue
			}
			channel, port := parts[0], parts[1]
			ip, _, _ := net.SplitHostPort(addr.String())
			key := ip + ":" + port
			if a, found := bigMap[channel]; found {
				a[key] = time.Now()
			} else {
				bigMap[channel] = make(map[string]time.Time)
				bigMap[channel][key] = time.Now()
			}
		case 2: //query
			log.Printf("got query rawData:%s", rawData[1:])
			if a, found := bigMap[string(rawData[1:])]; found {
				var hosts []string
				for k, v := range a {
					if time.Since(v) > maxLife {
						delete(a, k)
						continue
					}
					hosts = append(hosts, k)
				}
				listen.WriteToUDP([]byte(strings.Join(hosts, ",")), addr)
			} else {
				listen.WriteToUDP([]byte(""), addr)
			}
		}
	}
}
