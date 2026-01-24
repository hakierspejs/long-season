package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/urfave/cli/v3"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/bcrypt"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/exim"
	"github.com/hakierspejs/long-season/pkg/services/users"
	"github.com/hakierspejs/long-season/pkg/storage"
	"github.com/hakierspejs/long-season/pkg/storage/abstract"
	"github.com/hakierspejs/long-season/pkg/storage/memory"
)

const (
	usersBucket   = "ls::users"
	devicesBucket = "ls::devices"
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

func usersStorage(cmd *cli.Command) (storage.Users, func(), error) {
	if cmd.String("database") == "" {
		return nil, nil, fmt.Errorf("database flag is not set. see admin command.")
	}

	boltPath := cmd.String("database")
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

func app() *cli.Command {
	return &cli.Command{
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
				Action: func(ctx context.Context, cmd *cli.Command) error {
					api := cmd.String("api")
					apiKey := cmd.String("api-key")

					b := new(body)

					scanner := bufio.NewScanner(os.Stdin)

					for scanner.Scan() {
						address := scanner.Text()
						if address != "" {
							b.Addresses = append(b.Addresses, scanner.Text())
						}
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
						Usage:   "path to database",
						Value:   "long-season.db",
					},
					&cli.StringFlag{
						Name:    "database-type",
						Aliases: []string{"dt", "dbt"},
						Usage:   "type of database",
						Value:   "bolt",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return cli.DefaultShowCommandHelp(ctx, cmd, cmd.Name)
				},
				Commands: []*cli.Command{
					{
						Name:  "export",
						Usage: "export all data from database as single json object to stdout",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							dbPath := cmd.String("database")
							dbType := cmd.String("database-type")

							factory, closer, err := abstract.Factory(dbPath, dbType)
							if err != nil {
								return fmt.Errorf("abstract.Factory: %w", err)
							}
							defer closer()

							dump, err := exim.Export(ctx, exim.ExportRequest{
								UsersStorage:     factory.Users(),
								DevicesStorage:   factory.Devices(),
								TwoFactorStorage: factory.TwoFactor(),
							})
							if err != nil {
								return fmt.Errorf("exim.Export: %w", err)
							}

							if err := json.NewEncoder(os.Stdout).Encode(dump); err != nil {
								return fmt.Errorf("json.NewEncoder.Encode: %w", err)
							}

							return nil
						},
					},
					{
						Name:  "import",
						Usage: "import data as json object from stdin into database",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							dbPath := cmd.String("database")
							dbType := cmd.String("database-type")

							factory, closer, err := abstract.Factory(dbPath, dbType)
							if err != nil {
								return fmt.Errorf("abstract.Factory: %w", err)
							}
							defer closer()

							dump := exim.Data{}

							if err := json.NewDecoder(os.Stdin).Decode(&dump); err != nil {
								return fmt.Errorf("json.NewDecoder.Decode: %w", err)
							}

							err = exim.Import(ctx, exim.ImportRequest{
								Dump:             dump,
								UsersStorage:     factory.Users(),
								DevicesStorage:   factory.Devices(),
								TwoFactorStorage: factory.TwoFactor(),
							})
							if err != nil {
								return fmt.Errorf("exim.Import: %w", err)
							}

							return nil
						},
					},
					{
						Name:  "users",
						Usage: "show users stored in given database",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "user-id",
								Aliases:  []string{"id", "i"},
								Required: false,
							},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							storage, closer, err := usersStorage(cmd)
							if err != nil {
								return err
							}
							defer closer()

							users, err := storage.All(ctx)
							if err != nil {
								return err
							}

							var user *models.User = nil
							if cmd.IsSet("user-id") {
								target := cmd.String("user-id")
								for _, u := range users {
									if u.ID == target {
										user = &models.User{
											UserPublicData: models.UserPublicData{
												ID:       u.ID,
												Nickname: u.Nickname,
											},
											Password: u.HashedPassword,
											Private:  u.Private,
										}
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
						Commands: []*cli.Command{
							{
								Name:    "delete",
								Aliases: []string{"d"},
								Usage:   "delete user with given id",
								Action: func(ctx context.Context, cmd *cli.Command) error {
									if !cmd.IsSet("user-id") {
										return fmt.Errorf("set user-id flag with users subcommand")
									}

									storage, closer, err := usersStorage(cmd)
									defer closer()
									if err != nil {
										return err
									}

									return storage.Remove(ctx, cmd.String("user-id"))
								},
							},
							{
								Name:    "add",
								Aliases: []string{"a"},
								Usage:   "add new user",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:     "nickname",
										Aliases:  []string{"n"},
										Usage:    "nickname for new user",
										Required: true,
									},
									&cli.StringFlag{
										Name:     "password",
										Aliases:  []string{"p"},
										Usage:    "password for new user",
										Required: true,
									},
								},
								Action: func(ctx context.Context, cmd *cli.Command) error {
									if !cmd.IsSet("nickname") || !cmd.IsSet("password") {
										return fmt.Errorf("please set nickname and password for new user")
									}

									newNickname := cmd.String("nickname")
									newPassword := cmd.String("password")

									s, closer, err := usersStorage(cmd)
									if err != nil {
										return err
									}
									defer closer()

									hashedPassword, err := bcrypt.GenerateFromPassword(
										[]byte(newPassword), bcrypt.DefaultCost,
									)
									if err != nil {
										return err
									}

									_, err = s.New(ctx, storage.UserEntry{
										Nickname:       newNickname,
										HashedPassword: hashedPassword,
										Private:        false,
									})
									return err
								},
							},
							{
								Name:    "edit",
								Aliases: []string{"e"},
								Usage:   "edit user with given id",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:    "nickname",
										Aliases: []string{"n"},
										Usage:   "value of new nickname for user with given id",
										Value:   "",
									},
									&cli.StringFlag{
										Name:    "password",
										Aliases: []string{"p"},
										Usage:   "value of new password for user with given id",
										Value:   "",
									},
								},
								Action: func(ctx context.Context, cmd *cli.Command) error {
									if !cmd.IsSet("user-id") {
										return fmt.Errorf("set user-id flag with users subcommand")
									}

									s, closer, err := usersStorage(cmd)
									defer closer()
									if err != nil {
										return err
									}

									user, err := s.Read(ctx, cmd.String("user-id"))
									if err != nil {
										return err
									}

									newHashedPassword := []byte{}
									if newPassword := cmd.String("password"); newPassword != "" {
										newHashedPassword, err = bcrypt.GenerateFromPassword(
											[]byte(newPassword), bcrypt.DefaultCost,
										)
										if err != nil {
											return err
										}
									}

									newUser := users.Update(models.User{
										UserPublicData: models.UserPublicData{
											ID:       user.ID,
											Nickname: user.Nickname,
											Online:   false,
										},
										Password: user.HashedPassword,
										Private:  user.Private,
									}, &users.Changes{
										Nickname: cmd.String("nickname"),
										Password: newHashedPassword,
										Online:   nil,
									})

									return s.Update(ctx, newUser.ID, func(u *storage.UserEntry) error {
										u.Nickname = newUser.Nickname
										u.HashedPassword = newUser.Password
										return nil
									})
								},
							},
						},
					},
				},
			},
		},
	}
}

func main() {
	if err := app().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "short-season: %s\n", err.Error())
		os.Exit(1)
	}
}
