package verbio_speech_center

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSynthesizer(t *testing.T) {

	tests := []struct {
		name      string
		url       string
		tokenFile string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Invalid URL",
			url:       "",
			tokenFile: createTemporaryToken(t),
			wantErr:   true,
			errMsg:    "invalid URL: URL cannot be empty",
		},
		{
			name:      "Invalid token file",
			url:       "host:8080",
			tokenFile: "nonexistent.token",
			wantErr:   true,
			errMsg:    "error reading token file:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSynthesizer(tt.url, tt.tokenFile)
			if tt.wantErr {
				assert.Error(t, err, tt.name)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, s)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, s)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Empty URL",
			url:     "",
			wantErr: true,
			errMsg:  "URL cannot be empty",
		},
		{
			name:    "Valid URL with port",
			url:     "host:8080",
			wantErr: false,
		},
		{
			name:    "Valid URL without port",
			url:     "host",
			wantErr: false,
		},
		{
			name:    "Invalid URL format",
			url:     "host:port:extra",
			wantErr: true,
			errMsg:  "URL must be in format host:port",
		},
		{
			name:    "Empty host",
			url:     ":8080",
			wantErr: true,
			errMsg:  "host cannot be empty",
		},
		{
			name:    "Empty port",
			url:     "host:",
			wantErr: true,
			errMsg:  "port cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
