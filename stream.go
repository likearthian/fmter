package fmter

import (
	"bufio"
	"bytes"
	"io"
)

type streamLexer struct {
	name   string
	reader *bufio.Reader
	input  string
	start  int
	pos    int
	width  int
	buffer *bytes.Buffer
	items  chan item
}

func (s *streamLexer) run(startState streamStateFn) <-chan item {
	go func() {
		for state := startState; state != nil; {
			state = state(s)
		}
		close(s.items)
	}()

	return s.items
}

func (s *streamLexer) emit(t itemType) {
	val := s.buffer.String()[:s.pos]
	s.items <- item{t, val}
	s.pos = 0
	s.buffer.Reset()
}

func (s *streamLexer) next() rune {
	r, n, err := s.reader.ReadRune()
	if err != nil {
		if err == io.EOF {
			s.width = 0
			return eof
		}
		panic(err)
	}
	s.width = n
	s.pos += s.width
	s.buffer.WriteRune(r)
	return r
}

func (s *streamLexer) back() {
	err := s.reader.UnreadRune()
	if err != nil {
		panic(err)
	}

	s.pos -= s.width
}

func (s *streamLexer) ignore() {
	s.pos = 0
	s.buffer.Reset()
}

func (s *streamLexer) peek() rune {
	r, _, err := s.reader.ReadRune()
	if err != nil {
		if err == io.EOF {
			return eof
		}
		panic(err)
	}
	if err := s.reader.UnreadRune(); err != nil {
		panic(err)
	}
	return r
}

func (s *streamLexer) skip() {
	s.back()
	s.buffer = bytes.NewBuffer(s.buffer.Bytes()[:s.pos])
	_, _, _ = s.reader.ReadRune()
}

func lexStreamEOF(s *streamLexer) streamStateFn {
	if s.pos > 0 {
		s.emit(itemText)
	}
	s.emit(itemEOF)
	return nil
}
