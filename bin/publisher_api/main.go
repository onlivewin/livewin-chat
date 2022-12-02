package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/widaT/livewin-chat/pkg"
	"github.com/widaT/livewin-chat/pkg/publisher"
)

func main() {
	var port string
	var registerAddr string
	flag.StringVar(&registerAddr, "r", "localhost:9655", "register addr")
	flag.StringVar(&port, "p", "9653", "http service port")
	flag.Parse()

	discovery := pkg.NewUdpDiscovery(registerAddr, "hairy_crab")
	service, err := publisher.NewService(discovery)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/broadcast", func(w http.ResponseWriter, r *http.Request) {
		message, err := ioutil.ReadAll(r.Body)
		if err != nil || len(message) == 0 {
			w.Write([]byte("failed"))
			return
		}

		log.Printf("post data  msg:%q", message)
		err = service.Broadcast(message)
		if err != nil {
			log.Printf("[err] bloadcast msg:%q err:%s", message, err)
			w.Write([]byte(fmt.Sprintf("[err] bloadcast msg:%q err:%s", message, err)))
			return
		}
		log.Printf("post data  msg:%q success", message)
		w.Write([]byte("success"))
	})

	http.HandleFunc("/broadcastinchannel", func(w http.ResponseWriter, r *http.Request) {
		channel := r.URL.Query().Get("channel")
		if len(channel) == 0 {
			w.Write([]byte("failed"))
			return
		}
		message, err := ioutil.ReadAll(r.Body)
		if err != nil || len(message) == 0 {
			w.Write([]byte("failed"))
			return
		}

		err = service.BroadcastInGroup(channel, message)
		if err != nil {
			log.Printf("[err]channel:%q msg:%q err:%s", channel, message, err)
			w.Write([]byte(fmt.Sprintf("[err]channel:%q msg:%q err:%s", channel, message, err)))
			return
		}
		log.Printf("post data channel:%q msg:%q success", channel, message)
		w.Write([]byte("success"))
	})

	http.HandleFunc("/channels", func(w http.ResponseWriter, r *http.Request) {
		resp, err := service.Channels()
		if err != nil {
			log.Printf("[err]%s", err)
			return
		}

		jsonStr, err := json.Marshal(resp)
		if err != nil {
			log.Printf("[err]%s", err)
			return
		}
		w.Write(jsonStr)
	})

	log.Printf("publisher-api-svc run on port:%s", port)
	http.ListenAndServe(":"+port, nil)

}
