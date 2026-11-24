import * as v from "valibot";
import { mountIslands } from "./islands/mount.tsx";
import { ShaderPreview } from "./islands/shader-preview/ShaderPreview.tsx";

mountIslands(
  "shader-preview",
  v.object({ source: v.pipe(v.string(), v.nonEmpty()) }),
  ({ source }) => {
    const field = document.getElementById(source);
    if (!(field instanceof HTMLTextAreaElement)) {
      throw new Error(`shader-preview: #${source} is not a textarea`);
    }
    return <ShaderPreview code={field.value} />;
  },
);
