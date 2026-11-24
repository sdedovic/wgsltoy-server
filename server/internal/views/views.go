package views

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
)

// Templates are the pages, each rendered inside layout.html.
//
// Every page is its own template set: layout.html and partials/*.html, shared by all pages, plus the
// page's file, which fills the layout's blocks, e.g. {{define "content"}}. Separate sets let every page
// define the same block names.
type Templates struct {
	Gallery  *Page
	Login    *Page
	Register *Page
	Me       *Page
	Profile  *Page
	Shader   *Page
}

// Page is a template that can be rendered.
type Page struct {
	load func() (*template.Template, error)
	tmpl *template.Template
	// reload re-reads the templates on every render, so changes show without a restart
	reload bool
}

// NewTemplates parses the templates in dir. With reload, they're re-read on every render.
func NewTemplates(dir string, vite *Vite, reload bool) (*Templates, error) {
	shared := func() (*template.Template, error) {
		t, err := template.New("").
			Option("missingkey=error").
			Funcs(vite.FuncMap()).
			ParseFiles(filepath.Join(dir, "layout.html"))
		if err != nil {
			return nil, err
		}
		return t.ParseGlob(filepath.Join(dir, "partials", "*.html"))
	}
	base, err := shared()
	if err != nil {
		return nil, fmt.Errorf("parsing shared templates: %w", err)
	}

	page := func(name string) (*Page, error) {
		path := filepath.Join(dir, name)
		p := &Page{reload: reload}
		p.load = func() (*template.Template, error) {
			b := base
			if p.reload {
				var err error
				if b, err = shared(); err != nil {
					return nil, err
				}
			}
			t, err := b.Clone()
			if err != nil {
				return nil, err
			}
			return t.ParseFiles(path)
		}
		t, err := p.load()
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", name, err)
		}
		p.tmpl = t
		return p, nil
	}

	var t Templates
	for _, entry := range []struct {
		page **Page
		name string
	}{
		{&t.Gallery, "gallery.html"},
		{&t.Login, "login.html"},
		{&t.Register, "register.html"},
		{&t.Me, "me.html"},
		{&t.Profile, "profile.html"},
		{&t.Shader, "shader.html"},
	} {
		if *entry.page, err = page(entry.name); err != nil {
			return nil, err
		}
	}
	return &t, nil
}

func (p *Page) Execute(w io.Writer, data any) error {
	t := p.tmpl
	if p.reload {
		var err error
		if t, err = p.load(); err != nil {
			return err
		}
	}
	return t.ExecuteTemplate(w, "layout.html", data)
}
