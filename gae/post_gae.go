package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/profiler"

	"github.com/koooyooo/post-gae/gae/model"

	"cloud.google.com/go/storage"
	"google.golang.org/appengine"
)

type Done struct{}

func main() {
	if err := profiler.Start(profiler.Config{}); err != nil {
		log.Fatalf("init profiler fails: %v", err)
	}

	d := make(chan Done)
	go func() {
		checkAndLoadCache(context.Background())
		d <- Done{}
	}()

	http.HandleFunc("/", handle)
	http.HandleFunc("/v1/postcodes/", find)
	<-d
	appengine.Main()
}

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func find(w http.ResponseWriter, r *http.Request) {
	pathPostCode := strings.TrimPrefix(r.URL.Path, "/v1/postcodes/")

	var results []model.Postcode
	if 3 <= len(pathPostCode) && len(pathPostCode) <= 6 {
		for _, v := range postcodes {
			if strings.HasPrefix(v.Postcode, pathPostCode) {
				results = append(results, v)
			}
		}
	} else if len(pathPostCode) == 7 {
		results = postmap[pathPostCode]
		if results == nil {
			results = []model.Postcode{}
		}
	} else {
		queries := r.URL.Query()
		pref := queries["prefecture"][0]
		if pref != "" {
			results = postcodes
			results = FilterByParam(results, queries, "prefecture", func(p model.Postcode, params []string) bool {
				paramPref := params[0]
				return strings.Contains(p.Prefecture, paramPref)
			})
			results = FilterByParam(results, queries, "city", func(p model.Postcode, params []string) bool {
				paramCity := params[0]
				return strings.Contains(p.City, paramCity)
			})
			results = FilterByParam(results, queries, "area", func(p model.Postcode, params []string) bool {
				paramArea := params[0]
				return strings.Contains(p.Area, paramArea)
			})
		} else {
			results = []model.Postcode{}
		}
	}
	strResults, err := PostcodesForView(results)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(500)
		return
	}
	fmt.Fprintln(w, strResults)
}

func FilterByParam(pCodes []model.Postcode, params url.Values, paramName string, match func(model.Postcode, []string) bool) []model.Postcode {
	p, ok := params[paramName]
	if !ok {
		return pCodes
	}
	matched := []model.Postcode{}
	for _, r := range pCodes {
		if match(r, p) {
			matched = append(matched, r)
		}
	}
	return matched
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
	loaded, err := loadPostcodes(c)
	t1 := time.Now()
	if err != nil {
		log.Fatal(err)
	}
	postcodes = loaded
	postmap = map[string][]model.Postcode{}
	for _, p := range loaded {
		v, ok := postmap[p.Postcode]
		if !ok {
			postmap[p.Postcode] = []model.Postcode{p}
		} else {
			v = append(v, p)
		}
	}
	t2 := time.Now()
	fmt.Println("prepare", t2.Sub(t1))
	return nil
}

func loadPostcodes(c context.Context) ([]model.Postcode, error) {
	t1 := time.Now()
	client, err := storage.NewClient(c)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	reader, err := client.Bucket("dm-on-priv-post").Object("KEN_ALL.json").NewReader(c)
	if err != nil {
		log.Fatal(err)
	}
	t2 := time.Now()

	if err != nil {
		log.Fatal(err)
	}
	var p []model.Postcode
	json.NewDecoder(reader).Decode(&p) // better than... b, err := ioutil.ReadAll(reader) -> json.Unmarshal(b, &p)
	t3 := time.Now()
	fmt.Println("load-storage", t2.Sub(t1))
	fmt.Println("load-marshal", t3.Sub(t2))
	return p, nil
}
