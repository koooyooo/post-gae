package main

import (
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
	//postcodes, err := LoadPostcode()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//for _, p := range postcodes {
	//	fmt.Println(p)
	//}

	http.HandleFunc("/", handle)
	appengine.Main()
}

func handle(w http.ResponseWriter, r *http.Request) {
	cxt := appengine.NewContext(r)
	client, err := storage.NewClient(cxt)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	reader, err := client.Bucket("dm-on-priv-post").Object("KEN_ALL.json").NewReader(cxt)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(w, len(b))
	fmt.Fprintln(w, "Hello, world!")
}

func LoadPostcode() ([]model.Postcode, error) {
	b, err := ioutil.ReadFile("KEN_ALL.json")
	if err != nil {
		return nil, err
	}
	var postcodes []model.Postcode
	json.Unmarshal(b, &postcodes)
	return postcodes, nil
}
