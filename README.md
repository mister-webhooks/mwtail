# Mister Webhooks' `mwcat`

The `mwcat` command allows you to connect to a Mister Webhooks topic and stream its contents to the terminal.

## Usage

```
mwcat <topic> <connection_profile>
```

As you'd expect, `topic` is the the name of the topic to consume from (it'll be of the form `incoming.project.endpoint`), and `connection_profile` is a path to a connection profile for a consumer.


## Installation

### MacOS (homebrew)

```
$ brew install mister-webhooks/tools/mwcat
```
