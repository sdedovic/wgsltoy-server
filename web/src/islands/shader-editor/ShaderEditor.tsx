import { ArrowCounterClockwiseIcon, HammerIcon, PauseIcon, PlayIcon } from "@phosphor-icons/react";
import { useCallback, useEffect, useRef, useState, type PointerEvent } from "react";
import { showDiagnostics } from "../../editor/extensions.ts";
import styles from "./ShaderEditor.module.css";
import { useCodeMirror } from "./use-code-mirror.ts";
import { usePageTop } from "./use-page-top.ts";
import { useShaderRenderer } from "./use-shader-renderer.ts";

export interface ShaderEditorProps {
  readonly initialCode: string;
  readonly onChange: (code: string) => void;
}

export function ShaderEditor({ initialCode, onChange }: ShaderEditorProps) {
  const rootRef = useRef<HTMLDivElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const editorParentRef = useRef<HTMLDivElement>(null);
  const gpu = useShaderRenderer(canvasRef);
  const renderer = gpu.status === "ready" ? gpu.renderer : undefined;
  const [playing, setPlaying] = useState(true);
  usePageTop(rootRef);

  const editorRef = useCodeMirror(editorParentRef, initialCode, {
    onCompile: () => void compile(),
    onChange,
  });

  const compile = useCallback(async () => {
    const view = editorRef.current;
    if (renderer === undefined || view === null) {
      return;
    }
    showDiagnostics(view, await renderer.compile(view.state.doc.toString()));
  }, [editorRef, renderer]);

  useEffect(() => {
    void compile();
  }, [compile]);

  const togglePlaying = () => {
    if (renderer === undefined) return;
    if (renderer.clock.playing) {
      renderer.clock.pause();
    } else {
      renderer.clock.play();
    }
    setPlaying(renderer.clock.playing);
  };

  // pointer positions are in CSS pixels, the shader works in device pixels
  const toDevicePixels = (event: PointerEvent<HTMLCanvasElement>): [number, number] => {
    const target = event.currentTarget;
    const rect = target.getBoundingClientRect();
    return [
      ((event.clientX - rect.left) * target.width) / rect.width,
      ((event.clientY - rect.top) * target.height) / rect.height,
    ];
  };

  return (
    <div ref={rootRef} className={styles.shaderEditor}>
      <div className={styles.preview}>
        <canvas
          ref={canvasRef}
          className={styles.canvas}
          hidden={gpu.status === "unavailable" || gpu.status === "failed"}
          onPointerDown={(e) => {
            e.currentTarget.setPointerCapture(e.pointerId);
            renderer?.mouse.down(...toDevicePixels(e));
          }}
          onPointerMove={(e) => renderer?.mouse.move(...toDevicePixels(e))}
          onPointerUp={() => renderer?.mouse.up()}
        />
        {gpu.status === "unavailable" && <NoWebGpu />}
        {gpu.status === "failed" && (
          <p role="alert" className={styles.notice}>
            The preview stopped working: {gpu.reason}
          </p>
        )}
        <div className={styles.toolbar}>
          <button
            type="button"
            onClick={togglePlaying}
            disabled={renderer === undefined}
            aria-label={playing ? "Pause" : "Play"}
          >
            {playing ? <PauseIcon aria-hidden /> : <PlayIcon aria-hidden />}
          </button>
          <button
            type="button"
            onClick={() => renderer?.clock.reset()}
            disabled={renderer === undefined}
            aria-label="Restart"
          >
            <ArrowCounterClockwiseIcon aria-hidden />
          </button>
        </div>
      </div>
      <div className={styles.code}>
        <div ref={editorParentRef} className={styles.editor} />
        <div className={styles.toolbar}>
          <button type="button" onClick={() => void compile()} disabled={renderer === undefined}>
            <HammerIcon aria-hidden /> Compile <kbd>Ctrl+Enter</kbd>
          </button>
        </div>
      </div>
    </div>
  );
}

function NoWebGpu() {
  return (
    <div role="alert" className={styles.notice}>
      <p>
        <strong>WebGPU is not supported on this device.</strong> It's a relatively new API and may
        not be available in your browser yet.
      </p>
      <p>
        For a WebGL based alternative, check out <a href="https://www.shadertoy.com/">Shadertoy</a>.
        To see which browsers support it, see the{" "}
        <a href="https://developer.mozilla.org/en-US/docs/Web/API/WebGPU_API#browser_compatibility">
          MDN WebGPU compatibility table
        </a>
        .
      </p>
    </div>
  );
}
