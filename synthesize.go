package verbio_speech_center

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"verbio_speech_center/log"
	"verbio_speech_center/proto/texttospeech"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"google.golang.org/grpc"
)

func (s *Synthesizer) StreamingSynthesizeSpeech(text string, voice string, samplingRate texttospeech.VoiceSamplingRate, format texttospeech.AudioFormat, outputFile string) error {
	log.Logger.Infof("Streaming synthesis [text=%s] [voice=%s] [samplingRate=%v] [format=%v] [outputFile=%s]", text, voice, samplingRate, format, outputFile)

	if text == "" {
		return errors.New("text cannot be empty")
	}
	if voice == "" {
		return errors.New("voice cannot be empty")
	}
	if outputFile == "" {
		return errors.New("output file cannot be empty")
	}

	// Get streaming client
	stream, err := s.client.StreamingSynthesizeSpeech(context.Background(), grpc.WaitForReady(true))
	if err != nil {
		return fmt.Errorf("error obtaining streaming client: %+v", err)
	}

	// Send config first
	config := &texttospeech.StreamingSynthesisRequest{
		SynthesisRequest: &texttospeech.StreamingSynthesisRequest_Config{
			Config: &texttospeech.SynthesisConfig{
				Voice:        voice,
				SamplingRate: samplingRate,
			},
		},
	}
	if err := stream.Send(config); err != nil {
		return fmt.Errorf("error sending config: %+v", err)
	}
	log.Logger.Debugf("Sent config")

	// Send text
	textReq := &texttospeech.StreamingSynthesisRequest{
		SynthesisRequest: &texttospeech.StreamingSynthesisRequest_Text{
			Text: text,
		},
	}
	if err := stream.Send(textReq); err != nil {
		return fmt.Errorf("error sending text: %+v", err)
	}
	log.Logger.Debugf("Sent text")

	// Send end of utterance
	endReq := &texttospeech.StreamingSynthesisRequest{
		SynthesisRequest: &texttospeech.StreamingSynthesisRequest_EndOfUtterance{
			EndOfUtterance: &texttospeech.EndOfUtterance{},
		},
	}
	if err := stream.Send(endReq); err != nil {
		return fmt.Errorf("error sending end of utterance: %+v", err)
	}
	log.Logger.Debugf("Sent end of utterance")

	if err := stream.CloseSend(); err != nil {
		return fmt.Errorf("error closing send: %+v", err)
	}

	// Collect all audio chunks
	var allAudioData []byte
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Logger.Debugf("Received EOF")
			break
		}
		if err != nil {
			return fmt.Errorf("error receiving audio: %+v", err)
		}

		// Check response type
		if audio := resp.GetStreamingAudio(); audio != nil {
			audioSamples := audio.GetAudioSamples()
			allAudioData = append(allAudioData, audioSamples...)
			log.Logger.Debugf("Received audio chunk: %d bytes", len(audioSamples))
		} else if resp.GetEndOfUtterance() != nil {
			// End of stream
			log.Logger.Debugf("Received end of utterance")
			break
		}
	}

	if len(allAudioData) == 0 {
		return errors.New("received no audio data")
	}

	// Save audio in the requested format
	if format == texttospeech.AudioFormat_AUDIO_FORMAT_WAV_LPCM_S16LE {
		err = saveWavAudio(outputFile, allAudioData, samplingRate)
	} else {
		err = saveRawAudio(outputFile, allAudioData)
	}
	if err != nil {
		return fmt.Errorf("error saving audio file: %+v", err)
	}

	log.Logger.Infof("Successfully saved %d bytes of audio to %s", len(allAudioData), outputFile)
	return nil
}

func saveRawAudio(file string, pcmData []byte) error {
	err := os.WriteFile(file, pcmData, 0644)
	if err != nil {
		return fmt.Errorf("error writing raw audio file: %+v", err)
	}
	return nil
}

func saveWavAudio(file string, pcmData []byte, samplingRate texttospeech.VoiceSamplingRate) error {
	var sampleRate int
	switch samplingRate {
	case texttospeech.VoiceSamplingRate_VOICE_SAMPLING_RATE_8KHZ:
		sampleRate = 8000
	case texttospeech.VoiceSamplingRate_VOICE_SAMPLING_RATE_16KHZ:
		sampleRate = 16000
	default:
		sampleRate = 16000
	}

	// Convert raw PCM bytes to int16 samples
	// PCM data is 16-bit signed little-endian
	numSamples := len(pcmData) / 2
	samples := make([]int, numSamples)
	for i := 0; i < numSamples; i++ {
		samples[i] = int(int16(binary.LittleEndian.Uint16(pcmData[i*2 : i*2+2])))
	}

	audioBuffer := &audio.IntBuffer{
		Data: samples,
		Format: &audio.Format{
			NumChannels: 1, // Mono
			SampleRate:  sampleRate,
		},
	}

	outFile, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("error creating WAV file: %+v", err)
	}
	defer outFile.Close()

	format := 1
	bitDepth := 16
	numChannels := 1
	enc := wav.NewEncoder(outFile, sampleRate, bitDepth, numChannels, format)
	if err := enc.Write(audioBuffer); err != nil {
		return fmt.Errorf("error encoding WAV: %+v", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("error closing WAV encoder: %+v", err)
	}

	return nil
}
