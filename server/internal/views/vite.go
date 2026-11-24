package views

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// AssetBase is the URL path the built assets are served under. It must match `base` in ../web/vite.config.ts.
const AssetBase = "/static/"

// Vite renders the tags that load a Vite entry, e.g. "src/editor.tsx", following
// https://vite.dev/guide/backend-integration.html.
//
// With a dev server URL, pages load entries from the dev server, which provides HMR. Otherwise entries
// are resolved to their hashed files through the build manifest.
type Vite struct {
	devServer string
	manifest  map[string]manifestChunk
}

type manifestChunk struct {
	File    string   `json:"file"`
	CSS     []string `json:"css"`
	Imports []string `json:"imports"`
	IsEntry bool     `json:"isEntry"`
}

// NewDevVite loads entries from the dev server, which serves them under AssetBase too.
func NewDevVite(devServer string) *Vite {
	return &Vite{devServer: strings.TrimSuffix(devServer, "/") + AssetBase}
}

// NewVite reads the manifest written by `vite build` into distDir.
func NewVite(distDir string) (*Vite, error) {
	path := filepath.Join(distDir, ".vite", "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading vite manifest, was the frontend built? %w", err)
	}
	var manifest map[string]manifestChunk
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &Vite{manifest: manifest}, nil
}

func (v *Vite) FuncMap() template.FuncMap {
	return template.FuncMap{
		"viteClient": v.Client,
		"vite":       v.Entry,
	}
}

// Client renders the tags every page needs once, before any entry. Only the dev server needs any: its
// client for HMR, and the React Fast Refresh preamble, which @vitejs/plugin-react can't inject into HTML
// it doesn't serve.
func (v *Vite) Client() template.HTML {
	if v.devServer == "" {
		return ""
	}
	// devServer comes from config, not from users
	return template.HTML(fmt.Sprintf(`<script type="module">
import RefreshRuntime from "%[1]s@react-refresh";
RefreshRuntime.injectIntoGlobalHook(window);
window.$RefreshReg$ = () => {};
window.$RefreshSig$ = () => (type) => type;
window.__vite_plugin_react_preamble_installed__ = true;
</script>
<script type="module" src="%[1]s@vite/client"></script>`, v.devServer))
}

// Entry renders the tags loading one entry. An unknown entry fails the render, rather than the page
// silently missing its scripts.
func (v *Vite) Entry(name string) (template.HTML, error) {
	if v.devServer != "" {
		return template.HTML(fmt.Sprintf(`<script type="module" src="%s%s"></script>`,
			v.devServer, template.HTMLEscapeString(name))), nil
	}

	entry, ok := v.manifest[name]
	if !ok || !entry.IsEntry {
		return "", fmt.Errorf("vite entry %q not in manifest", name)
	}
	imported := v.importedChunks(entry)

	var b strings.Builder
	for _, css := range entry.CSS {
		writeTag(&b, `<link rel="stylesheet" href="%s">`, css)
	}
	for _, chunk := range imported {
		for _, css := range chunk.CSS {
			writeTag(&b, `<link rel="stylesheet" href="%s">`, css)
		}
	}
	writeTag(&b, `<script type="module" src="%s"></script>`, entry.File)
	for _, chunk := range imported {
		writeTag(&b, `<link rel="modulepreload" href="%s">`, chunk.File)
	}
	// file names come from the build manifest, and are escaped anyway
	return template.HTML(b.String()), nil
}

// importedChunks returns the chunks entry statically imports, transitively, dependencies first.
func (v *Vite) importedChunks(entry manifestChunk) []manifestChunk {
	seen := map[string]bool{}
	var chunks []manifestChunk
	var visit func(manifestChunk)
	visit = func(chunk manifestChunk) {
		for _, name := range chunk.Imports {
			if seen[name] {
				continue
			}
			seen[name] = true
			imported, ok := v.manifest[name]
			if !ok {
				continue
			}
			visit(imported)
			chunks = append(chunks, imported)
		}
	}
	visit(entry)
	return chunks
}

func writeTag(b *strings.Builder, format, file string) {
	_, _ = fmt.Fprintf(b, format, template.HTMLEscapeString(AssetBase+file))
	b.WriteByte('\n')
}
