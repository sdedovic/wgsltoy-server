import * as v from "valibot";
import { mountIslands } from "./islands/mount.tsx";
import { ShaderEditor } from "./islands/shader-editor/ShaderEditor.tsx";
import { DEFAULT_SHADER } from "./shader/program.ts";

// The island edits the code held by a form field, so saving the form saves the code.
mountIslands(
  "shader-editor",
  v.object({ source: v.pipe(v.string(), v.nonEmpty()) }),
  ({ source }) => {
    const field = document.getElementById(source);
    if (!(field instanceof HTMLTextAreaElement)) {
      throw new Error(`shader-editor: #${source} is not a textarea`);
    }
    if (field.value.trim() === "") {
      field.value = DEFAULT_SHADER;
    }
    return (
      <ShaderEditor
        initialCode={field.value}
        onChange={(code) => {
          field.value = code;
        }}
      />
    );
  },
);
