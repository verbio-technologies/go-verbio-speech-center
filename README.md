# Verbio Speech Center (Go client)

[![Lint](https://github.com/cquintana92/go-verbio-speech-center/actions/workflows/lint.yaml/badge.svg)](https://github.com/cquintana92/go-verbio-speech-center/actions/workflows/lint.yaml)


## How to build

You will need Go installed in your machine in order to build this client.

You will also need the protobuffer compiler with the go plugin. You can install using:
```shell
$sudo apt install -y protobuf-compiler 
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

For generating the grpc files you can run the following command

```shell
$ make grpc
```

```shell
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
$ bin/speech_center recognize -a your_audio_file.wav -t your_token.txt  --language language-id -T GENERIC

# Audio synthesis
$ bin/speech_center synthesize -s "your string" -v voice-id -o output.wav --format wav --sampling-rate 8 -t your_token.txt

```
