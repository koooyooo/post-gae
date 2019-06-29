package main

import (
	"archive/zip"
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("Hello Download")
	err := DownloadFile("https://www.post.japanpost.jp/zipcode/dl/kogaki/zip/ken_all.zip", "ken_all.zip")
	if err != nil {
		log.Fatal(err.Error())
	}

	err = Unzip("ken_all.zip", ".")
	if err != nil {
		log.Fatal(err.Error())
	}

	bSJIS, err := ioutil.ReadFile("KEN_ALL.CSV")
	if err != nil {
		log.Fatal(err.Error())
	}
	sSJIS := string(bSJIS)

	sUTF, err := DecodeSJIS(sSJIS)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = ioutil.WriteFile("KEN_ALL_UT8.CSV", []byte(sUTF), 0664)

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
			os.MkdirAll(path, f.Mode())
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