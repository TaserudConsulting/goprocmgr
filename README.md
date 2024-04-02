[![Check](https://github.com/TaserudConsulting/goprocmgr/actions/workflows/check.yml/badge.svg)](https://github.com/TaserudConsulting/goprocmgr/actions/workflows/check.yml)
[![Update](https://github.com/TaserudConsulting/goprocmgr/actions/workflows/update.yml/badge.svg)](https://github.com/TaserudConsulting/goprocmgr/actions/workflows/update.yml)

# goprocmgr
This program is a configuration manager and process runner for servers, it
has an http API to manage and retrieve the configuration. It also provides a
CLI client for interacting with the API.

It's inspired by [Chalet](https://github.com/jeansaad/chalet) which is a fork
from [Hotel](https://github.com/typicode/hotel). However, this aims to be way
simpler in design, feature set and implementation.

## Features
- Remember configured "servers" by storing certain environment variables, directory and command to run to start it.
- Start, stop and read logs from the different servers.
- Simple http API to interact with the servers.
- Command line tool to interact with the API.
- Web UI to interact with the API.
- Random port assignment for servers with the environment variable `PORT`.

![Screenshot](./docs/screenshot.png)

## TODO
- Implement `direnv` support `direnv exec $dirname $command`.
- Implement "pause" function in the web interface.
- Implement keybind support in the web interface.
- Implement an overview of the keybind in the web interface.
- Implement a getting started overview in the web interface on the frontpage.
