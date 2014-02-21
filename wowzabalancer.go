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
		if !(&server).IsAlive() {
			dead++
		}
	}
	return
}

func (d *Datacenter) addServer(domain, vid string) {
	d.Servers = append(d.Servers, NewServer(domain, vid))
}

func (d *Datacenter) pickServer() *Server {
	load := float32(99999)
	sindex := 0
	for index, server := range d.Servers {
		log.Println(server)

		if server.IsAlive() {
			if server.Load < load {
				load = server.Load
				log.Println("pick server", load, d.Servers)
				sindex = index
				log.Println(index, sindex)
				log.Println(d.Servers[sindex])
			}

		}

	}
	return d.Servers[sindex]

}

type Cloud struct {
	Datacenters map[string]*Datacenter
}

func (c *Cloud) GetDatacenterByContinent(continent string) (dc *Datacenter) {

	for _, datacenter := range c.Datacenters {
		if datacenter.isTagetTo(continent) {
			return datacenter
		}
	}

	//no match use default
	return c.Datacenters[DefaultDC]
}

func (c *Cloud) AddDatacenter(id string, datacenter *Datacenter) {
	c.Datacenters[id] = datacenter
}

func (c *Cloud) PickServerByContinent(continent string) *Server {
	return c.GetDatacenterByContinent(continent).pickServer()
}

func (c *Cloud) getContinentByIp(Ip string) string {
	Maxmind, err := maxminddb.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatalf("Error: failed to opem db : %s", err)
	}

	defer Maxmind.Close()
	ip := net.ParseIP(Ip)

	mmrec, err := Maxmind.Lookup(ip)

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

	return record.Continent.Code
}
