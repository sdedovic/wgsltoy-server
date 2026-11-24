import { StrictMode, type ReactNode } from "react";
import { createRoot } from "react-dom/client";
import * as v from "valibot";

type DatasetSchema = v.GenericSchema<Record<string, string | undefined>, unknown>;

/**
 * Mounts a React island into every `<div data-island="name" data-*>` on the page. The element's other
 * data attributes are its props, validated by `props`, since the server wrote them as untyped strings.
 */
export function mountIslands<S extends DatasetSchema>(
  name: string,
  props: S,
  render: (props: v.InferOutput<S>) => ReactNode,
): void {
  for (const element of document.querySelectorAll<HTMLElement>(`[data-island="${name}"]`)) {
    const dataset = Object.fromEntries(
      Object.entries(element.dataset).filter(([key]) => key !== "island"),
    );
    const parsed = v.safeParse(props, dataset);
    if (!parsed.success) {
      throw new Error(`island ${name}: ${v.summarize(parsed.issues)}`);
    }
    createRoot(element).render(<StrictMode>{render(parsed.output)}</StrictMode>);
  }
}
