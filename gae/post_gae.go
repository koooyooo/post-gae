package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/koooyooo/post-gae/gae/model"

	"cloud.google.com/go/storage"
	"google.golang.org/appengine"
)

func main() {
	http.HandleFunc("/", handle)
	appengine.Main()
}

func handle(w http.ResponseWriter, r *http.Request) {
	cxt := appengine.NewContext(r)
	if postcodes == nil {
		p, err := loadPostcodes(cxt)
		if err != nil {
			log.Fatal(err)
		}
		postcodes = p
	}
	for _, p := range postcodes {
		if p.Postcode == "1050011" {
			//fmt.Fprintln(w, p)
		}
	}
	fmt.Fprintln(w, "Hello, world!")
}

var postcodes []model.Postcode

func loadPostcodes(c context.Context) ([]model.Postcode, error) {
	client, err := storage.NewClient(c)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	reader, err := client.Bucket("dm-on-priv-post").Object("KEN_ALL.json").NewReader(c)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	var p []model.Postcode
	json.Unmarshal(b, &p)
	return p, nil
}
