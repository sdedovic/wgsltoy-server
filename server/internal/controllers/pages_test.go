package controllers

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
	"github.com/sdedovic/wgsltoy-server/internal/views"
)

// Renders every page with the data its handler passes, signed in and out, so a field renamed on one side
// fails here rather than on a request.
func TestPagesRenderWithTheirData(t *testing.T) {
	signedIn := &domain.Principal{User: domain.User{Username: "someone"}}
	shader := domain.Shader{Content: "fn mainImage(p: vec2f) -> vec4f { return vec4f(1.0); }"}
	shader.Tags = domain.ShaderTags{"a", "b"}

	for name, vite := range map[string]*views.Vite{"dev": views.NewDevVite("http://localhost:5173"), "build": builtVite(t)} {
		tmpl, err := views.NewTemplates("../../../web/templates", vite, false)
		if err != nil {
			t.Fatal(err)
		}

		for _, principal := range []*domain.Principal{nil, signedIn} {
			pages := map[string]struct {
				page *views.Page
				data any
			}{
				"gallery":  {tmpl.Gallery, galleryPageData{Principal: principal, Shaders: []domain.ShaderSummary{shader.ShaderSummary}}},
				"login":    {tmpl.Login, loginPageData{Principal: principal, Error: "bad login"}},
				"register": {tmpl.Register, registerPageData{Principal: principal, Error: "taken"}},
				"profile":  {tmpl.Profile, profilePageData{Principal: principal, User: signedIn.User}},
				"shader":   {tmpl.Shader, shaderPageData{Principal: principal, Shader: shader, CanEdit: principal != nil}},
				// only reachable signed in
				"me": {tmpl.Me, mePageData{Principal: signedIn, Error: "invalid"}},
			}
			for page, tc := range pages {
				t.Run(name+"/"+page, func(t *testing.T) {
					if err := tc.page.Execute(io.Discard, tc.data); err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

// builtVite uses a fake build manifest with the entries the templates load.
func builtVite(t *testing.T) *views.Vite {
	t.Helper()
	dir := t.TempDir()
	manifest := `{
		"src/site.ts": {"file": "assets/site-abc.js", "isEntry": true, "css": ["assets/site-abc.css"]},
		"src/editor.tsx": {"file": "assets/editor-abc.js", "isEntry": true}
	}`
	if err := os.MkdirAll(filepath.Join(dir, ".vite"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".vite", "manifest.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	vite, err := views.NewVite(dir)
	if err != nil {
		t.Fatal(err)
	}
	return vite
}
