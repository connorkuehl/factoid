package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	var (
		addr string
		file string
		auth string
	)

	flag.StringVar(&addr, "addr", "http://localhost:8080/v1", "REST API addr")
	flag.StringVar(&file, "csv", "", "path to CSV")
	flag.StringVar(&auth, "auth", "", "authorization secret")
	flag.Parse()

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	pairs, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	if len(pairs) == 0 {
		log.Fatal("nothing to import")
	}

	for _, pair := range pairs {
		content, source := pair[0], pair[1]

		var obj struct {
			Content string `json:"content"`
			Source  string `json:"source"`
		}
		obj.Content = content
		obj.Source = source

		blob, err := json.Marshal(obj)
		if err != nil {
			log.Println("failed to marshal", err)
			continue
		}

		url := addr + "/facts"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(blob))
		if err != nil {
			log.Println("Create request failed", err)
			continue
		}

		req.Header.Set("Authorization", auth)
		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("failed to POST", err)
			continue
		}
		defer rsp.Body.Close()

		rspblob, _ := io.ReadAll(rsp.Body)

		log.Print(string(rspblob))
	}
}
