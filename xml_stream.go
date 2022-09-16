package fmter

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

type streamStateFn func(*streamLexer) streamStateFn

func NewXMLStreamLexer(reader io.Reader) *streamLexer {
	return &streamLexer{
		name:   "xmlStream",
		reader: bufio.NewReader(reader),
		start:  0,
		pos:    0,
		width:  0,
		buffer: new(bytes.Buffer),
		items:  make(chan item),
	}
}

func xmlStreamLexStart(s *streamLexer) streamStateFn {
	for {
		if s.peek() == '<' {
			if s.pos > s.start {
				s.emit(itemText)
			}
			return xmlStreamLexOpenTag
		}

		r := s.next()
		if r == eof {
			break
		}

		if isWhitespace(r) {
			s.skip()
		}
	}

	return lexStreamEOF
}

func xmlStreamLexOpenTag(s *streamLexer) streamStateFn {
	s.next()
	for {
		r := s.peek()
		if r == '!' {
			return xmlStreamLexComment
		}

		if r == '/' {
			return xmlStreamLexEndOpenTag
		}

		if r == '?' {
			return xmlStreamLexMeta
		}

		r = s.next()
		if r == eof {
			break
		}

		if isWhitespace(r) {
			s.skip()
			continue
		}

		if isAlphabet(r) {
			s.back()
			s.emit(itemXmlOpenTag)
			s.next()
			return xmlStreamLexElementName
		}
	}

	return lexStreamEOF
}

func xmlStreamLexEndOpenTag(s *streamLexer) streamStateFn {
	s.next()
	for {
		r := s.next()
		if r == eof {
			break
		}

		if isWhitespace(r) {
			s.skip()
			continue
		}

		if isAlphabet(r) {
			s.back()
			s.emit(itemXmlEndOpenTag)
			s.next()
			return xmlStreamLexElementName
		}
	}
	return lexStreamEOF
}

func xmlStreamLexMeta(s *streamLexer) streamStateFn {
	s.next()
	isInsideDblQuote := false
	for {
		r := s.next()
		if !isInsideDblQuote && r == '>' {
			s.emit(itemXmlMeta)
			return xmlStreamLexStart
		}

		if r == eof {
			break
		}

		if r == '"' {
			isInsideDblQuote = !isInsideDblQuote
		}
	}

	return lexStreamEOF
}

func xmlStreamLexComment(s *streamLexer) streamStateFn {
	s.next()
	isInsideDblQuote := false
	for {
		r := s.next()
		if !isInsideDblQuote && r == '>' {
			s.emit(itemXmlComment)
			return xmlStreamLexStart
		}

		if r == eof {
			break
		}

		if r == '"' {
			isInsideDblQuote = !isInsideDblQuote
		}
	}

	return lexStreamEOF
}

func xmlStreamLexElementName(s *streamLexer) streamStateFn {
	for {
		if r := s.peek(); r == ' ' || r == '/' || r == '>' {
			s.emit(itemXmlElementName)
			return xmlStreamLexCloseTag
		}

		r := s.next()
		if r == eof {
			break
		}
	}

	return lexStreamEOF
}

func xmlStreamLexCloseTag(s *streamLexer) streamStateFn {
	isInsideDblQuote := false
	for {
		r := s.peek()
		if !isInsideDblQuote && r == '/' {
			return xmlStreamLexEndCloseTag
		}

		if !isInsideDblQuote && r == '>' {
			s.next()
			s.emit(itemXmlCloseTag)
			return xmlStreamLexStart
		}

		r = s.next()
		if r == eof {
			break
		}

		if r == '"' {
			isInsideDblQuote = !isInsideDblQuote
		}
	}

	return lexStreamEOF
}

func xmlStreamLexEndCloseTag(s *streamLexer) streamStateFn {
	isInsideDblQuote := false
	for {
		if !isInsideDblQuote && s.peek() == '>' {
			s.next()
			s.emit(itemXmlEndCloseTag)
			return xmlStreamLexStart
		}

		r := s.next()
		if r == eof {
			break
		}

		if r == '"' {
			isInsideDblQuote = !isInsideDblQuote
		}
	}

	return lexStreamEOF
}

func PrettyXMLStream(reader io.Reader, options ...PrettyXMLOption) string {
	def := xmlOptions{indent: "  "}
	for _, op := range options {
		op(&def)
	}

	l := NewXMLStreamLexer(reader)
	res := ""
	depth := 0
	indent := def.indent
	isClosing := false
	isClosed := true

	token := l.run(xmlStreamLexStart)
	for t := range token {
		switch t.typ {
		case itemXmlOpenTag:
			isClosing = false
			isClosed = false
		case itemXmlElementName:
			tagStr := "<"

			if isClosing {
				tagStr = "</"
				if isClosed {
					res += "\n" + strings.Repeat(indent, depth)
				}
			} else {
				res += "\n" + strings.Repeat(indent, depth)
			}

			res += tagStr + t.String()
		case itemXmlCloseTag:
			res += t.String()
			if isClosing {
				isClosed = true
			} else {
				depth++
			}
		case itemXmlEndCloseTag:
			isClosed = true
			res += t.String()
		case itemText:
			res += t.String()
		case itemXmlEndOpenTag:
			isClosing = true
			depth--
		case itemXmlMeta:
			res += "\n" + strings.Repeat(indent, depth) + t.String()
			isClosed = true
		case itemXmlComment:
			res += "\n" + strings.Repeat(indent, depth) + t.String()
			isClosed = true
		}
	}

	return res
}
