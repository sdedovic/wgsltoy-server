import type { ShaderProgram } from "./program.ts";

export type Severity = "error" | "warning" | "info";

/** A compiler message located within the user's code. */
export interface ShaderDiagnostic {
  readonly from: number;
  readonly to: number;
  readonly severity: Severity;
  readonly message: string;
}

export type CompilationMessage = Pick<
  GPUCompilationMessage,
  "type" | "message" | "offset" | "length"
>;

// Messages about the prelude, e.g. a missing or mistyped `mainImage`, are pinned to the start of the
// user's code since that's where the fix is.
export function toDiagnostics(
  program: ShaderProgram,
  messages: readonly CompilationMessage[],
): ShaderDiagnostic[] {
  return messages.map((m) => {
    const start = m.offset - program.userOffset;
    const from = clamp(start, 0, program.userLength);
    const to = start < 0 ? from : clamp(start + m.length, from, program.userLength);
    return { from, to, severity: m.type, message: m.message.trim() };
  });
}

export function hasErrors(diagnostics: readonly ShaderDiagnostic[]): boolean {
  return diagnostics.some((d) => d.severity === "error");
}

function clamp(n: number, min: number, max: number): number {
  return Math.min(Math.max(n, min), max);
}
