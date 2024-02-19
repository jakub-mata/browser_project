package main

import (
	"fmt"
)

var testHTML string = `
<!DOCTYPE html>
<html>
    <head>

    </head>
    <body>
        <h1>
            Hello World
        </h1>
    </body>
</html>
`

func main() {
	//httpClient("https://www.google.com")
	testToken := NewHTMLTokenizer(testHTML)
	tokens := TokenizeHTML(testToken)
	fmt.Println(tokens)
}
