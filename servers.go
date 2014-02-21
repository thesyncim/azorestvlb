package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"time"
	//"log"
	"encoding/xml"
	"net/http"
	//"strings"
)

const MaxBytesOut = 62500000 //500Mb

const updateInterval = 5 //seconds

type serverStats struct {
	XMLName                  xml.Name `xml:"WowzaMediaServer"`
	ConnectionsCurrent       int64    `xml:"ConnectionsCurrent"`
	ConnectionsTotal         int64    `xml:"ConnectionsTotal"`
	ConnectionsTotalAccepted int64    `xml:"ConnectionsTotalAccepted"`
	ConnectionsTotalRejected int64    `xml:"ConnectionsTotalRejected"`
	MessagesInBytesRate      float32  `xml:"MessagesInBytesRate"`
	MessagesOutBytesRate     float32  `xml:"MessagesOutBytesRate"`
}

type Server struct {
	MaxBytesOut int64
	LastUpdated time.Time
	Hits        int64
	Load        float32
	VID         string
	Domain      string
	alive       bool
	Stats       *serverStats
}

var timeout = time.Duration(5 * time.Second)

func NewServer(domain string, vid string) (s *Server) {
	s = new(Server)
	s.Domain = domain
	s.VID = vid
	s.Stats = new(serverStats)
	go s.update()
	return
}

func (s *Server) update() {

	interval := time.Tick(updateInterval * time.Second)

	for _ = range interval {

		transport := http.Transport{
			Dial: dialTimeout,
		}

		client := http.Client{
			Transport: &transport,
		}

		resp, err := client.Get("http://" + s.Domain + ":8086/connectioncounts")
		if err != nil {
			s.alive = false
			log.Printf("Error : %s", err)
			return
		}
		s.alive = true

		var body bytes.Buffer
		io.Copy(&body, resp.Body)

		err = xml.Unmarshal(body.Bytes(), s.Stats)
		if err != nil {
			fmt.Printf("error: %v", err)

		}
		s.Load = 0

		if s.Stats.MessagesOutBytesRate > 0 {
			s.Load = (s.Stats.MessagesOutBytesRate / MaxBytesOut) * 100
		}

		fmt.Println(s.Load, s.VID)
	}

}

func (s *Server) IsAlive() bool {
	return s.alive == true
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}
