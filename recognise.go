package verbio_speech_center

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"os"
	"strings"
	"verbio_speech_center/log"
	"verbio_speech_center/proto/speech_center"
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

func (r *Recogniser) performRecognition(audioFile string, initial *speech_center.RecognitionRequest) (string, error) {
	audio, err := loadAudio(audioFile)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error loading audio file %+v", err))
	}

	streamClient, err := r.client.RecognizeStream(context.Background(), grpc.WaitForReady(true))
	if err != nil {
		return "", errors.New(fmt.Sprintf("error obtaining streaming client: %+v", err))
	}

	c := make(chan recogResult)
	go func() {
		recog := make([]string, 0)
		for {
			resp := &speech_center.RecognitionResponse{}
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
				log.Logger.Debugf("Got partial recog: %s", resp.Text)
				recog = append(recog, resp.Text)
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
	audioRequest := &speech_center.RecognitionRequest{
		RequestUnion: &speech_center.RecognitionRequest_Audio{
			Audio: audio,
		},
	}
	err = streamClient.Send(audioRequest)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error sending audio request: %+v", err))
	}
	log.Logger.Debug("Sent audio request")

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

func generateGrammarRequest(grammar string, language string) *speech_center.RecognitionRequest {
	resource := &speech_center.RecognitionResource{
		Resource: &speech_center.RecognitionResource_InlineGrammar{
			InlineGrammar: grammar,
		},
	}

	return &speech_center.RecognitionRequest{
		RequestUnion: &speech_center.RecognitionRequest_Init{
			Init: &speech_center.RecognitionInit{
				Parameters: &speech_center.RecognitionParameters{
					Language: language,
				},
				Resource: resource,
			},
		},
	}
}

func generateTopicRequest(topic string, language string) (*speech_center.RecognitionRequest, error) {
	var model speech_center.RecognitionResource_Model
	topicLower := strings.ToLower(topic)
	if topicLower == "generic" {
		model = speech_center.RecognitionResource_GENERIC
	} else if topicLower == "banking" {
		model = speech_center.RecognitionResource_BANKING
	} else if topicLower == "telco" {
		model = speech_center.RecognitionResource_TELCO
	} else {
		return nil, errors.New(fmt.Sprintf("unrecognized topic: %s", topic))
	}

	log.Logger.Infof("Performing recognition with topic: %s", topicLower)
	resource := &speech_center.RecognitionResource{
		Resource: &speech_center.RecognitionResource_Model_{
			Model: model,
		},
	}

	return &speech_center.RecognitionRequest{
		RequestUnion: &speech_center.RecognitionRequest_Init{
			Init: &speech_center.RecognitionInit{
				Parameters: &speech_center.RecognitionParameters{
					Language: language,
				},
				Resource: resource,
			},
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
