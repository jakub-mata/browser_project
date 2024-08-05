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

First, clone the project from [GitHub](https://github.com/jakub-mata/browser_project):
```
git clone https://github.com/jakub-mata/browser_project.git
```
or [GitLab](https://gitlab.mff.cuni.cz/teaching/nprg031/2324-summer/student-mataj):
```
git clone https://gitlab.mff.cuni.cz/teaching/nprg031/2324-summer/student-mataj.git
```
You will also need to install Go if you haven't done so yet. The project was built with Go 1.21 and 1.22, but any newest stable version should work fine. You can follow these [instructions](https://go.dev/dl/).

Next, change to the directory where you cloned the projrect and go to the script directory. Make sure you have make installed (e.g. Windows doesn't have it by default: you can install it on Windows with `choco install make`, provided you have the [Chocolatey package manager](https://chocolatey.org/)). Run:
```
make
```
to install dependecies and build the project (you can view this in the *Makefile*). This will take a while (seriously, some of the dependencies are written in C and its compiler is slower at times; you can expect several minutes at most). After this step, you should be ready to run the application.

## Usage
Switch to the script directory inside the cloned repo. Then can view a webpage by using this command:
```
go run dora -web *websiteAddress*
```
The address has to be precise. If nothing is passed, the default address is *https://jakub-mata.github.io/* (see more in Demo). There are two additional flags you can use for debugging.

`-log-tokens` prints out the result of the tokenization stage to the terminal

`-log-parser` prints out the result of the parsing and tree construction stage

After you call the main command, a new window will pop up and display the desired web page.

> A slight warning: The shown webpages will look off. This is due to the app only supporting HTML (and only a subset of all HTML tags, for that matter).

### Demo
The default website is another repo at my [Github](https://github.com/jakub-mata/jakub-mata.github.io), which contains all the elements currently supported. You can view this page in your favorite browser [here](https://jakub-mata.github.io/).

### Recommended websites
These websites are purely HTML and should mostly looks as intended:
- http://motherfuckingwebsite.com/
- https://ukarim.com/ (a software engineer's personal page)
- https://wittallen.net/ (the same thing)

Most common sites rely on dynamic rendering with Javascript, but you can still try to view them in their pure HTML form:
- https://www.google.com/ (might throw errors if they have a special day animation)
- https://www.wikipedia.org/
- https://viethung.space/blog/2020/10/24/Browser-from-Scratch-HTML-parsing/ (a source used for this project)


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

## Known issues
There are a few issues which need to address in future releases:
1. *Character references*: currently treated as plain text. For example, the default google web page for Czech uses Latin1 encoding with some characters (e.g. "ř") represented with their character reference code.
2. *Encodings*: only utf-8 and Latin1 are currently supported
3. *Long text*: due to limitations of the [Fyne.io](https://fyne.io/) library, there isn't an easy way to wrap text horizontally. As of now, a long text would expand beyond the page.
4. *Text within text*: Tags like `<strong>` or `<a>` can be found inside a text (like a `<p>` tag). Contents of these tags will have an unwanted padding, which separates them from the surrounding text. This is a default behavior of containers in [Fyne.io](https://fyne.io/), which is difficult to overwrite.