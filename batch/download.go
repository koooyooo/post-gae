package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/koooyooo/post-gae/batch/model"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func main() {
	//fmt.Println("Hello Download")
	//err := DownloadFile("https://www.post.japanpost.jp/zipcode/dl/kogaki/zip/ken_all.zip", "ken_all.zip")
	//if err != nil {
	//	log.Fatal(err.Error())
	//}
	//
	//err = Unzip("ken_all.zip", ".")
	//if err != nil {
	//	log.Fatal(err.Error())
	//}
	//
	//bSJIS, err := ioutil.ReadFile("KEN_ALL.CSV")
	//if err != nil {
	//	log.Fatal(err.Error())
	//}
	//sSJIS := string(bSJIS)
	//
	//sUTF, err := DecodeSJIS(sSJIS)
	//if err != nil {
	//	log.Fatal(err.Error())
	//}
	//err = ioutil.WriteFile("KEN_ALL_UTF8.CSV", []byte(sUTF), 0664)
	//if err != nil {
	//	log.Fatal(err.Error())
	//}

	postcodes, err := LoadStruct("KEN_ALL_UTF8.CSV")
	if err != nil {
		log.Fatal(err.Error())
	}

	err = WriteJson(postcodes, "KEN_ALL.json")
	if err != nil {
		log.Fatal(err.Error())
	}
}

func DownloadFile(url string, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fmt.Println(f.Name)
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(path, f.Mode())
			if err != nil {
				return err
			}
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			defer f.Close()
			if err != nil {
				return err
			}
			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func DecodeSJIS(s string) (string, error) {
	rInUTF8 := transform.NewReader(strings.NewReader(s), japanese.ShiftJIS.NewDecoder())
	// decode our string
	decBytes, err := ioutil.ReadAll(rInUTF8)
	if err != nil {
		return "", err
	}
	return string(decBytes), nil
}

func LoadStruct(file string) ([]model.Postcode, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := csv.NewReader(f)
	var postcodes []model.Postcode
	for {
		elements, err := reader.Read()
		if err != nil {
			break // err means EOF
		}
		p := model.Postcode{
			OrgCode:        elements[0],
			PostcodeOld:    elements[1],
			Postcode:       elements[2],
			PrefectureRuby: elements[3],
			CityRuby:       elements[4],
			AreaRuby:       elements[5],
			Prefecture:     elements[6],
			City:           elements[7],
			Area:           elements[8],
			Flag1:          elements[9],
			Flag2:          elements[10],
			Flag3:          elements[11],
			Flag4:          elements[12],
			Flag5:          elements[13],
			Flag6:          elements[14],
		}
		postcodes = append(postcodes, p)
	}
	return postcodes, nil
}

func WriteJson(postcodes []model.Postcode, file string) error {
	var buf bytes.Buffer
	buf.WriteString("[\n")
	for i, p := range postcodes {
		b, err := json.Marshal(p)
		if err != nil {
			return err
		}
		buf.WriteString("  ")
		buf.Write(b)
		if i != len(postcodes)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString("]\n")

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	w.Write(buf.Bytes())
	w.Flush()

	return nil
}
