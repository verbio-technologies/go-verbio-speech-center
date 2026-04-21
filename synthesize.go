package verbio_speech_center

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"verbio_speech_center/log"
	ttsv1 "verbio_speech_center/proto/speechcenter/tts"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"google.golang.org/grpc"
)

func (s *Synthesizer) getStreamingClient() error {
	var err error
	s.stream, err = s.client.StreamingSynthesizeSpeech(context.Background(), grpc.WaitForReady(true))
	if err != nil {
		return fmt.Errorf("error obtaining streaming client: %+v", err)
	}
	return nil
}

func buildPronunciationEntries(dict map[string]string) []*ttsv1.PronunciationEntry {
	if len(dict) == 0 {
		return nil
	}
	entries := make([]*ttsv1.PronunciationEntry, 0, len(dict))
	for term, ipa := range dict {
		entries = append(entries, &ttsv1.PronunciationEntry{
			Term: term,
			PronunciationFormat: &ttsv1.PronunciationEntry_Ipa{
				Ipa: ipa,
			},
		})
	}
	return entries
}

func (s *Synthesizer) sendConfig(voice string, samplingRate ttsv1.VoiceSamplingRate, pronunciationDict map[string]string) error {
	config := &ttsv1.StreamingSynthesisRequest{
		SynthesisRequest: &ttsv1.StreamingSynthesisRequest_Config{
			Config: &ttsv1.SynthesisConfig{
				Voice:                   voice,
				SamplingRate:            samplingRate,
				PronunciationDictionary: buildPronunciationEntries(pronunciationDict),
			},
		},
	}
	if err := s.stream.Send(config); err != nil {
		return fmt.Errorf("error sending config: %+v", err)
	}
	log.Logger.Debugf("Sent config")
	return nil
}

func (s *Synthesizer) sendText(text string) error {
	textReq := &ttsv1.StreamingSynthesisRequest{
		SynthesisRequest: &ttsv1.StreamingSynthesisRequest_Text{
			Text: text,
		},
	}
	if err := s.stream.Send(textReq); err != nil {
		return fmt.Errorf("error sending text: %+v", err)
	}
	log.Logger.Debugf("Sent text")
	return nil
}

func (s *Synthesizer) sendEndOfUtterance() error {
	endReq := &ttsv1.StreamingSynthesisRequest{
		SynthesisRequest: &ttsv1.StreamingSynthesisRequest_EndOfUtterance{
			EndOfUtterance: &ttsv1.EndOfUtterance{},
		},
	}
	if err := s.stream.Send(endReq); err != nil {
		return fmt.Errorf("error sending end of utterance: %+v", err)
	}
	log.Logger.Debugf("Sent end of utterance")
	return nil
}

func (s *Synthesizer) closeSend() error {
	if err := s.stream.CloseSend(); err != nil {
		return fmt.Errorf("error closing send: %+v", err)
	}
	return nil
}

type audioResult struct {
	audioData []byte
	err       error
}

func (s *Synthesizer) collectAudioChunks(c chan audioResult) chan audioResult {
	var allAudioData []byte
	log.Logger.Debugf("> Waiting for audio responses ...")
	for {
		resp, err := s.stream.Recv()
		if err == io.EOF {
			log.Logger.Debugf("Received EOF")
			break
		}
		if err != nil {
			c <- audioResult{audioData: nil, err: fmt.Errorf("error receiving audio: %+v", err)}
			return c
		}

		if audio := resp.GetStreamingAudio(); audio != nil {
			audioSamples := audio.GetAudioSamples()
			allAudioData = append(allAudioData, audioSamples...)
			log.Logger.Debugf("Received audio chunk: %d bytes", len(audioSamples))
		} else if resp.GetEndOfUtterance() != nil {
			log.Logger.Debugf("Received end of utterance")
			break
		}
	}

	if len(allAudioData) == 0 {
		c <- audioResult{audioData: nil, err: errors.New("received no audio data")}
		return c
	}

	log.Logger.Debugf("< all audio responses received")
	c <- audioResult{audioData: allAudioData, err: nil}
	return c
}

func (s *Synthesizer) StreamingSynthesizeSpeech(text string, voice string, samplingRate ttsv1.VoiceSamplingRate, format ttsv1.AudioFormat, outputFile string, pronunciationDict map[string]string) error {
	log.Logger.Infof("Streaming synthesis [text=%s] [voice=%s] [samplingRate=%v] [format=%v] [outputFile=%s] [pronunciationEntries=%d]", text, voice, samplingRate, format, outputFile, len(pronunciationDict))

	if text == "" {
		return errors.New("text cannot be empty")
	}
	if voice == "" {
		return errors.New("voice cannot be empty")
	}
	if outputFile == "" {
		return errors.New("output file cannot be empty")
	}

	if err := s.getStreamingClient(); err != nil {
		return err
	}

	c := make(chan audioResult)
	go func() {
		c = s.collectAudioChunks(c)
	}()

	if err := s.sendConfig(voice, samplingRate, pronunciationDict); err != nil {
		return err
	}

	if err := s.sendText(text); err != nil {
		return err
	}

	if err := s.sendEndOfUtterance(); err != nil {
		return err
	}

	if err := s.closeSend(); err != nil {
		return err
	}

	log.Logger.Info("Waiting for audio collection to finish")
	result := <-c
	if result.err != nil {
		return result.err
	}

	allAudioData := result.audioData

	var err error
	if format == ttsv1.AudioFormat_AUDIO_FORMAT_WAV_LPCM_S16LE {
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

func saveWavAudio(file string, pcmData []byte, samplingRate ttsv1.VoiceSamplingRate) error {
	var sampleRate int
	switch samplingRate {
	case ttsv1.VoiceSamplingRate_VOICE_SAMPLING_RATE_8KHZ:
		sampleRate = 8000
	case ttsv1.VoiceSamplingRate_VOICE_SAMPLING_RATE_16KHZ:
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

	defer func() {
		if err := outFile.Close(); err != nil {
			log.Logger.Errorf("Error closing output file: %+v", err)
		}
	}()

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
