> WARNING: This repository is no longer updated or mantained. Please refer to our [python](https://github.com/verbio-technologies/python-verbio-speech-center) or [c++](https://github.com/verbio-technologies/cpp-verbio-speech-center) alternatives.

# Verbio Speech Center (Go client)

[![Lint](https://github.com/cquintana92/go-verbio-speech-center/actions/workflows/lint.yaml/badge.svg)](https://github.com/cquintana92/go-verbio-speech-center/actions/workflows/lint.yaml)


## How to build

You will need Go installed in your machine in order to build this client.

For generating the grpc files you can run the following command

```shell
$ make grpc

# Or if you prefer to avoid installing the grpc dependencies in
# your system, use this to use a docker container for that
$ make grpc-docker
```

Then, run the following command for building the binary:

```shell
$ make build
```

The binary will be placed at `bin/`

## How to use

```shell
# Topic recognition
$ bin/speech_center -a audios/example.wav -t your.speech-center.token.file --topic generic --language en-US

# Grammar recognition
$ bin/speech_center -a audios/example.wav -t your.speech-center.token.file --grammar path/to/grammar.bnf --language en-US
```
