import { useEffect, useState, type RefObject } from "react";
import { ShaderRenderer, WebGpuUnavailableError } from "../../gpu/renderer.ts";

export type RendererState =
  | { readonly status: "starting" }
  | { readonly status: "ready"; readonly renderer: ShaderRenderer }
  | { readonly status: "unavailable"; readonly reason: string }
  | { readonly status: "failed"; readonly reason: string };

export function useShaderRenderer(canvas: RefObject<HTMLCanvasElement | null>): RendererState {
  const [state, setState] = useState<RendererState>({ status: "starting" });

  useEffect(() => {
    const element = canvas.current;
    if (element === null) {
      return;
    }
    const abort = new AbortController();
    let renderer: ShaderRenderer | undefined;

    ShaderRenderer.create(element, {
      signal: abort.signal,
      onDeviceLost: (reason) => {
        setState({ status: "failed", reason: `GPU device lost: ${reason}` });
      },
    }).then(
      (created) => {
        renderer = created;
        setState({ status: "ready", renderer: created });
      },
      (error: unknown) => {
        if (abort.signal.aborted) {
          return;
        }
        setState(
          error instanceof WebGpuUnavailableError
            ? { status: "unavailable", reason: error.message }
            : { status: "failed", reason: String(error) },
        );
      },
    );

    return () => {
      abort.abort();
      renderer?.dispose();
    };
  }, [canvas]);

  return state;
}
