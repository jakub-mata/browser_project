package main

import (
	"flag"
	"fmt"
)

var defaultSite string = "http://0.0.0.0:8080/"

func startCLI() {
	logTokensBool := flag.Bool("log-tokens", false, "Logs tokenizer's output")
	logParserBool := flag.Bool("log-parser", true, "Logs constructed parsing tree")
	websiteToFetch := flag.String("address", defaultSite, "The web address to search")
	flag.Parse()
	fmt.Println("tokens:", *logTokensBool)
	fmt.Println("parser:", *logParserBool)
	fmt.Println("website:", *websiteToFetch)
}
