package verbio_speech_center

import (
	"crypto/tls"
	"errors"
	"fmt"

	"os"
	"strings"
	"verbio_speech_center/log"
	pb "verbio_speech_center/proto/speech_center"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

type Recogniser struct {
	conn   *grpc.ClientConn
	client pb.RecognizerClient
}

func NewRecogniser(url string, tokenFile string) (*Recogniser, error) {
	token, err := loadToken(tokenFile)
	log.Logger.Infof("Loaded token from file: [%s]", tokenFile)
	if err != nil {
		return nil, err
	}

	token = strings.TrimSpace(token)
	conn, err := initConnection(url, token)
	log.Logger.Infof("Established connection to the URL: [%s]", url)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error establishing connection: %+v", err))
	}

	client := pb.NewRecognizerClient(conn)
	return &Recogniser{
		conn:   conn,
		client: client,
	}, nil
}

func (r *Recogniser) Close() error {
	return r.conn.Close()
}

func initConnection(url string, token string) (*grpc.ClientConn, error) {
	log.Logger.Debugf("Initializing connection to the URL: [%s]", url)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS13,
		})),
		grpc.WithPerRPCCredentials(oauth.TokenSource{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token, TokenType: "Bearer"})}),
	}
	conn, err := grpc.NewClient(url, opts...)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error in grpc dial: %+v", err))
	}

	return conn, nil
}

func loadToken(file string) (string, error) {
	contents, err := os.ReadFile(file)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error reading token file: %+v", err))
	}

	return string(contents), nil
}
