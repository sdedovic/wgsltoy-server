package views

import (
	"fmt"
	"path/filepath"

	"github.com/sdedovic/htpp"
)

type Templates struct {
	Gallery  htpp.Template
	Login    htpp.Template
	Register htpp.Template
	Me       htpp.Template
	Profile  htpp.Template
	Shader   htpp.Template
	Editor   htpp.Template
}

func NewTemplates(dir string) (*Templates, error) {
	compile := func(name string) (htpp.Template, error) {
		t, err := htpp.Compile(filepath.Join(dir, name))
		if err != nil {
			return htpp.Template{}, fmt.Errorf("%s: %w", name, err)
		}
		return t, nil
	}

	gallery, err := compile("gallery.html")
	if err != nil {
		return nil, err
	}
	login, err := compile("login.html")
	if err != nil {
		return nil, err
	}
	register, err := compile("register.html")
	if err != nil {
		return nil, err
	}
	me, err := compile("me.html")
	if err != nil {
		return nil, err
	}
	profile, err := compile("profile.html")
	if err != nil {
		return nil, err
	}
	shader, err := compile("shader.html")
	if err != nil {
		return nil, err
	}
	editor, err := compile("editor.html")
	if err != nil {
		return nil, err
	}

	return &Templates{
		Gallery:  gallery,
		Login:    login,
		Register: register,
		Me:       me,
		Profile:  profile,
		Shader:   shader,
		Editor:   editor,
	}, nil
}
