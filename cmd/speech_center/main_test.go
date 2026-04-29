package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePronunciationDict(t *testing.T) {
	t.Run("empty string returns nil", func(t *testing.T) {
		dict, err := parsePronunciationDict("")
		assert.NoError(t, err)
		assert.Nil(t, dict)
	})

	t.Run("valid JSON string", func(t *testing.T) {
		dict, err := parsePronunciationDict(`{"Claughton": "ˈklɒftən", "live": "laɪv"}`)
		assert.NoError(t, err)
		assert.Equal(t, "ˈklɒftən", dict["Claughton"])
		assert.Equal(t, "laɪv", dict["live"])
	})

	t.Run("valid JSON file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "pronunciation.json")
		err := os.WriteFile(filePath, []byte(`{"word": "wɜːrd"}`), 0600)
		assert.NoError(t, err)

		dict, err := parsePronunciationDict(filePath)
		assert.NoError(t, err)
		assert.Equal(t, "wɜːrd", dict["word"])
	})

	t.Run("invalid JSON and nonexistent file", func(t *testing.T) {
		_, err := parsePronunciationDict("not-json-and-not-a-file")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not valid JSON and cannot read file")
	})

	t.Run("file with invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "bad.json")
		err := os.WriteFile(filePath, []byte(`not json`), 0600)
		assert.NoError(t, err)

		_, err = parsePronunciationDict(filePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON")
	})
}
