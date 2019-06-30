package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/koooyooo/post-gae/gae/model"

	"cloud.google.com/go/storage"
	"google.golang.org/appengine"
)

func main() {
	http.HandleFunc("/", handle)
	http.HandleFunc("/v1/postcodes/", find)
	appengine.Main()
}

func handle(w http.ResponseWriter, r *http.Request) {
	cxt := appengine.NewContext(r)
	loadCache(cxt)

	postcodes := postmap["1060032"]
	postcodesStr, err := PostcodesForView(postcodes)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(500)
		return
	}
	fmt.Fprintln(w, postcodesStr)
	fmt.Fprintln(w, "Hello, world!")
}

func find(w http.ResponseWriter, r *http.Request) {
	cxt := appengine.NewContext(r)
	loadCache(cxt)
	id := strings.TrimPrefix(r.URL.Path, "/v1/postcodes/")

	var postcodes []model.Postcode
	if len(id) < 3 || 7 < len(id) {
		postcodes = []model.Postcode{}
	} else if len(id) == 7 {
		postcodes = postmap[id]
		if postcodes == nil {
			postcodes = []model.Postcode{}
		}
	} else {
		postcodes = []model.Postcode{}
		for k, v := range postmap {
			if strings.HasPrefix(k, id) {
				postcodes = append(postcodes, v...)
			}
		}
	}
	postcodesStr, err := PostcodesForView(postcodes)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(500)
		return
	}
	fmt.Fprintln(w, postcodesStr)
}

func PostcodesForView(postcodes []model.Postcode) (string, error) {
	postcodesJSON, err := json.Marshal(postcodes)
	if err != nil {
		return "", err
	}
	return string(postcodesJSON), nil
}

var postmap map[string][]model.Postcode

func loadCache(c context.Context) error {
	if postmap != nil {
		return nil
	}
	postcodes, err := loadPostcodes(c)
	if err != nil {
		return err
	}
	postmap = map[string][]model.Postcode{}
	for _, p := range postcodes {
		v, ok := postmap[p.Postcode]
		if !ok {
			postmap[p.Postcode] = []model.Postcode{p}
		} else {
			v = append(v, p)
		}
	}
	return nil
}

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
