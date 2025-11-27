package verbio_speech_center

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRecogniser(t *testing.T) {
	// Create temporary token file
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.txt")
	err := os.WriteFile(tokenFile, []byte("test-token"), 0600)
	assert.NoError(t, err)

	recognizer, err := NewRecogniser("localhost:50051", tokenFile)
	assert.NoError(t, err)
	assert.NotNil(t, recognizer)
	assert.NotNil(t, recognizer.conn)
	assert.NotNil(t, recognizer.client)
	assert.Nil(t, recognizer.streamClient)

	err = recognizer.Close()
	assert.NoError(t, err)
}

func TestNewRecogniserErrors(t *testing.T) {
	// Test with non-existent token file
	recognizer, err := NewRecogniser("localhost:50051", "non-existent-file")
	assert.Error(t, err)
	assert.Nil(t, recognizer)

	// Test with invalid URL
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.txt")
	err = os.WriteFile(tokenFile, []byte("test-token"), 0600)
	assert.NoError(t, err)

	recognizer, err = NewRecogniser("invalid-url", tokenFile)
	assert.Error(t, err)
	assert.Nil(t, recognizer)

}

func TestEmptyToken(t *testing.T) {
	recognizer, err := NewRecogniser("invalid-url", "")
	assert.Error(t, err)
	assert.Nil(t, recognizer)
}

func TestLoadToken(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.txt")
	expectedToken := "test-token"
	err := os.WriteFile(tokenFile, []byte(expectedToken), 0600)
	assert.NoError(t, err)

	token, err := loadToken(tokenFile)
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
}

func TestNotExistentFile(t *testing.T) {
	token, err := loadToken("non-existent-file")
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.txt")
	err := os.WriteFile(tokenFile, []byte("test-token"), 0600)
	assert.NoError(t, err)

	recognizer, err := NewRecogniser("localhost:50051", tokenFile)
	assert.NoError(t, err)

	err = recognizer.Close()
	assert.NoError(t, err)
}
