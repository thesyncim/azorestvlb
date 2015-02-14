package main

import (
	"github.com/oschwald/maxminddb-golang"

	"encoding/json"
	"log"

	"net"
)

const DefaultDC = "EU"

type Datacenter struct {
	TargetContinents []string
	Servers          []*Server
}

func (d *Datacenter) isTagetTo(continent string) bool {
	for _, cc := range d.TargetContinents {
		if cc == continent {
			return true
		}
	}
	return false
}

func (d *Datacenter) DeadServers() (total, dead int) {
	total = len(d.Servers)
	for _, server := range d.Servers {
		if !(server).IsAlive() {
			dead++
		}
	}
	return
}

func (d *Datacenter) addServer(domain, vid string, mbo int64, n bool) {
	d.Servers = append(d.Servers, NewServer(domain, vid, mbo, n))
}

func (d *Datacenter) pickServer() *Server {
	load := float32(99999)
	sindex := 0
	for index, server := range d.Servers {

		if server.IsAlive() {
			if server.Load > 90 {
				return cloud.Datacenters[DefaultDC].pickServer()
			}
			if server.Load < load {
				load = server.Load

				sindex = index

			}

		}

	}

	return d.Servers[sindex]

}

type WowzaBalancer struct {
	Datacenters map[string]*Datacenter
}

func (w *WowzaBalancer) GetStats() (stats map[string]int64) {

	stats = make(map[string]int64)

	for _, datacenter := range w.Datacenters {
		for _, server := range datacenter.Servers {
			stats[server.Domain] = server.BytesOut
		}

	}
	return
}
func (c *WowzaBalancer) GetDatacenterByContinent(continent string) (dc *Datacenter) {

	for _, datacenter := range c.Datacenters {
		if datacenter.isTagetTo(continent) {
			return datacenter
		}
	}

	//no match use default
	return c.Datacenters[DefaultDC]
}

func (c *WowzaBalancer) AddDatacenter(id string, datacenter *Datacenter) {
	c.Datacenters[id] = datacenter
}

func (c *WowzaBalancer) PickServerByContinent(continent string) *Server {
	return c.GetDatacenterByContinent(continent).pickServer()
}

func (c *WowzaBalancer) getContinentByIp(Ip string) string {

	host, _, _ := net.SplitHostPort(Ip)
	Maxmind, err := maxminddb.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatalf("Error: failed to opem db : %s", err)
	}

	defer Maxmind.Close()
	ip := net.ParseIP(host)

	var mmrec interface{}
	err = Maxmind.Lookup(ip, &mmrec)

	if err != nil {
		log.Println("error lookup ip", err)
		return DefaultContinent
	}
	b, err := json.Marshal(mmrec)
	if err != nil {
		log.Println("error unmarshal mmrec:", err)
		return DefaultContinent
	}
	record := new(Maxmindrecord)
	err = json.Unmarshal(b, record)
	if err != nil {
		log.Println(err)
		return DefaultContinent
	}
	log.Println(Ip, record.Continent.Code)
	return record.Continent.Code
}
