package fmter

import (
	"io"
	"strings"
)

type stateFn func(*lexer) stateFn

const (
	xmlOpenTag      = "<"
	xmlCloseTag     = ">"
	xmlEndOpenTag   = "</"
	xmlEndCloseTag  = "/>"
	xmlMetaOpenTag  = "<?"
	xmlMetaCloseTag = "?>"
)

type xmlOptions struct {
	indent string
}

type PrettyXMLOption func(*xmlOptions)

func XmlIndent(indent string) PrettyXMLOption {
	return func(op *xmlOptions) {
		op.indent = indent
	}
}

func NewXMLLexer(xmlStr string) *lexer {
	return &lexer{
		name:   "xml",
		input:  xmlStr,
		start:  0,
		pos:    0,
		width:  0,
		buffer: "",
		items:  make(chan item),
	}
}

func xmlLexStart(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], "<") {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return xmlLexOpenTag
		}

		r := l.next()
		if r == eof {
			break
		}

		if isWhitespace(r) {
			l.skip()
		}
	}

	return lexEOF
}

func xmlLexOpenTag(l *lexer) stateFn {
	l.next()
	for {
		if strings.HasPrefix(l.input[l.pos:], "!") {
			return xmlLexComment
		}

		if strings.HasPrefix(l.input[l.pos:], "/") {
			return xmlLexEndOpenTag
		}

		if strings.HasPrefix(l.input[l.pos:], "?") {
			return xmlLexMeta
		}

		r := l.next()
		if r == eof {
			break
		}

		if isWhitespace(r) {
			l.skip()
			continue
		}

		if isAlphabet(r) {
			l.back()
			l.emit(itemXmlOpenTag)
			l.next()
			return xmlLexElementName
		}
	}

	return lexEOF
}

func xmlLexEndOpenTag(l *lexer) stateFn {
	l.next()
	for {
		r := l.next()
		if r == eof {
			break
		}

		if isWhitespace(r) {
			l.skip()
			continue
		}

		if isAlphabet(r) {
			l.back()
			l.emit(itemXmlEndOpenTag)
			l.next()
			return xmlLexElementName
		}
	}
	return lexEOF
}

func xmlLexMeta(l *lexer) stateFn {
	l.next()
	isInsideDblQuote := false
	for {
		r := l.next()
		if !isInsideDblQuote && r == '>' {
			l.emit(itemXmlMeta)
			return xmlLexStart
		}

		if r == eof {
			break
		}

		if r == '"' {
			isInsideDblQuote = !isInsideDblQuote
		}
	}

	return lexEOF
}

func xmlLexComment(l *lexer) stateFn {
	l.next()
	isInsideDblQuote := false
	for {
		r := l.next()
		if !isInsideDblQuote && r == '>' {
			l.emit(itemXmlComment)
			return xmlLexStart
		}

		if r == eof {
			break
		}

		if r == '"' {
			isInsideDblQuote = !isInsideDblQuote
		}
	}

	return lexEOF
}

func xmlLexElementName(l *lexer) stateFn {
	for {
		if r := l.peek(); r == ' ' || r == '/' || r == '>' {
			l.emit(itemXmlElementName)
			return xmlLexCloseTag
		}

		r := l.next()
		if r == eof {
			break
		}
	}
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
}

func xmlLexCloseTag(l *lexer) stateFn {
	isInsideDblQuote := false
	for {
		r := l.peek()
		if !isInsideDblQuote && r == '/' {
			return xmlLexEndCloseTag
		}

		if !isInsideDblQuote && strings.HasPrefix(l.input[l.pos:], ">") {
			l.next()
			l.emit(itemXmlCloseTag)
			return xmlLexStart
		}

		r = l.next()
		if r == eof {
			break
		}

		if r == '"' {
			isInsideDblQuote = !isInsideDblQuote
		}
	}

	return nil
}

func xmlLexEndCloseTag(l *lexer) stateFn {
	isInsideDblQuote := false
	for {
		if !isInsideDblQuote && strings.HasPrefix(l.input[l.pos:], ">") {
			l.next()
			l.emit(itemXmlEndCloseTag)
			return xmlLexStart
		}

		r := l.next()
		if r == eof {
			break
		}

		if r == '"' {
			isInsideDblQuote = !isInsideDblQuote
		}
	}

	return nil
}

func PrettyXML(reader io.Reader, options ...PrettyXMLOption) string {
	def := xmlOptions{indent: "  "}
	for _, op := range options {
		op(&def)
	}

	buf, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	input := string(buf)
	l := NewXMLLexer(input)
	res := ""
	depth := 0
	indent := def.indent
	isClosing := false
	isClosed := true

	token := l.run(xmlLexStart)
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
