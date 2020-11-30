package single

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"text/template"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/config"
)

type Assets struct {
	Content string
	Scripts []string
	Styles  []string
}

type Page struct {
	assets Assets
	opener config.Opener
}

func New(c *models.Config, assets Assets) *Page {
	return &Page{
		assets: assets,
		opener: config.MakeOpener(c),
	}
}

func minifier() *minify.M {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	return m
}

func (p *Page) Execute(w io.Writer) error {
	tmpl, err := renderWithOpener(p.assets.Content, p.opener)
	if err != nil {
		return err
	}

	data := new(struct {
		Scripts []string
		Styles  []string
	})

	m := minifier()
	for _, script := range p.assets.Scripts {
		b, err := p.opener(script)
		if err != nil {
			return err
		}
		b, err = m.Bytes("application/javascript", b)
		if err != nil {
			return err
		}
		data.Scripts = append(data.Scripts, string(b))
	}

	for _, style := range p.assets.Styles {
		b, err := p.opener(style)
		if err != nil {
			return err
		}
		b, err = m.Bytes("text/css", b)
		if err != nil {
			return err
		}
		data.Styles = append(data.Styles, string(b))
	}

	return tmpl.ExecuteTemplate(w, "layout", data)
}

func (p *Page) Executor() (func(io.Writer), error) {
	b := bytes.NewBuffer([]byte{})
	err := p.Execute(b)
	if err != nil {
		return nil, err
	}
	return func(w io.Writer) {
		w.Write(b.Bytes())
	}, nil
}

func renderWithOpener(path string, readFunc config.Opener) (*template.Template, error) {
	str := new(strings.Builder)

	b, err := readFunc("web/tmpl/layout.html")
	if err != nil {
		return nil, err
	}
	_, err = str.Write(b)
	if err != nil {
		return nil, err
	}

	b, err = readFunc(path)
	if err != nil {
		return nil, err
	}
	_, err = str.Write(b)
	if err != nil {
		return nil, err
	}

	return template.New("ui").Parse(str.String())
}
