package main

import (

	//"fmt"
	"html/template"
	//"io/ioutil"

	"log"
	"net/http"
)

var cloud *Cloud

type httpHandler struct {
	templateCache *template.Template
}

func (lb *httpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	co := cloud.getContinentByIp(r.RemoteAddr)
	VID := cloud.PickServerByContinent(co).VID
	err := lb.templateCache.ExecuteTemplate(rw, "template.html", VID)
	if err != nil {
		log.Println(err)
	}
}

func main() {

	cloud = new(Cloud)
	cloud.Datacenters = make(map[string]*Datacenter)
	cloud.AddDatacenter(
		"EU",
		&Datacenter{
			TargetContinents: []string{"EU", "AF", "AS"},
			Servers: []*Server{
				NewServer("wms1.azorestv.com", "110"),
				NewServer("wms2.azorestv.com", "112"),
			},
		})

	//cloud.AddDatacenter(
	//	"USA",
	//	&Datacenter{
	//		TargetContinents: []string{"OC", "AN", "SA", "NA"},
	//		Servers: []*Server{
	//			NewServer("wms2.azorestv.com", "112"),
	//		},
	//	})
	var err error

	lb := new(httpHandler)
	lb.templateCache = new(template.Template)
	lb.templateCache, err = template.New("template").ParseFiles("template.html")

	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", lb)

	http.ListenAndServe(":8080", nil)

}
