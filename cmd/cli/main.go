package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
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

func app() *cli.App {
	return &cli.App{
		Name:  "short-season",
		Usage: "command line interface tool for managing long-season",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "api",
				Value: "",
				Usage: "api address",
			},
			&cli.StringFlag{
				Name:  "api-key",
				Value: "",
				Usage: "api key updating statuses",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "macs",
				Usage: "upload list of mac addresses to given long-season API",
				Action: func(ctx *cli.Context) error {
					api := ctx.String("api")
					apiKey := ctx.String("api-key")

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
				},
			},
			{
				Name:  "admin",
				Usage: "set of administration tools for managing content of long-season database",
				Flags: []cli.Flag{},
				Subcommands: []*cli.Command{
					{
						Name: "users",
						Subcommands: []*cli.Command{
							{
								Name: "delete",
							},
							{
								Name: "add",
							},
							{
								Name: "edit",
							},
						},
					},
				},
			},
		},
	}
}

func main() {
	if err := app().Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
