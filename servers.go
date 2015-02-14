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

const updateInterval = 5 //seconds

type serverStats struct {
	XMLName                  xml.Name `xml:"WowzaStreamingEngine"`
	ConnectionsCurrent       int64    `xml:"ConnectionsCurrent"`
	ConnectionsTotal         int64    `xml:"ConnectionsTotal"`
	ConnectionsTotalAccepted int64    `xml:"ConnectionsTotalAccepted"`
	ConnectionsTotalRejected int64    `xml:"ConnectionsTotalRejected"`
	MessagesInBytesRate      float32  `xml:"MessagesInBytesRate"`
	MessagesOutBytesRate     float32  `xml:"MessagesOutBytesRate"`
}

type nginxserverStats struct {
	XMLName   xml.Name `xml:"rtmp"`
	Bytes_out int64    `xml:"bw_out"`
}

type Server struct {
	MaxBytesOut int64
	LastUpdated time.Time
	Hits        int64
	Load        float32
	VID         string
	Domain      string
	alive       bool
	nginx       bool
	BytesOut    int64
}

var timeout = time.Duration(5 * time.Second)

func NewServer(domain string, vid string, mbo int64, isnginx bool) (s *Server) {
	s = new(Server)
	s.Domain = domain
	s.VID = vid
	s.MaxBytesOut = mbo
	s.nginx = isnginx
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

		if s.nginx {
			resp, err := client.Get("http://" + s.Domain + "/stats")
			if err != nil {
				s.alive = false
				log.Printf("Error : %s", err)
				continue
			}
			s.alive = true

			var body bytes.Buffer
			io.Copy(&body, resp.Body)
			resp.Body.Close()

			nstats := nginxserverStats{}

			err = xml.Unmarshal(body.Bytes(), &nstats)
			if err != nil {
				fmt.Printf("error: %v", err)

			}
			s.Load = 0

			if nstats.Bytes_out > 0 {
				s.Load = float32(nstats.Bytes_out) / float32(s.MaxBytesOut) * 100
			}

			log.Println(s.Load, nstats.Bytes_out, s.MaxBytesOut)

			s.BytesOut = nstats.Bytes_out

		} else {
			resp, err := client.Get("http://" + s.Domain + ":8086/connectioncounts")
			if err != nil {
				s.alive = false
				log.Printf("Error : %s", err)
				continue
			}
			s.alive = true

			var body bytes.Buffer
			io.Copy(&body, resp.Body)
			resp.Body.Close()

			stats := serverStats{}

			err = xml.Unmarshal(body.Bytes(), &stats)
			if err != nil {
				fmt.Printf("error: %v", err)

			}
			s.Load = 0

			if stats.MessagesOutBytesRate > 0 {
				s.Load = (stats.MessagesOutBytesRate / float32(s.MaxBytesOut)) * 100
			}

			s.BytesOut = int64(stats.MessagesOutBytesRate)
			log.Println(s.Load, stats.MessagesOutBytesRate, s.MaxBytesOut)

		}

		//fmt.Println(s.Load, s.VID)
	}

}

func (s *Server) IsAlive() bool {
	return s.alive == true
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}
