# long-season

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/hakierspejs/long-season/Go?style=flat-square)](https://github.com/hakierspejs/long-season/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/hakierspejs/long-season?style=flat-square)](https://goreportcard.com/report/github.com/hakierspejs/long-season)
[![Docker Pulls](https://img.shields.io/docker/pulls/thinkofher/long-season?style=flat-square)](https://hub.docker.com/r/thinkofher/long-season)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/hakierspejs/long-season?style=flat-square)](https://github.com/hakierspejs/long-season/blob/main/go.mod)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/hakierspejs/long-season?style=flat-square)](https://github.com/hakierspejs/long-season/releases)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/hakierspejs/long-season)](https://pkg.go.dev/github.com/hakierspejs/long-season)

The long-awaited web service for checking who is currently hacking at hackerspace (or doing anything else anywhere). `long-season` provides simple, responsive web user interface and REST API, for interacting with service without need for using browser. `long-season` is fully fledged, you don't need to setup external database or download interpreter. Everything you need is within this repository or included in docker image.

**Although long-season is now usable, it is still in early development stage.**

## How it's working?

Basically: users are adding MAC addresses of their devices to their profile and administrator of `long-season` has to check online devices at his local network (you can achieve this with `nmap`) and send their MAC addresses to specified endpoint. That's it folks. For more question check [FAQ](./docs/FAQ.md) or open new issue.

## Installing

Right now you have to build `long-season` manually. But you can facilitate your work by using a [make.bash](./make.bash) script.

Clone this repository and enter below command in your shell (or whatever you are using):

    $ ./make.bash build

Now you have `long-season` and `short-season` in your root directory. You can use [make.bash](./make.bash) for running project and rebuilding it whenever some changes occurs.

    $ ./make.bash watch

You can also build docker image or just use `docker-compose`, which is the simplest way to start development or use `long-season`.

## License

[BSD 3-Clause License](./LICENSE)
