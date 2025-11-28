package verbio_speech_center

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRecogniser(t *testing.T) {
	recognizer, err := NewRecogniser("localhost:50051", createTemporaryToken(t))
	assert.NoError(t, err)
	assert.NotNil(t, recognizer)
	assert.NotNil(t, recognizer.conn)
	assert.NotNil(t, recognizer.client)
	assert.Nil(t, recognizer.streamClient)

	err = recognizer.Close()
	assert.NoError(t, err)
}

func createTemporaryToken(t *testing.T) string {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.txt")
	err := os.WriteFile(tokenFile, []byte("test-token"), 0600)
	assert.NoError(t, err)
	return tokenFile
}

func TestNewRecogniserErrors(t *testing.T) {
	recognizer, err := NewRecogniser("localhost", "non-existent-file")
	assert.Error(t, err)
	assert.Nil(t, recognizer)

	recognizer, err = NewRecogniser("invalid-url:", createTemporaryToken(t))
	assert.Error(t, err)
	assert.Nil(t, recognizer)

}

func TestNonExistentToken(t *testing.T) {
	recognizer, err := NewRecogniser("localhost:50051", "non-existent-file")
	assert.Error(t, err)
	assert.Nil(t, recognizer)
}

func TestEmptyToken(t *testing.T) {
	recognizer, err := NewRecogniser("invalid-url", "")
	assert.Error(t, err)
	assert.Nil(t, recognizer)
}

func TestLoadToken(t *testing.T) {
	token, err := loadToken(createTemporaryToken(t))
	assert.NoError(t, err)
	assert.Equal(t, "test-token", token)
}

func TestNotExistentFile(t *testing.T) {
	token, err := loadToken("non-existent-file")
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestClose(t *testing.T) {
	recognizer, err := NewRecogniser("localhost:50051", createTemporaryToken(t))
	assert.NoError(t, err)

	err = recognizer.Close()
	assert.NoError(t, err)
}
