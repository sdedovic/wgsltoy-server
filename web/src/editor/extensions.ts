import { indentWithTab } from "@codemirror/commands";
import { lintGutter, setDiagnostics, type Diagnostic } from "@codemirror/lint";
import type { Extension } from "@codemirror/state";
import { EditorView, keymap } from "@codemirror/view";
import { basicSetup } from "codemirror";
import type { ShaderDiagnostic } from "../shader/diagnostics.ts";
import { wgsl } from "./wgsl.ts";

export interface EditorCallbacks {
  /** Mod-Enter */
  readonly onCompile: () => void;
  readonly onChange: (code: string) => void;
}

export function shaderEditorExtensions(callbacks: EditorCallbacks): Extension[] {
  return [
    basicSetup,
    wgsl,
    lintGutter(),
    keymap.of([
      {
        key: "Mod-Enter",
        preventDefault: true,
        run: () => {
          callbacks.onCompile();
          return true;
        },
      },
      indentWithTab,
    ]),
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
