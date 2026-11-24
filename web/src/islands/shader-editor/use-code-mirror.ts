import { EditorView } from "@codemirror/view";
import { useEffect, useLayoutEffect, useRef, type RefObject } from "react";
import { shaderEditorExtensions, type EditorCallbacks } from "../../editor/extensions.ts";

/**
 * Creates a CodeMirror editor in `parent`. Returns a ref, not state: the view is an imperative object
 * only used from event handlers and effects. The latest document survives the editor being recreated,
 * e.g. by Fast Refresh.
 */
export function useCodeMirror(
  parent: RefObject<HTMLElement | null>,
  initialDoc: string,
  callbacks: EditorCallbacks,
): RefObject<EditorView | null> {
  const viewRef = useRef<EditorView>(null);
  const docRef = useRef(initialDoc);
  const callbacksRef = useRef(callbacks);
  useLayoutEffect(() => {
    callbacksRef.current = callbacks;
  });

  useEffect(() => {
    const element = parent.current;
    if (element === null) {
      return;
    }
    const created = new EditorView({
      doc: docRef.current,
      parent: element,
      extensions: shaderEditorExtensions({
        onCompile: () => {
          callbacksRef.current.onCompile();
        },
        onChange: (code) => {
          docRef.current = code;
          callbacksRef.current.onChange(code);
        },
      }),
    });
    viewRef.current = created;
    return () => {
      created.destroy();
      viewRef.current = null;
    };
  }, [parent]);

  return viewRef;
}
