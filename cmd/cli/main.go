package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
)

func putRequest(url string, data io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, data)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type body struct {
	Addresses []string `json:"addresses"`
}

var (
	api string
)

func run() error {
	flag.StringVar(&api, "api", "", "api address")
	flag.Parse()

	b := new(body)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		b.Addresses = append(b.Addresses, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	body, err := json.Marshal(b)
	if err != nil {
		return err
	}

	buff := bytes.NewBuffer(body)

	_, err = putRequest(api+"/api/v1/update", buff)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
