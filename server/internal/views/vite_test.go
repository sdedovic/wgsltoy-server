package views

import (
	"strings"
	"testing"
)

// the example from https://vite.dev/guide/backend-integration.html
var exampleManifest = map[string]manifestChunk{
	"_shared-B7PI925R.js": {File: "assets/shared-B7PI925R.js", CSS: []string{"assets/shared-ChJ_j-JJ.css"}},
	"views/foo.js": {
		File:    "assets/foo-BRBmoGS9.js",
		IsEntry: true,
		Imports: []string{"_shared-B7PI925R.js"},
		CSS:     []string{"assets/foo-5UjPuW-k.css"},
	},
	"baz.js": {File: "assets/baz-B2H3sXNv.js"},
}

func TestEntry_Build(t *testing.T) {
	vite := &Vite{manifest: exampleManifest}

	html, err := vite.Entry("views/foo.js")
	if err != nil {
		t.Fatal(err)
	}

	want := strings.Join([]string{
		`<link rel="stylesheet" href="/static/assets/foo-5UjPuW-k.css">`,
		`<link rel="stylesheet" href="/static/assets/shared-ChJ_j-JJ.css">`,
		`<script type="module" src="/static/assets/foo-BRBmoGS9.js"></script>`,
		`<link rel="modulepreload" href="/static/assets/shared-B7PI925R.js">`,
		``,
	}, "\n")
	if string(html) != want {
		t.Errorf("got\n%s\nwant\n%s", html, want)
	}
}

func TestEntry_BuildRejectsUnknownAndNonEntries(t *testing.T) {
	vite := &Vite{manifest: exampleManifest}
	for _, name := range []string{"views/missing.js", "baz.js"} {
		if _, err := vite.Entry(name); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestEntry_Dev(t *testing.T) {
	vite := NewDevVite("http://localhost:5173/")

	html, err := vite.Entry("src/editor.tsx")
	if err != nil {
		t.Fatal(err)
	}
	if want := `<script type="module" src="http://localhost:5173/static/src/editor.tsx"></script>`; string(html) != want {
		t.Errorf("got %s, want %s", html, want)
	}
	if !strings.Contains(string(vite.Client()), `src="http://localhost:5173/static/@vite/client"`) {
		t.Error("expected the vite client")
	}
}

func TestClient_BuildIsEmpty(t *testing.T) {
	if html := (&Vite{}).Client(); html != "" {
		t.Errorf("got %q", html)
	}
}
