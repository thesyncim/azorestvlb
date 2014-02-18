package main

import (
	//"fmt"
	"encoding/json"
	"github.com/oschwald/maxminddb-golang"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	//"time"
	//strconv"
)

//AF	Africa
//AN	Antarctica
//AS	Asia
//EU	Europe
//NA	North america
//OC	Oceania
//SA	South america

var (
	serversUS = []string{"OC", "AN", "SA", "NA"}
	serversEU = []string{"EU", "AF", "AS"}
)

type Maxmindrecord struct {
	Continent          Continent
	Country            Country
	Registered_country Country
}

type Country struct {
	Geoname_id int64
	Iso_code   string
	Names      map[string]string
}

type Continent struct {
	Code       string
	Geoname_id int64
	Names      map[string]string
}

type VitecLB struct {
	maxminddb     maxminddb.Reader
	templateCache *template.Template
}

func (lb *VitecLB) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	ip := net.ParseIP("5.39.99.224")
	record, err := lb.maxminddb.Lookup(ip)

	if err != nil {
		log.Fatal(err)
	}
	b, err := json.Marshal(record)
	if err != nil {
		log.Println("error:", err)

	}
	log.Println(record)
	mmrec := new(Maxmindrecord)
	err = json.Unmarshal(b, mmrec)
	if err != nil {
		log.Println("error:", err)

	}
	log.Println(mmrec.Continent.Code)
	var ID = "101"
	err = lb.templateCache.ExecuteTemplate(rw, "template.html", ID)

	if err != nil {
		log.Println(err)
	}
}

func main() {

	// Error checking elided

	db, err := maxminddb.Open("GeoLite2-Country.mmdb")

	if err != nil {
		log.Fatal(err)
	}
	lb := new(VitecLB)
	lb.maxminddb = db
	lb.templateCache = new(template.Template)
	lb.templateCache, err = template.New("template").ParseFiles("template.html")
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", lb)
	go func() {
		http.ListenAndServe(":8080", nil)
		db.Close()

	}()

	for i := 1; i <= 1; i++ {
		go func() {
			resp, err := http.Get("http://localhost:8080/")
			if err != nil {
				log.Fatalln("failed http.get", err)
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("failed read resp.Boby", err)
			}
			log.Println(string(body))

		}()

	}
	select {}

}
