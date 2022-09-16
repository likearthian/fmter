package fmter

import (
	"runtime"
)

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}

func isNumber(r rune) bool {
	return '0' <= r && r <= '9'
}

func isAlphabet(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

func isNewLine(r rune) bool {
	nl := '\n'
	if runtime.GOOS == "linux" {
		nl = '\r'
	}

	return r == nl
}

func isDblQuote(r rune) bool {
	return r == '"'
}
