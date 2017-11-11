// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Renderer is the interface for template rendering.
type Renderer interface {
	Render(out io.Writer, name string, data interface{}) error
	Parse(name string, data []byte) error
}

var funcMap template.FuncMap

// View implements the Renderer interface.
type View struct {
	template.Template
	viewDir string
}

// AddFunc register a func in the view.
func AddFunc(key string, fn interface{}) {
	funcMap[key] = fn
}

// NewView returns a View with templates loaded from viewDir.
func NewView(viewDir string) (Renderer, error) {
	info, err := os.Stat(viewDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("View: %s is not a directory\n", viewDir)
	}

	// Register internal view funcs
	funcMap = GetTemplateFuncs()

	s := &View{
		viewDir:  viewDir,
		Template: *template.New("").Delims(Config.String("template_left"), Config.String("template_right")).Funcs(funcMap),
	}

	s.EmbedShortcodes()
	s.EmbedTemplates()

	return s.load(viewDir)
}

// load loads the .html templates from the specified dir.
func (s *View) load(dir string) (Renderer, error) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		extension := filepath.Ext(path)
		if extension != ".html" {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		name := path[len(dir):]
		name = filepath.ToSlash(name)
		name = strings.TrimPrefix(name, "/")
		name = strings.TrimSuffix(name, extension)

		t := s.New(name)

		s := string(data)

		if Config.Bool("compress_html") {
			s = MinifyHTML([]byte(s))
		}

		if Config.Bool("compress_css") {
			re := regexp.MustCompile(`<style type="text/css">([\s\S]*)<\/style>`)
			matches := re.FindStringSubmatch(s)
			for i := 1; i < len(matches); i++ {
				re.ReplaceAllString(matches[i], MinifyCSS([]byte(matches[i])))
			}
		}

		if Config.Bool("compress_js") {
			re := regexp.MustCompile(`<script type="text/javascript">([\s\S]*)<\/script>`)
			matches := re.FindStringSubmatch(s)
			for i := 1; i < len(matches); i++ {
				re.ReplaceAllString(matches[i], MinifyJS([]byte(matches[i])))
			}
		}

		_, err = t.Parse(s)

		return err
	})

	if err != nil {
		return nil, err
	}

	return s, nil
}

// Render executes the template by name.
func (s *View) Render(out io.Writer, name string, data interface{}) error {
	return s.ExecuteTemplate(out, name, data)
}

// Parse the data template.
func (s *View) Parse(name string, data []byte) error {
	t := s.New(name)
	_, err := t.Parse(string(data))
	return err
}
