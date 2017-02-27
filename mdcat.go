//
// mdcat
// Available at http://github.com/ec1oud/mdcat
// based on example code from http://github.com/russross/blackfriday-tool
//
// Copyright © 2017 Shawn Rutledge <s@ecloud.org>.
// Distributed under the Simplified BSD License.
// See README.md for details.
//

package main

import (
	"flag"
	"fmt"
	"github.com/ec1oud/blackfriday"
	//~ "github.com/russross/blackfriday"
	"io/ioutil"
	"os"
	"strings"
)

const DEFAULT_TITLE = ""

func main() {
	// parse command-line options
	var ansi, html, page, toc, toconly, xhtml, latex, smartypants, latexdashes, fractions bool
	var css string
	flag.BoolVar(&ansi, "ansi", true,
		"Generate ANSI terminal escape codes")
	flag.BoolVar(&page, "page", false,
		"Generate a standalone HTML page (implies -latex=false)")
	flag.BoolVar(&toc, "toc", false,
		"Generate a table of contents (implies -latex=false)")
	flag.BoolVar(&toconly, "toconly", false,
		"Generate a table of contents only (implies -toc)")
	flag.BoolVar(&xhtml, "xhtml", true,
		"Use XHTML-style tags in HTML output")
	flag.BoolVar(&latex, "latex", false,
		"Generate LaTeX output instead of HTML")
	flag.BoolVar(&smartypants, "smartypants", true,
		"Apply smartypants-style substitutions")
	flag.BoolVar(&latexdashes, "latexdashes", true,
		"Use LaTeX-style dash rules for smartypants")
	flag.BoolVar(&fractions, "fractions", true,
		"Use improved fraction rules for smartypants")
	flag.StringVar(&css, "css", "",
		"Link to a CSS stylesheet (implies -page)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "mdcat Markdown Renderer using Blackfriday v"+blackfriday.VERSION+
			"\nAvailable at http://github.com/ec1oud/mdcat\n\n"+
			"Usage:\n"+
			"  %s [options] [inputfile [outputfile]]\n\n"+
			"Options:\n",
			os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// enforce implied options
	if css != "" {
		page = true
		html = true
		ansi = false
	}
	if page {
		latex = false
		html = true
		ansi = false
	}
	if toconly {
		toc = true
	}
	if toc {
		latex = false
		html = true
		ansi = false
	}

	// read the input
	var input []byte
	var err error
	args := flag.Args()
	switch len(args) {
	case 0:
		if input, err = ioutil.ReadAll(os.Stdin); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading from Stdin:", err)
			os.Exit(-1)
		}
	case 1, 2:
		if input, err = ioutil.ReadFile(args[0]); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading from", args[0], ":", err)
			os.Exit(-1)
		}
	default:
		flag.Usage()
		os.Exit(-1)
	}

	// set up options
	extensions := 0
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS

	var renderer blackfriday.Renderer
	if latex {
		// render the data into LaTeX
		renderer = blackfriday.LatexRenderer(0)
	} else if html {
		// render the data into HTML
		htmlFlags := 0
		if xhtml {
			htmlFlags |= blackfriday.HTML_USE_XHTML
		}
		if smartypants {
			htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
		}
		if fractions {
			htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
		}
		if latexdashes {
			htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
		}
		title := ""
		if page {
			htmlFlags |= blackfriday.HTML_COMPLETE_PAGE
			title = getTitle(input)
		}
		if toconly {
			htmlFlags |= blackfriday.HTML_OMIT_CONTENTS
		}
		if toc {
			htmlFlags |= blackfriday.HTML_TOC
		}
		renderer = blackfriday.HtmlRenderer(htmlFlags, title, css)
	} else {
		renderer = blackfriday.AnsiRenderer(0)
	}

	// parse and render
	var output []byte = blackfriday.Markdown(input, renderer, extensions)

	// output the result
	var out *os.File
	if len(args) == 2 {
		if out, err = os.Create(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating %s: %v", args[1], err)
			os.Exit(-1)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	if _, err = out.Write(output); err != nil {
		fmt.Fprintln(os.Stderr, "Error writing output:", err)
		os.Exit(-1)
	}
}

// try to guess the title from the input buffer
// just check if it starts with an <h1> element and use that
func getTitle(input []byte) string {
	i := 0

	// skip blank lines
	for i < len(input) && (input[i] == '\n' || input[i] == '\r') {
		i++
	}
	if i >= len(input) {
		return DEFAULT_TITLE
	}
	if input[i] == '\r' && i+1 < len(input) && input[i+1] == '\n' {
		i++
	}

	// find the first line
	start := i
	for i < len(input) && input[i] != '\n' && input[i] != '\r' {
		i++
	}
	line1 := input[start:i]
	if input[i] == '\r' && i+1 < len(input) && input[i+1] == '\n' {
		i++
	}
	i++

	// check for a prefix header
	if len(line1) >= 3 && line1[0] == '#' && (line1[1] == ' ' || line1[1] == '\t') {
		return strings.TrimSpace(string(line1[2:]))
	}

	// check for an underlined header
	if i >= len(input) || input[i] != '=' {
		return DEFAULT_TITLE
	}
	for i < len(input) && input[i] == '=' {
		i++
	}
	for i < len(input) && (input[i] == ' ' || input[i] == '\t') {
		i++
	}
	if i >= len(input) || (input[i] != '\n' && input[i] != '\r') {
		return DEFAULT_TITLE
	}

	return strings.TrimSpace(string(line1))
}
