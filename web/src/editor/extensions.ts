import { indentWithTab } from "@codemirror/commands";
import { lintGutter, setDiagnostics, type Diagnostic } from "@codemirror/lint";
import { Prec, type Extension } from "@codemirror/state";
import { EditorView, keymap } from "@codemirror/view";
import { basicSetup } from "codemirror";
import type { ShaderDiagnostic } from "../shader/diagnostics.ts";
import { wgsl } from "./wgsl.ts";

export interface EditorCallbacks {
  /** Mod-Enter */
  readonly onCompile: () => void;
  readonly onChange: (code: string) => void;
}

// The WGSL language has no folding, so the fold gutter would always be empty. The lint gutter only shows
// while there's something to mark, so without errors the line numbers are the whole gutter. Gutters are
// `display: flex !important`, so they're collapsed to no width instead of hidden.
const compactGutters = EditorView.theme({
  ".cm-foldGutter": { width: "0" },
  ".cm-gutter-lint:not(:has(.cm-lint-marker))": { width: "0" },
});

export function shaderEditorExtensions(callbacks: EditorCallbacks): Extension[] {
  return [
    basicSetup,
    wgsl,
    lintGutter(),
    compactGutters,
    // basicSetup binds Mod-Enter to insert a blank line, and earlier keymaps win, so this one is raised
    Prec.highest(
      keymap.of([
        {
          key: "Mod-Enter",
          preventDefault: true,
          run: () => {
            callbacks.onCompile();
            return true;
          },
        },
      ]),
    ),
    keymap.of([indentWithTab]),
    EditorView.updateListener.of((update) => {
      if (update.docChanged) {
        callbacks.onChange(update.state.doc.toString());
      }
    }),
  ];
}

// Compiling is async, so the document may have shrunk since; out of range positions would throw.
export function showDiagnostics(view: EditorView, diagnostics: readonly ShaderDiagnostic[]): void {
  const length = view.state.doc.length;
  const converted: Diagnostic[] = diagnostics.map((d) => ({
    ...d,
    from: Math.min(d.from, length),
    to: Math.min(d.to, length),
    source: "WGSL",
  }));
  view.dispatch(setDiagnostics(view.state, converted));
}
