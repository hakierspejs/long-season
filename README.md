# long-season

The long-awaited web service for checking who is currently hacking at hackerspace (or doing anything else anywhere). `long-season` provides simple, responsive web user interface and REST API, for interacting with service without need for using browser. `long-season` is fully fledged, you don't need to setup external database or download interpreter. Everything you need is within this repository or included in docker image.

**Although long-season is now usable, it is still in early development stage.**

## How it's working?

Basically: users are adding MAC addresses of their devices to their profile and administrator of `long-season` has to check online devices at his local network (you can achieve this with `nmap`) and send their MAC addresses to specified endpoint. That's it folks. For more question check [FAQ](./docs/FAQ.md) or open new issue.

## Installing

Right now you have to build `long-season` manually. But you can facilitate your work by using a [task](https://taskfile.dev).

Clone this repository and enter below command in your shell (or whatever you are using):

    $ task

Now you have `long-season` and `long-season-cli` in your root directory. You can use `task` for running project and rebuilding it whenever some changes occurs.

    $ task --watch run

You can also build docker image or just use `docker-compose`, which is the simplest way to start development or use `long-season`.

## License

[BSD 3-Clause License](./LICENSE)
