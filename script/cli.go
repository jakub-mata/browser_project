package main

import (
	"flag"
	"fmt"
)

var defaultSite string = "https://jakub-mata.github.io/"

func startCLI() (bool, bool, string) {
	logTokensBool := flag.Bool("log-tokens", false, "Logs tokenizer's output")
	logParserBool := flag.Bool("log-parser", false, "Logs constructed parsing tree")
	//websiteToFetch := flag.String("address", defaultSite, "The web address to search")
	flag.Parse()
	websiteToFetch := defaultSite
	if flag.Arg(0) != "" {
		websiteToFetch = flag.Arg(0)
	}
	fmt.Println("tokens:", *logTokensBool)
	fmt.Println("parser:", *logParserBool)
	fmt.Println("website:", websiteToFetch)
	return *logTokensBool, *logParserBool, websiteToFetch
}
