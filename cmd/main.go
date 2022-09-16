package main

import (
	"fmt"
	"strings"

	"github.com/likearthian/fmter"
)

func main() {
	//str := "  <foo>bar</foo><foo><baz>boo</baz><baz>bee</baz></foo>"
	res := fmter.PrettyXMLStream(strings.NewReader(xmlStr2), fmter.XmlIndent("  "))
	fmt.Println(res)
}

var xmlStr2 = `
<root><this><is>a</is><test /><message><!-- with comment --><org><cn>Some org-or-other</cn><ph>Wouldnt you like to know</ph></org><contact><fn>Pat</fn><ln>Califia</ln></contact></message></this></root>
`
