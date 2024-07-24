# Dora
A simple browser written in Go to learn Go fundamentals and basics of the web.

## Description
The browser is basically a simple HTML viewer. As of now, the supported tags include:
- meta tags: *DOCTYPE, html, head, title*
- structure tags: *header, main, body, footer, div, span*
- text-based tags: *p, h1 - h6, a, br, hr*
- lists: *ul, li*
- text-style tags: *strong, em*

Only essential tag attributes are supported, e.g. *href* for anchor tags. Style tags (used for CSS styling) inside HTML tag elements are not supported.

## Support
The program should run on all platforms supported by [Fyne.io](https://fyne.io/) (including Linux, Windows, and MacOS). 

## Installation

## Usage
You can view a webpage by using this command:
```
dora -web *websiteAddress*
```
The address has to be precise. There are two additional flags you can use for debugging.

`-log-tokens` prints out the result of the tokenization stage to the terminal

`-log-parser` prints out the result of the parsing and tree construction stage

After you call the main command, a new window will pop up and display the desired web page.

## Developer information
Based on the provided web address, a client sends a get request. The web client is created from Go's standard library.

The main program, which handles the response body, is structured into 3 parts:
1. Tokenizer (tokenizer.go)
2. Parser / Tree-constructer (treeConstruction.go)
3. Viewer (viewer.go)

### Tokenizer
Our tokenizer takes a stream of bytes and creates a slice of HTML Tokens. In theory, the tokenizer should emit tokens inidividually to allow the parser to insert and change the DOM (parsing tree). However, this is not required with HTML alone and so I've allowed myself to make this simplication and return the slice of tokens at once.
The output of the tokenization stage can be turned on with the `log-tokens` flag

### Parser
The parser gets a slice of HTML Tokens from the tokenizer and creates a parse tree. The output of the tree-construction stage can be turned on with the `log-parser` flag.

### Viewer
Finally, an HTML viewer will render our page when given a root of the parsing tree. It is written with [Fyne.io](https://fyne.io/). 

Each node has a container. Before initializing this container, each node receives (recursively) objects from its children which should be contained within the container. After the recursion ends (and our root has its container filled), the page is showed.