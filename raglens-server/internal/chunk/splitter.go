package chunk

import (
	"errors"
	"strings"
)

const (
	DefaultSize    = 500
	DefaultOverlap = 100
)

var ErrInvalidOptions = errors.New("invalid chunk options")

type Options struct {
	Size    int
	Overlap int
}

type Splitter struct {
	options Options
}

type Chunk struct {
	Index         int
	Content       string
	ContentLength int
}

func NewSplitter(options Options) *Splitter {
	if options.Size == 0 {
		options.Size = DefaultSize
	}
	if options.Overlap == 0 {
		options.Overlap = DefaultOverlap
	}
	return &Splitter{options: options}
}

func (s *Splitter) Split(text string) ([]Chunk, error) {
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}
	if s.options.Size <= 0 || s.options.Overlap < 0 || s.options.Overlap >= s.options.Size {
		return nil, ErrInvalidOptions
	}

	runes := []rune(text)
	chunks := make([]Chunk, 0, len(runes)/s.options.Size+1)
	step := s.options.Size - s.options.Overlap

	for start := 0; start < len(runes); start += step {
		end := start + s.options.Size
		if end > len(runes) {
			end = len(runes)
		}

		content := string(runes[start:end])
		chunks = append(chunks, Chunk{
			Index:         len(chunks),
			Content:       content,
			ContentLength: len([]rune(content)),
		})

		if end == len(runes) {
			break
		}
	}

	return chunks, nil
}
