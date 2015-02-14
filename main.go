package main

import (
	"fmt"
	"html/template"
	//"io/ioutil"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	bandwidthGold   = 22222222 //500Mb
	bandwidthSilver = 22222222 //177Mb
)

var cloud *WowzaBalancer

type httpHandler struct {
	templateCache *template.Template
}

func (lb *httpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	co := cloud.getContinentByIp(r.RemoteAddr)
	VID := cloud.PickServerByContinent(co).VID
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	if strings.Contains(r.URL.Path, "@stats") {
		b, err := json.MarshalIndent(cloud.GetStats(), "", "  ")
		if err != nil {
			log.Println("error:", err)
		}
		var total int64
		for _, server := range cloud.GetStats() {
			total += server
		}
		rw.Write(append(b, []byte(strconv.Itoa(int(total)))...))
		return
	} else if strings.Contains(r.URL.Path, "mobile") {
		VID = "410"
	}

	a := `<div class="desktop"><iframe style="border: none; overflow: hidden; width: 720px; height: 405px;" src="http://www.azorestv.com/embed.php?id=%s&amp;w=720&amp;h=405" frameborder="0" scrolling="no" width="320" height="240"></iframe></div>
<div class="mobile"><iframe style="border: none; overflow: hidden; width: 320px; height: 180px;" src="http://www.azorestv.com/embed.php?id=%s&amp;w=320&amp;h=180" frameborder="0" scrolling="no" width="320" height="240"></iframe></div>
`
	_, err := fmt.Fprint(rw, fmt.Sprintf(a, VID, VID))
	if err != nil {
		log.Println(err)
	}
}

func main() {

	cloud = new(WowzaBalancer)
	cloud.Datacenters = make(map[string]*Datacenter)
	cloud.AddDatacenter(
		"EU",
		&Datacenter{
			TargetContinents: []string{"EU", "AF", "AS"},
			Servers: []*Server{
				NewServer("azorestv.com", "410", bandwidthSilver, false),
				NewServer("192.99.78.97", "799", bandwidthSilver, true),
			},
		})

	cloud.AddDatacenter(
		"USA",
		&Datacenter{
			TargetContinents: []string{"OC", "AN", "SA", "NA"},
			Servers: []*Server{
				NewServer("192.99.78.97", "799", bandwidthSilver, true),
			},
		})
	var err error

	lb := new(httpHandler)
	lb.templateCache = new(template.Template)
	lb.templateCache, err = template.New("template").ParseFiles("template.html")

	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", lb)

	http.ListenAndServe(":5555", nil)

}
