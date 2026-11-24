import { StringStream } from "@codemirror/language";
import { describe, expect, it } from "vitest";
import { wgslParser } from "./wgsl.ts";

function tokenize(lines: string[]): [string, string | null][] {
  const state = wgslParser.startState?.(2) ?? { commentDepth: 0 };
  const tokens: [string, string | null][] = [];
  for (const line of lines) {
    const stream = new StringStream(line, 2, 2);
    while (!stream.eol()) {
      const style = wgslParser.token(stream, state);
      const text = stream.current();
      if (text.trim() !== "") {
        tokens.push([text, style]);
      }
      stream.start = stream.pos;
    }
  }
  return tokens;
}

describe("wgsl", () => {
  it("highlights declarations", () => {
    expect(tokenize(["@fragment fn main(p: vec2f) -> vec4<f32> { return vec4f(1.0); }"])).toEqual([
      ["@fragment", "meta"],
      ["fn", "keyword"],
      ["main", "variableName"],
      ["(", null],
      ["p", "variableName"],
      [":", null],
      ["vec2f", "typeName"],
      [")", null],
      ["-", null],
      [">", null],
      ["vec4", "typeName"],
      ["<", null],
      ["f32", "typeName"],
      [">", null],
      ["{", null],
      ["return", "keyword"],
      ["vec4f", "typeName"],
      ["(", null],
      ["1.0", "number"],
      [")", null],
      [";", null],
      ["}", null],
    ]);
  });

  it("highlights numbers with suffixes", () => {
    expect(tokenize(["0x1Fu 3i 1e-3f .5h"]).map(([, style]) => style)).toEqual([
      "number",
      "number",
      "number",
      "number",
    ]);
  });

  it("highlights builtins and shader inputs", () => {
    expect(tokenize(["sin(iTime)"])).toEqual([
      ["sin", "builtin"],
      ["(", null],
      ["iTime", "variableName.special"],
      [")", null],
    ]);
  });

  it("handles nested block comments across lines", () => {
    expect(tokenize(["/* a /* b */", "still */ let"])).toEqual([
      ["/* a /* b */", "comment"],
      ["still */", "comment"],
      ["let", "keyword"],
    ]);
  });
});
