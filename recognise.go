package verbio_speech_center

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"verbio_speech_center/log"
	"verbio_speech_center/proto/speech_center"

	"google.golang.org/grpc"
)

func (r *Recogniser) RecogniseWithGrammar(audioFile string, grammarFile string, language string) (string, error) {
	log.Logger.Infof("Performing Grammar recognition [audioFile=%s] [grammarFile=%s] [language=%s]", audioFile, grammarFile, language)

	if grammarFile != "" {
		grammar, err := loadGrammar(grammarFile)
		if err != nil {
			return "", errors.New(fmt.Sprintf("error loading grammar: %+v", err))
		}

		configuration := generateGrammarRequest(grammar, language)
		return r.performRecognition(audioFile, configuration)

	} else {
		return "", errors.New("received an empty grammarFile path")
	}
}

func (r *Recogniser) RecogniseWithTopic(audioFile string, topic string, language string) (string, error) {
	log.Logger.Infof("Performing Topic recognition [audioFile=%s] [topic=%s] [language=%s]", audioFile, topic, language)
	configuration, err := generateTopicRequest(topic, language)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error creating topic request: %+v", err))
	}
	return r.performRecognition(audioFile, configuration)
}

type recogResult struct {
	recognition string
	err         error
}

func (r *Recogniser) performRecognition(audioFile string, configuration *speech_center.RecognitionStreamingRequest) (string, error) {
	audio, err := loadAudio(audioFile)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error loading audio file %+v", err))
	}

	r.streamClient, err = r.client.StreamingRecognize(context.Background(), grpc.WaitForReady(true))
	if err != nil {
		return "", errors.New(fmt.Sprintf("error obtaining streaming client: %+v", err))
	}

	c := make(chan recogResult)
	go func() {
		c = r.collectResponses(c)
	}()

	if err = r.sendAudio(err, configuration, audio); err != nil {
		return "", err
	}

	log.Logger.Info("Waiting for recognition to finish")
	recog := <-c
	if recog.err != nil {
		return "", errors.New(fmt.Sprintf("got error during recognition: %+v", recog.err))
	}

	return recog.recognition, nil
}

func (r *Recogniser) collectResponses(c chan recogResult) chan recogResult {
	recog := make([]string, 0)
	log.Logger.Debugf("> Waiting for responses ...")
	totalAudioLengthInMs := float32(0)
	for {
		resp := &speech_center.RecognitionStreamingResponse{}
		err := r.streamClient.RecvMsg(resp)
		if err != nil {
			if err == io.EOF {
				log.Logger.Debugf("Got EOF")
				c <- recogResult{recognition: strings.Join(recog, " "), err: nil}
				break
			} else {
				log.Logger.Debugf("Got result")
				c <- recogResult{recognition: "", err: err}
				break
			}
		} else {
			// Check for errors in response
			if resp.GetError() != nil {
				errMsg := fmt.Sprintf("recognition error: %s (domain: %s)", resp.GetError().Reason, resp.GetError().Domain)
				c <- recogResult{recognition: "", err: errors.New(errMsg)}
				break
			}
			// Extract transcript from result
			if result := resp.GetResult(); result != nil && len(result.Alternatives) > 0 {
				log.Logger.Debugf("Got partial recog: %s (is_final: %v) (silence: %d ms)",
					result.Alternatives[0].Transcript, result.IsFinal, r.calculateEndOfUtteranceSilence(result, totalAudioLengthInMs))
				if result.IsFinal {
					recog = append(recog, result.Alternatives[0].Transcript)
					totalAudioLengthInMs += result.Duration
				}
			}
		}
	}
	log.Logger.Debugf("< all responses recevied")
	return c
}

func (r *Recogniser) calculateEndOfUtteranceSilence(result *speech_center.RecognitionResult, totalAudioLengthInMs float32) int32 {
	words := result.Alternatives[0].Words
	finalSilenceInMs := int32(0)
	if len(words) > 0 {
		finalSilenceInMs = int32((totalAudioLengthInMs + result.Duration - words[len(words)-1].EndTime) * 1000)
	}
	return finalSilenceInMs
}

func (r *Recogniser) sendAudio(err error, configuration *speech_center.RecognitionStreamingRequest, audio []byte) error {
	log.Logger.Info("Sending configuration request")
	if err = r.streamClient.Send(configuration); err != nil {
		return errors.New(fmt.Sprintf("error sending configuration request: %+v", err))
	}

	if err = r.sendAudioStream(audio); err != nil {
		return err
	}

	if err = r.streamClient.CloseSend(); err != nil {
		return errors.New(fmt.Sprintf("error closing send: %+v", err))
	}
	return nil
}

