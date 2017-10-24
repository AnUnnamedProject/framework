// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// RemoveHTMLComments removes the HTML comments from provided []byte, returns a []byte
func RemoveHTMLComments(content []byte) []byte {
	htmlcmt := regexp.MustCompile(`<!--[^>]*-->`)
	return htmlcmt.ReplaceAll(content, []byte(""))
}

// MinifyHTML returns a minified HTML version.
// Removes all HTML comments.
// Remove all leading and trailing white space of each line.
// Pad a single space to the line if its length > 0
func MinifyHTML(html []byte) string {
	// read line by line
	minifiedHTML := ""
	scanner := bufio.NewScanner(bytes.NewReader(RemoveHTMLComments(html)))
	for scanner.Scan() {
		// all leading and trailing white space of each line are removed
		lineTrimmed := strings.TrimSpace(scanner.Text())
		minifiedHTML += lineTrimmed
		if lineTrimmed != "" {
			// in case of following trimmed line:
			// <div id="foo"
			minifiedHTML += " "
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return minifiedHTML
}

// ConcatenateCSS returns concatenated CSS paths to []byte
func ConcatenateCSS(paths []string) []byte {
	var concatenated []byte
	for _, path := range paths {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		concatenated = append(concatenated, b...)
	}
	return concatenated
}

// RemoveCSSComments returns a CSS comments striped string
func RemoveCSSComments(content []byte) []byte {
	return regexp.MustCompile(`/\*([^*]|[\r\n]|(\*+([^*/]|[\r\n])))*\*+/`).ReplaceAll(content, []byte(""))
}

// MinifyCSS returns the minified CSS string
func MinifyCSS(css []byte) string {
	cssAllNoComments := RemoveCSSComments(css)

	// read line by line
	minifiedCSS := ""
	scanner := bufio.NewScanner(bytes.NewReader(cssAllNoComments))
	for scanner.Scan() {
		// all leading and trailing white space of each line are removed
		minifiedCSS += strings.TrimSpace(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return minifiedCSS
}

// MinifyJS returns a compressed javascript using Google Closure Compiler
func MinifyJS(js []byte) string {

	params := url.Values{}
	params.Set("js_code", string(js))
	params.Set("compilation_level", "SIMPLE_OPTIMIZATIONS")
	params.Set("output_format", "text")
	params.Set("output_info", "compiled_code")

	resp, err := http.PostForm("https://closure-compiler.appspot.com/compile", params)
	if err != nil {
		panic(err)
	}

	defer func() {
		berr := resp.Body.Close()
		if berr != nil {
			Log.Error(berr)
		}
	}()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Log.Error(err)
		return ""
	}

	return string(b)
}
