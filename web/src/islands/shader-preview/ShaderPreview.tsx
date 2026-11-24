import { useEffect, useRef } from "react";
import { useShaderRenderer } from "../shader-editor/use-shader-renderer.ts";
import styles from "./ShaderPreview.module.css";

export interface ShaderPreviewProps {
  readonly code: string;
}

/** A small, non-interactive render of a shader. A shader that doesn't compile stays black. */
export function ShaderPreview({ code }: ShaderPreviewProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const gpu = useShaderRenderer(canvasRef);
  const renderer = gpu.status === "ready" ? gpu.renderer : undefined;

  useEffect(() => {
    void renderer?.compile(code);
  }, [renderer, code]);

  return <canvas ref={canvasRef} className={styles.canvas} />;
}
