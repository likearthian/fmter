package fmter

import (
	"strings"
	"testing"
)

func TestPrettyXML(t *testing.T) {
	str := `  <foo>bar</foo><foo><baz>boo</baz></foo>`
	t.Log("source:", str)
	res := PrettyXML(strings.NewReader(str))
	t.Log(res)
}
