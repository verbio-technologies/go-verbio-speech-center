# Verbio Speech Center (Go client)

[![Lint](https://github.com/cquintana92/go-verbio-speech-center/actions/workflows/lint.yaml/badge.svg)](https://github.com/cquintana92/go-verbio-speech-center/actions/workflows/lint.yaml)


## How to build

You will need Go installed in your machine in order to build this client.

Run the following command for building the binary:

```shell
$ make build
```

The binary will be placed at `bin/`

## How to use

```shell
# Topic recognition
$ bin/speech_center recognize -a your_audio_file.wav -t your_token.txt  --language language-id -T GENERIC --word-boosting term1 --word-boosting term2

# Audio synthesis
$ bin/speech_center synthesize -s "your string" -v voice-id -o output.wav --format wav --sampling-rate 8 -t your_token.txt

```
