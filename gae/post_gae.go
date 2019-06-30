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
	checkAndLoadCache(cxt)

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
	checkAndLoadCache(cxt)
	id := strings.TrimPrefix(r.URL.Path, "/v1/postcodes/")

	var results []model.Postcode
	if 3 <= len(id) && len(id) <= 6 {
		for k, v := range postmap {
			if strings.HasPrefix(k, id) {
				results = append(results, v...)
			}
		}
	} else if len(id) == 7 {
		results = postmap[id]
		if results == nil {
			results = []model.Postcode{}
		}
	} else {
		results = []model.Postcode{}
	}

	params := r.URL.Query()
	v, ok := params["prefecture"]
	if ok {
		pref := v[0]
		fmt.Println(pref)
	}

	strResults, err := PostcodesForView(results)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(500)
		return
	}
	fmt.Fprintln(w, strResults)
}

func PostcodesForView(postcodes []model.Postcode) (string, error) {
	postcodesJSON, err := json.Marshal(postcodes)
	if err != nil {
		return "", err
	}
	return string(postcodesJSON), nil
}

var postcodes []model.Postcode
var postmap map[string][]model.Postcode

func checkAndLoadCache(c context.Context) error {
	if postmap != nil {
		return nil
	}
	p, err := loadPostcodes(c)
	if err != nil {
		return err
	}
	postcodes = p
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
