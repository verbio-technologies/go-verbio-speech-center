package main

import (
	"github.com/jessevdk/go-flags"
	"verbio_speech_center"
	"verbio_speech_center/constants"
	"verbio_speech_center/log"
)

const (
	BACKEND_URL = "speechcenter.verbio.com:2424"
)

var opts struct {
	LogLevel  string `short:"l" long:"log-level" description:"Log Level (must be one of TRACE DEBUG INFO WARN ERROR)" default:"info"`
	TokenFile string `short:"t" long:"token-file" description:"Path to the Token File" required:"true"`
	Url       string `short:"u" long:"url" description:"Url of the service" default:""`
	Audio     string `short:"a" long:"audio" description:"Audio file to be sent" required:"true"`
	Grammar   string `short:"g" long:"grammar" description:"Path to the grammar to be used" required:"false"`
	Topic     string `short:"T" long:"topic" description:"Topic to be used" required:"false"`
	Language  string `short:"L" long:"language" description:"Language to be used" default:"en-US" required:"false"`
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Logger.Fatalf("Invalid usage: %+v", err)
	}

	log.InitLogger(opts.LogLevel)
	log.Logger.Infof("Starting %s (%s)", constants.APP_NAME, constants.VERSION)

	url := BACKEND_URL
	if opts.Url != "" {
		url = opts.Url
	}

	log.Logger.Infof("Using URL: [%s]", url)
	recogniser, err := verbio_speech_center.NewRecogniser(url, opts.TokenFile)
	if err != nil {
		log.Logger.Fatalf("Error creating recogniser: %+v", err)
	}

	defer recogniser.Close()

	var res string
	if opts.Grammar != "" {
		res, err = recogniser.RecogniseWithGrammar(opts.Audio, opts.Grammar, opts.Language)
	} else if opts.Topic != "" {
		res, err = recogniser.RecogniseWithTopic(opts.Audio, opts.Topic, opts.Language)
	} else {
		log.Logger.Fatal("Either a grammar or a topic must be specified")
	}
	if err != nil {
		log.Logger.Fatalf("Error in recognition: %+v", err)
	}

	log.Logger.Infof("Result: %s", res)
}
