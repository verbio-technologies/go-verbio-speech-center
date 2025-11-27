package verbio_speech_center

import (
	"fmt"
	"strings"
	"verbio_speech_center/log"
	pb "verbio_speech_center/proto/texttospeech"

	"google.golang.org/grpc"
)

type Synthesizer struct {
	conn   *grpc.ClientConn
	client pb.TextToSpeechClient
}

func NewSynthesizer(url string, tokenFile string) (*Synthesizer, error) {
	token, err := loadToken(tokenFile)
	log.Logger.Infof("Loaded token from file: [%s]", tokenFile)
	if err != nil {
		return nil, err
	}

	token = strings.TrimSpace(token)
	conn, err := initConnection(url, token)
	log.Logger.Infof("Established connection to the URL: [%s]", url)
	if err != nil {
		return nil, fmt.Errorf("error establishing connection: %+v", err)
	}

	client := pb.NewTextToSpeechClient(conn)
	return &Synthesizer{
		conn:   conn,
		client: client,
	}, nil
}

func (s *Synthesizer) Close() error {
	return s.conn.Close()
}
