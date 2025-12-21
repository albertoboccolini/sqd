package tests

import (
	"os"
	"testing"

	"github.com/albertoboccolini/sqd/services"
)

func TestIsTextFileText(t *testing.T) {
	file, _ := os.CreateTemp("", "test*.txt")
	defer os.Remove(file.Name())
	file.WriteString("This is plain text\n")
	file.Close()

	if !services.IsTextFile(file.Name()) {
		t.Error("text file should be detected as text")
	}
}

func TestIsTextFileBinary(t *testing.T) {
	file, _ := os.CreateTemp("", "test*.bin")
	defer os.Remove(file.Name())
	file.Write([]byte{0x00, 0x01, 0xFF, 0xFE, 0x00, 0x00})
	file.Close()

	if services.IsTextFile(file.Name()) {
		t.Error("binary file should not be detected as text")
	}
}

func TestIsTextFileNullByte(t *testing.T) {
	file, _ := os.CreateTemp("", "test*.txt")
	defer os.Remove(file.Name())
	file.WriteString("text\x00more")
	file.Close()

	if services.IsTextFile(file.Name()) {
		t.Error("file with null byte should not be text")
	}
}

func TestIsTextFileControlChars(t *testing.T) {
	file, _ := os.CreateTemp("", "test*.txt")
	defer os.Remove(file.Name())
	file.Write([]byte{0x01, 0x02, 0x03})
	file.Close()

	if services.IsTextFile(file.Name()) {
		t.Error("file with control chars should not be text")
	}
}