func (r *Recogniser) sendAudioStream(audio []byte) error {
	log.Logger.Info("Sending audio stream.")
	if err := r.sendAudioChunks(audio); err != nil {
		return errors.New(fmt.Sprintf("error sending Audio chunks: %+v", err))
	}
	if err := r.sendEndOfStream(); err != nil {
		return errors.New(fmt.Sprintf("error sending END_OF_STREAM event: %+v", err))
	}
	return nil
}

func (r *Recogniser) sendAudioChunks(audio []byte) error {
	const chunkSize = 800
	for i := 0; i < len(audio); i += chunkSize {
		end := i + chunkSize
		if end > len(audio) {
			end = len(audio)
		}

		audioChunk := audio[i:end]
		err := r.SendAudioRequest(audioChunk)
		if err != nil {
			return errors.New(fmt.Sprintf("error sending audio chunk: %+v", err))
		}
	}
	return nil
}

func (r *Recogniser) sendEndOfStream() error {
	// Send END_OF_STREAM event
	log.Logger.Info("Sending END_OF_STREAM event")
	endOfStreamRequest := &speech_center.RecognitionStreamingRequest{
		RecognitionRequest: &speech_center.RecognitionStreamingRequest_EventMessage{
			EventMessage: &speech_center.EventMessage{
				Event: speech_center.EventMessage_END_OF_STREAM,
			},
		},
	}
	return r.streamClient.Send(endOfStreamRequest)
}

func (r *Recogniser) SendAudioRequest(audioChunk []byte) error {
	log.Logger.Tracef("Sending audio chunk (size: %d bytes)", len(audioChunk))
	const sampleRate = int32(8000)
	endOfRequest := time.Now().Add(time.Duration(float64(len(audioChunk)) / float64(sampleRate) * float64(time.Second)))
	audioRequest := &speech_center.RecognitionStreamingRequest{
		RecognitionRequest: &speech_center.RecognitionStreamingRequest_Audio{
			Audio: audioChunk,
		},
	}

	log.Logger.Tracef("Audio chunk will be sent until %d", time.Until(endOfRequest).Milliseconds())
	time.Sleep(time.Until(endOfRequest))
	return r.streamClient.Send(audioRequest)
}

func generateGrammarRequest(grammar string, language string) *speech_center.RecognitionStreamingRequest {
	sampleRate := uint32(8000)

	resource := &speech_center.RecognitionResource{
		Resource: &speech_center.RecognitionResource_Grammar{
			Grammar: &speech_center.GrammarResource{
				Grammar: &speech_center.GrammarResource_InlineGrammar{
					InlineGrammar: grammar,
				},
			},
		},
	}

	config := &speech_center.RecognitionConfig{
		Parameters: &speech_center.RecognitionParameters{
			Language: language,
			AudioEncoding: &speech_center.RecognitionParameters_Pcm{
				Pcm: &speech_center.PCM{
					SampleRateHz: sampleRate,
				},
			},
		},
		Resource: resource,
		Version:  speech_center.RecognitionConfig_V2,
	}

	return &speech_center.RecognitionStreamingRequest{
		RecognitionRequest: &speech_center.RecognitionStreamingRequest_Config{
			Config: config,
		},
	}
}

func generateTopicRequest(topic string, language string) (*speech_center.RecognitionStreamingRequest, error) {
	topicLower := strings.ToLower(topic)
	if topicLower != "generic" {
		return nil, errors.New(fmt.Sprintf("unrecognized topic: %s (only 'generic' is supported)", topic))
	}

	// Default sample rate for speech recognition (16kHz is common)
	sampleRate := uint32(8000)

	log.Logger.Infof("Performing recognition with topic: %s", topicLower)
	resource := &speech_center.RecognitionResource{
		Resource: &speech_center.RecognitionResource_Topic_{
			Topic: speech_center.RecognitionResource_GENERIC,
		},
	}

	config := &speech_center.RecognitionConfig{
		Parameters: &speech_center.RecognitionParameters{
			Language: language,
			AudioEncoding: &speech_center.RecognitionParameters_Pcm{
				Pcm: &speech_center.PCM{
					SampleRateHz: sampleRate,
				},
			},
		},
		Resource: resource,
		Version:  speech_center.RecognitionConfig_V2,
	}

	return &speech_center.RecognitionStreamingRequest{
		RecognitionRequest: &speech_center.RecognitionStreamingRequest_Config{
			Config: config,
		},
	}, nil
}

func loadAudio(file string) ([]byte, error) {
	contents, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error reading audio file: %+v", err))
	}

	return contents, nil
}

func loadGrammar(file string) (string, error) {
	contents, err := os.ReadFile(file)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error reading grammar file: %+v", err))
	}

	return string(contents), nil
}
