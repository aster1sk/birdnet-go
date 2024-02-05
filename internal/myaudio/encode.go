package myaudio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/tphakala/birdnet-go/internal/conf"
)

// seekableBuffer extends bytes.Buffer to add a Seek method, making it compatible with io.WriteSeeker.
type seekableBuffer struct {
	bytes.Buffer
	pos int64
}

// Seek implements the io.Seeker interface.
func (b *seekableBuffer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		b.pos = offset
	case io.SeekCurrent:
		b.pos += offset
	case io.SeekEnd:
		b.pos = int64(b.Len()) + offset
	default:
		return 0, errors.New("seekableBuffer.Seek: invalid whence")
	}

	if b.pos < 0 {
		return 0, errors.New("seekableBuffer.Seek: negative position")
	}

	return b.pos, nil
}

// ConvertPCMToWAV converts PCM data to WAV format.
func ConvertPCMToWAV(pcmData []byte, sampleRate, bitDepth, numChannels int) ([]byte, error) {
	// Create a new seekable buffer
	buf := new(seekableBuffer)

	// Create a new WAV encoder with the seekable buffer as the writer
	enc := wav.NewEncoder(buf, conf.SampleRate, conf.BitDepth, conf.NumChannels, 1)

	// Convert the byte slice to a slice of integer samples
	intSamples := byteSliceToInts(pcmData)

	// Create an audio buffer from the integer samples
	audioBuf := &audio.IntBuffer{Data: intSamples, Format: &audio.Format{
		SampleRate:  sampleRate,
		NumChannels: numChannels,
	}}

	// Write the PCM data to the encoder
	if err := enc.Write(audioBuf); err != nil {
		return nil, err
	}

	// Close the encoder to finalize the WAV format
	if err := enc.Close(); err != nil {
		return nil, err
	}

	// Return the bytes from the buffer
	return buf.Bytes(), nil
}

// SavePCMDataToWAV saves the given PCM data as a WAV file at the specified filePath.
func SavePCMDataToWAV(filePath string, pcmData []byte) error {
	// Create the directory structure if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Open a new file for writing.
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close() // Ensure file closure on function exit.

	// Create a new WAV encoder with the specified format settings.
	enc := wav.NewEncoder(outFile, conf.SampleRate, conf.BitDepth, conf.NumChannels, 1)

	// Convert the byte slice to a slice of integer samples.
	intSamples := byteSliceToInts(pcmData)

	// Write the integer samples to the WAV file.
	if err := enc.Write(&audio.IntBuffer{Data: intSamples, Format: &audio.Format{SampleRate: conf.SampleRate, NumChannels: conf.NumChannels}}); err != nil {
		return fmt.Errorf("failed to write to WAV encoder: %w", err)
	}

	// Close the WAV encoder, which finalizes the file format.
	return enc.Close()
}

// byteSliceToInts converts a byte slice to a slice of integers.
// Each pair of bytes is treated as a single 16-bit sample.
func byteSliceToInts(pcmData []byte) []int {
	var samples []int
	buf := bytes.NewBuffer(pcmData)

	// Read each 16-bit sample from the byte buffer and store it as an int.
	for {
		var sample int16
		if err := binary.Read(buf, binary.LittleEndian, &sample); err != nil {
			break // Exit loop on read error (e.g., end of buffer).
		}
		samples = append(samples, int(sample))
	}

	return samples
}
