package fmter

type itemType int

const (
	itemError itemType = iota
	itemDot
	itemEOF
	itemLess
	itemExclamation
	itemQuestion
	itemHyphen
	itemGreater
	itemNumber
	itemText
	itemEqual
	itemQuote
	itemDoubleQuote
	itemXmlOpenTag
	itemXmlCloseTag
	itemXmlEndOpenTag
	itemXmlEndCloseTag
	itemXmlElementName
	itemXmlMeta
	itemXmlComment
)

type item struct {
	typ itemType
	val string
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}

	return i.val
}
