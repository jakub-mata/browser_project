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
	tokens := TokenizeHTML(testHTML)
	for _, val := range tokens {
		fmt.Println(val)
	}
}
