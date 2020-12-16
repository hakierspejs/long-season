package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/storage/memory"

	"github.com/urfave/cli/v2"
	bolt "go.etcd.io/bbolt"
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

func usersStorage(boltPath string) (*memory.UsersStorage, func(), error) {
	boltDB, err := bolt.Open(boltPath, 0666, nil)
	if err != nil {
		return nil, nil, err
	}
	closer := func() {
		boltDB.Close()
	}

	factoryStorage, err := memory.New(boltDB)
	if err != nil {
		return nil, nil, err
	}

	return factoryStorage.Users(), closer, nil
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "database",
						Aliases: []string{"d", "db"},
						Usage:   "path to bolt database",
						Value:   "long-season.db",
					},
				},
				Action: func(ctx *cli.Context) error {
					return cli.ShowCommandHelp(ctx, ctx.Command.Name)
				},
				Subcommands: []*cli.Command{
					{
						Name:  "users",
						Usage: "show users stored in given database",
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name:       "user-id",
								Aliases:    []string{"id", "i"},
								HasBeenSet: false,
								Required:   false,
							},
						},
						Action: func(ctx *cli.Context) error {
							dbPath := ctx.String("database")

							storage, closer, err := usersStorage(dbPath)
							defer closer()
							if err != nil {
								return err
							}

							users, err := storage.All(ctx.Context)
							if err != nil {
								return err
							}

							var user *models.User = nil
							if ctx.IsSet("user-id") {
								target := ctx.Int("user-id")
								for _, u := range users {
									if u.ID == target {
										user = new(models.User)
										*user = u
									}
								}

								if user == nil {
									return cli.Exit("there is no user with given id", 1)
								}
							}

							if user != nil {
								return json.NewEncoder(os.Stdout).Encode(user)
							}

							return json.NewEncoder(os.Stdout).Encode(&users)
						},
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
