package main

import (
	"verbio_speech_center"
	"verbio_speech_center/constants"
	"verbio_speech_center/log"
	"verbio_speech_center/proto/texttospeech"

	"github.com/jessevdk/go-flags"
)

const (
	BACKEND_URL = "us.speechcenter.verbio.com"
)

var opts struct {
	LogLevel  string `short:"l" long:"log-level" description:"Log Level (must be one of TRACE DEBUG INFO WARN ERROR)" default:"info"`
	TokenFile string `short:"t" long:"token-file" description:"Path to the Token File" required:"true"`
	Url       string `short:"u" long:"url" description:"Url of the service" default:""`
	// Recognition options
	Audio    string `short:"a" long:"audio" description:"Audio file to be sent (required for recognition)" required:"false"`
	Grammar  string `short:"g" long:"grammar" description:"Path to the grammar to be used" required:"false"`
	Topic    string `short:"T" long:"topic" description:"Topic to be used" required:"false"`
	Language string `short:"L" long:"language" description:"Language to be used" default:"en-US" required:"false"`
	// Synthesis options
	Text         string `short:"s" long:"text" description:"Text to synthesize (required for synthesis)" required:"false"`
	Voice        string `short:"v" long:"voice" description:"Voice code to use for synthesis" required:"false"`
	SamplingRate string `long:"sampling-rate" description:"Sampling rate for synthesis (8khz or 16khz)" default:"16khz" required:"false"`
	Format       string `long:"format" description:"Audio format for synthesis (wav or raw)" default:"wav" required:"false"`
	Output       string `short:"o" long:"output" description:"Output file for synthesized audio" required:"false"`
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

	log.Logger.Infof("Using the URL: [%s]", url)

	// Determine if we're doing recognition or synthesis
	isRecognition := opts.Audio != "" && (opts.Grammar != "" || opts.Topic != "")
	isSynthesis := opts.Text != "" && opts.Voice != "" && opts.Output != ""

	if !isRecognition && !isSynthesis {
		log.Logger.Fatal("Either recognition (--audio with --grammar or --topic) or synthesis (--text, --voice, --output) must be specified")
	}

	if isRecognition {
		recogniser, err := verbio_speech_center.NewRecogniser(url, opts.TokenFile)
		log.Logger.Infof("Created recogniser")
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
			log.Logger.Fatal("Either a grammar or a topic must be specified for recognition")
		}
		if err != nil {
			log.Logger.Fatalf("Error in recognition: %+v", err)
		}

		log.Logger.Infof("Result: %s", res)
	}

	if isSynthesis {
		synthesizer, err := verbio_speech_center.NewSynthesizer(url, opts.TokenFile)
		log.Logger.Infof("Created synthesizer")
		if err != nil {
			log.Logger.Fatalf("Error creating synthesizer: %+v", err)
		}
		defer synthesizer.Close()

		// Parse sampling rate
		var samplingRate texttospeech.VoiceSamplingRate
		switch opts.SamplingRate {
		case "8khz", "8kHz", "8":
			samplingRate = texttospeech.VoiceSamplingRate_VOICE_SAMPLING_RATE_8KHZ
		case "16khz", "16kHz", "16":
			samplingRate = texttospeech.VoiceSamplingRate_VOICE_SAMPLING_RATE_16KHZ
		default:
			log.Logger.Fatalf("Invalid sampling rate: %s (must be 8khz or 16khz)", opts.SamplingRate)
		}

		// Parse format
		var format texttospeech.AudioFormat
		switch opts.Format {
		case "wav":
			format = texttospeech.AudioFormat_AUDIO_FORMAT_WAV_LPCM_S16LE
		case "raw":
			format = texttospeech.AudioFormat_AUDIO_FORMAT_RAW_LPCM_S16LE
		default:
			log.Logger.Fatalf("Invalid format: %s (must be wav or raw)", opts.Format)
		}

		err = synthesizer.SynthesizeSpeech(opts.Text, opts.Voice, samplingRate, format, opts.Output)
		if err != nil {
			log.Logger.Fatalf("Error in synthesis: %+v", err)
		}

		log.Logger.Infof("Successfully synthesized speech to %s", opts.Output)
	}
}
