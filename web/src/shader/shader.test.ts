import { describe, expect, it } from "vitest";
import { Clock } from "./clock.ts";
import { hasErrors, toDiagnostics } from "./diagnostics.ts";
import { INPUTS_BYTE_SIZE, packInputs } from "./inputs.ts";
import { Mouse } from "./mouse.ts";
import { buildProgram } from "./program.ts";

describe("packInputs", () => {
  it("matches the WGSL Inputs struct layout", () => {
    const buffer = new ArrayBuffer(INPUTS_BYTE_SIZE);
    packInputs(
      { resolution: [800, 600, 1], time: 1.5, timeDelta: 0.25, frame: 7, mouse: [1, 2, 3, 4] },
      buffer,
    );
    const f32 = new Float32Array(buffer);
    const u32 = new Uint32Array(buffer);
    expect([...f32.slice(0, 5)]).toEqual([800, 600, 1, 1.5, 0.25]);
    expect(u32[5]).toBe(7);
    expect([...f32.slice(8, 12)]).toEqual([1, 2, 3, 4]);
  });
});

describe("toDiagnostics", () => {
  const program = buildProgram("fn mainImage(p: vec2f) -> vec4f { return x; }");

  it("locates messages within the user's code", () => {
    const offset = program.code.indexOf("x;");
    const [diagnostic] = toDiagnostics(program, [
      { type: "error", message: " unresolved value 'x'\n", offset, length: 1 },
    ]);
    expect(diagnostic).toEqual({
      from: offset - program.userOffset,
      to: offset - program.userOffset + 1,
      severity: "error",
      message: "unresolved value 'x'",
    });
  });

  it("pins messages about the prelude to the start of the user's code", () => {
    const [diagnostic] = toDiagnostics(program, [
      { type: "error", message: "unresolved function 'mainImage'", offset: 10, length: 9 },
    ]);
    expect(diagnostic).toMatchObject({ from: 0, to: 0 });
  });

  it("clamps messages past the end", () => {
    const [diagnostic] = toDiagnostics(program, [
      { type: "warning", message: "w", offset: program.code.length, length: 5 },
    ]);
    expect(diagnostic).toMatchObject({ from: program.userLength, to: program.userLength });
  });

  it("only counts errors as failures", () => {
    const warning = { from: 0, to: 0, severity: "warning", message: "" } as const;
    expect(hasErrors([warning])).toBe(false);
    expect(hasErrors([warning, { ...warning, severity: "error" }])).toBe(true);
  });
});

describe("Clock", () => {
  it("advances while playing", () => {
    const clock = new Clock();
    expect(clock.tick(1000)).toEqual({ time: 0, timeDelta: 0, frame: 0 });
    expect(clock.tick(1500)).toEqual({ time: 0.5, timeDelta: 0.5, frame: 1 });
  });

  it("excludes time spent paused", () => {
    const clock = new Clock();
    clock.tick(0);
    clock.tick(1000);
    clock.pause();
    expect(clock.tick(5000)).toEqual({ time: 1, timeDelta: 0, frame: 2 });
    clock.play();
    clock.tick(9000);
    expect(clock.tick(9250)).toEqual({ time: 1.25, timeDelta: 0.25, frame: 3 });
  });

  it("restarts from zero on reset", () => {
    const clock = new Clock();
    clock.tick(0);
    clock.tick(1000);
    clock.reset();
    expect(clock.tick(2000)).toEqual({ time: 0, timeDelta: 0, frame: 0 });
  });
});

describe("Mouse", () => {
  it("tracks drags and negates the click position on release", () => {
    const mouse = new Mouse();
    mouse.move(5, 5);
    expect(mouse.value).toEqual([0, 0, -0, -0]);
    mouse.down(10, 20);
    mouse.move(30, 40);
    expect(mouse.value).toEqual([30, 40, 10, 20]);
    mouse.up();
    mouse.move(50, 60);
    expect(mouse.value).toEqual([30, 40, -10, -20]);
  });
});
