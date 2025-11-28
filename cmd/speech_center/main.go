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

type GlobalOpts struct {
	LogLevel  string `short:"l" long:"log-level" description:"Log Level (must be one of TRACE DEBUG INFO WARN ERROR)" default:"info"`
	TokenFile string `short:"t" long:"token-file" description:"Path to the Token File"`
	Url       string `short:"u" long:"url" description:"Url of the service" default:""`
}

type RecognizeCmd struct {
	Audio    string `short:"a" long:"audio" description:"Audio file to be sent" required:"true"`
	Grammar  string `short:"g" long:"grammar" description:"Path to the grammar to be used"`
	Topic    string `short:"T" long:"topic" description:"Topic to be used"`
	Language string `short:"L" long:"language" description:"Language to be used" default:"en-US"`
}

type SynthesizeCmd struct {
	Text         string `short:"s" long:"text" description:"Text to synthesize" required:"true"`
	Voice        string `short:"v" long:"voice" description:"Voice code to use for synthesis" required:"true"`
	SamplingRate string `long:"sampling-rate" description:"Sampling rate for synthesis (8khz or 16khz)" default:"16khz"`
	Format       string `long:"format" description:"Audio format for synthesis (wav or raw)" default:"wav"`
	Output       string `short:"o" long:"output" description:"Output file for synthesized audio" required:"true"`
}

var globalOpts GlobalOpts
var parser = flags.NewParser(&globalOpts, flags.Default)

func main() {
	recognizeCmd := RecognizeCmd{}
	parser.AddCommand("recognize", "Recognize speech from audio file", "Recognize speech from an audio file using grammar or topic", &recognizeCmd)

	synthesizeCmd := SynthesizeCmd{}
	parser.AddCommand("synthesize", "Synthesize speech from text", "Synthesize speech from text to audio file", &synthesizeCmd)

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok {
			if flagsErr.Type == flags.ErrUnknownCommand {
				return
			}
		}
		log.Logger.Fatalf("Invalid usage: %+v", err)
	}

	if parser.Active == nil {
		parser.WriteHelp(nil)
		log.Logger.Fatal("No command specified. Use 'recognize' or 'synthesize'")
	}

	if globalOpts.TokenFile == "" {
		log.Logger.Fatal("Token file is required. Use -t or --token-file")
	}

	commandName := parser.Active.Name

	log.InitLogger(globalOpts.LogLevel)
	log.Logger.Infof("Starting %s (%s)", constants.APP_NAME, constants.VERSION)

	url := BACKEND_URL
	if globalOpts.Url != "" {
		url = globalOpts.Url
	}

	log.Logger.Infof("Using the URL: [%s]", url)

	switch commandName {
	case "recognize":
		executeRecognize(url, globalOpts.TokenFile, &recognizeCmd)
	case "synthesize":
		executeSynthesize(url, globalOpts.TokenFile, &synthesizeCmd)
	}
}

func executeRecognize(url, tokenFile string, cmd *RecognizeCmd) {
	recogniser, err := verbio_speech_center.NewRecogniser(url, tokenFile)
	log.Logger.Infof("Created recogniser")
	if err != nil {
		log.Logger.Fatalf("Error creating recogniser: %+v", err)
	}
	defer recogniser.Close()

	var res string
	if cmd.Grammar != "" {
		res, err = recogniser.RecogniseWithGrammar(cmd.Audio, cmd.Grammar, cmd.Language)
	} else if cmd.Topic != "" {
		res, err = recogniser.RecogniseWithTopic(cmd.Audio, cmd.Topic, cmd.Language)
	} else {
		log.Logger.Fatal("Either a grammar or a topic must be specified for recognition")
	}
	if err != nil {
		log.Logger.Fatalf("Error in recognition: %+v", err)
	}

	log.Logger.Infof("Result: %s", res)
}

func executeSynthesize(url, tokenFile string, cmd *SynthesizeCmd) {
	synthesizer, err := verbio_speech_center.NewSynthesizer(url, tokenFile)
	log.Logger.Infof("Created synthesizer")
	if err != nil {
		log.Logger.Fatalf("Error creating synthesizer: %+v", err)
	}
	defer synthesizer.Close()

	// Parse sampling rate
	var samplingRate texttospeech.VoiceSamplingRate
	switch cmd.SamplingRate {
	case "8khz", "8kHz", "8":
		samplingRate = texttospeech.VoiceSamplingRate_VOICE_SAMPLING_RATE_8KHZ
	case "16khz", "16kHz", "16":
		samplingRate = texttospeech.VoiceSamplingRate_VOICE_SAMPLING_RATE_16KHZ
	default:
		log.Logger.Fatalf("Invalid sampling rate: %s (must be 8khz or 16khz)", cmd.SamplingRate)
	}

	// Parse format
	var format texttospeech.AudioFormat
	switch cmd.Format {
	case "wav":
		format = texttospeech.AudioFormat_AUDIO_FORMAT_WAV_LPCM_S16LE
	case "raw":
		format = texttospeech.AudioFormat_AUDIO_FORMAT_RAW_LPCM_S16LE
	default:
		log.Logger.Fatalf("Invalid format: %s (must be wav or raw)", cmd.Format)
	}

	err = synthesizer.StreamingSynthesizeSpeech(cmd.Text, cmd.Voice, samplingRate, format, cmd.Output)
	if err != nil {
		log.Logger.Fatalf("Error in synthesis: %+v", err)
	}

	log.Logger.Infof("Successfully synthesized speech to %s", cmd.Output)
}
