# Dora
A simple browser written in Go to learn Go fundamentals and basics of the web.

## Description
The browser is basically a simple HTML viewer. As of now, the supported tags include:
- meta tags: *DOCTYPE, html, head, title*
- structure tags: *header, main, body, footer, div, span*
- text-based tags: *p, h1 - h6, a, br, hr*
- lists: *ul, li*
- text-style tags: *strong, em*

Only essential tag attributes are supported ()e.g. href for anchor tags. Style tags (used for CSS styling) inside HTML tag elements are not supported.

## Developer information
Based on the provided web address, a client sends a get request. The web client is created from Go's standard library.

The main program, which handles the response body, is structured into 3 parts:
1. Tokenizer (tokenizer.go)
2. Parser / Tree-constructer (treeConstruction.go)
3. Viewer (viewer.go)

### Tokenizer
Our tokenizer takes a stream of bytes and creates a slice of HTML Tokens. In theory, the tokenizer should emit tokens inidividually to allow the parser to insert and change the DOM (parsing tree). However, this is not required with HTML alone and so I've allowed myself to make this simplication and return the slice of tokens at once.