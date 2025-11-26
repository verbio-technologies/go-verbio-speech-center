package verbio_speech_center

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
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

		initial := generateGrammarRequest(grammar, language)
		return r.performRecognition(audioFile, initial)

	} else {
		return "", errors.New("received an empty grammarFile path")
	}
}

func (r *Recogniser) RecogniseWithTopic(audioFile string, topic string, language string) (string, error) {
	log.Logger.Infof("Performing Topic recognition [audioFile=%s] [topic=%s] [language=%s]", audioFile, topic, language)
	initial, err := generateTopicRequest(topic, language)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error creating topic request: %+v", err))
	}
	return r.performRecognition(audioFile, initial)
}

type recogResult struct {
	recognition string
	err         error
}

func (r *Recogniser) performRecognition(audioFile string, initial *speech_center.RecognitionStreamingRequest) (string, error) {
	audio, err := loadAudio(audioFile)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error loading audio file %+v", err))
	}

	streamClient, err := r.client.StreamingRecognize(context.Background(), grpc.WaitForReady(true))
	if err != nil {
		return "", errors.New(fmt.Sprintf("error obtaining streaming client: %+v", err))
	}

	c := make(chan recogResult)
	go func() {
		recog := make([]string, 0)
		for {
			resp := &speech_center.RecognitionStreamingResponse{}
			log.Logger.Debugf("Waiting for response")
			err := streamClient.RecvMsg(resp)
			if err != nil {
				if err == io.EOF {
					log.Logger.Debugf("Got EOF")
					c <- recogResult{recognition: strings.Join(recog, " "), err: nil}
					break
				} else {
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
					transcript := result.Alternatives[0].Transcript
					log.Logger.Debugf("Got partial recog: %s (is_final: %v)", transcript, result.IsFinal)
					if result.IsFinal {
						recog = append(recog, transcript)
					}
				}
			}
		}
	}()

	log.Logger.Info("Sending initial request")
	err = streamClient.Send(initial)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error sending initial request: %+v", err))
	}
	log.Logger.Debug("Sent initial request")

	log.Logger.Info("Sending audio request")
	audioRequest := &speech_center.RecognitionStreamingRequest{
		RecognitionRequest: &speech_center.RecognitionStreamingRequest_Audio{
			Audio: audio,
		},
	}
	err = streamClient.Send(audioRequest)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error sending audio request: %+v", err))
	}
	log.Logger.Debug("Sent audio request")

	// Send END_OF_STREAM event
	log.Logger.Info("Sending END_OF_STREAM event")
	endOfStreamRequest := &speech_center.RecognitionStreamingRequest{
		RecognitionRequest: &speech_center.RecognitionStreamingRequest_EventMessage{
			EventMessage: &speech_center.EventMessage{
				Event: speech_center.EventMessage_END_OF_STREAM,
			},
		},
	}
	err = streamClient.Send(endOfStreamRequest)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error sending END_OF_STREAM event: %+v", err))
	}
	log.Logger.Debug("Sent END_OF_STREAM event")

	err = streamClient.CloseSend()
	if err != nil {
		return "", errors.New(fmt.Sprintf("error closing send: %+v", err))
	}

	log.Logger.Info("Waiting for recognition result")
	recog := <-c
	if recog.err != nil {
		return "", errors.New(fmt.Sprintf("got error during recognition: %+v", recog.err))
	}

	return recog.recognition, nil
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
		Version:  speech_center.RecognitionConfig_V1,
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
		Version:  speech_center.RecognitionConfig_V1,
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
