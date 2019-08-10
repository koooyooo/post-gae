package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"

	"github.com/koooyooo/post-gae/batch/model"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const (
	ZipName      = "ken_all.zip"
	RawFileName  = "KEN_ALL.CSV"
	UTF8FileName = "KEN_ALL_UTF8.CSV"
	JSONFileName = "KEN_ALL.json"
)

func main() {
	flag.Parse()
	args := flag.Args()
	bucket := args[0]
	gcsPath := args[1]

	Update(bucket, gcsPath)
}

func Update(bucket, gcsPath string) {
	fmt.Println("Start Downloading...")
	fmt.Printf("  1. downloading %s\n", ZipName)
	err := DownloadFile("https://www.post.japanpost.jp/zipcode/dl/kogaki/zip/"+ZipName, ZipName)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer os.Remove(ZipName)

	fmt.Printf("  2. unzip %s\n", ZipName)
	err = Unzip(ZipName, ".")
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("  3. convert sjis %s to utf8 %s\n", RawFileName, UTF8FileName)
	bSJIS, err := ioutil.ReadFile(RawFileName)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer os.Remove(RawFileName)
	sSJIS := string(bSJIS)

	sUTF, err := DecodeSJIS(sSJIS)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = ioutil.WriteFile(UTF8FileName, []byte(sUTF), 0664)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer os.Remove(UTF8FileName)

	fmt.Printf("  4. read data from %s\n", UTF8FileName)
	postcodes, err := LoadStruct(UTF8FileName)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("  5. write data as %s\n", JSONFileName)
	err = WriteJson(postcodes, JSONFileName)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("  6. upload data to Cloud Storage")
	err = UploadJsonToGCS(JSONFileName, bucket, gcsPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Finish Downloading.")
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

func UploadJsonToGCS(filePath, bucketName, gcsPath string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	w := client.Bucket(bucketName).Object(gcsPath).NewWriter(ctx)
	w.ContentType = "application/json"

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	written, err := io.Copy(w, f)
	fmt.Println(written)
	if err != nil {
		return err
	}
	defer w.Close()
	return nil
}
