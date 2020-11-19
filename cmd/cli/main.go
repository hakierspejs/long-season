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

func putRequest(url string, headers map[string]string, data io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, data)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
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
	api    string
	apiKey string
)

func run() error {
	flag.StringVar(&api, "api", "", "api address")
	flag.StringVar(&apiKey, "api-key", "", "api key for updating statuses")
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

	headers := map[string]string{
		"Authorization": "Status " + apiKey,
	}
	_, err = putRequest(api+"/api/v1/update", headers, buff)
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
