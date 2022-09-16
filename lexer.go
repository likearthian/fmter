package fmter

import (
	"unicode/utf8"
)

type lexerItem struct {
	typ itemType
	val string
}

type lexer struct {
	name   string
	input  string
	start  int
	pos    int
	width  int
	buffer string
	items  chan item
}

func (l *lexer) run(startState stateFn) <-chan item {
	go func() {
		for state := startState; state != nil; {
			state = state(l)
		}
		close(l.items)
	}()

	return l.items
}

func (l *lexer) emit(t itemType) {
	val := l.buffer + l.input[l.start:l.pos]
	l.items <- item{t, val}
	l.start = l.pos
	l.buffer = ""
}

func (l *lexer) next() rune {
	var r rune
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}

	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) back() {
	l.pos -= l.width
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) peek() rune {
	r := l.next()
	l.back()
	return r
}

func (l *lexer) skip() {
	l.back()
	l.buffer = l.input[l.start:l.pos]
	l.next()
	l.ignore()
}

func lexEOF(l *lexer) stateFn {
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
}

var eof = int32(-1)
