// The values a shader reads each frame, packed into one uniform buffer. The layout must match the
// `Inputs` struct in program.ts: WGSL aligns vec3f and vec4f to 16 bytes.
export interface Inputs {
  /** width and height in device pixels, and pixel aspect ratio */
  readonly resolution: readonly [number, number, number];
  /** seconds since start, excluding time spent paused */
  readonly time: number;
  /** seconds since the previous frame */
  readonly timeDelta: number;
  /** frames rendered since start */
  readonly frame: number;
  /** see Mouse */
  readonly mouse: readonly [number, number, number, number];
}

export const INPUTS_BYTE_SIZE = 48;

export function packInputs(inputs: Inputs, buffer: ArrayBuffer): void {
  const f32 = new Float32Array(buffer);
  const u32 = new Uint32Array(buffer);
  f32.set(inputs.resolution, 0); // offset 0
  f32[3] = inputs.time; // offset 12
  f32[4] = inputs.timeDelta; // offset 16
  u32[5] = inputs.frame; // offset 20
  f32.set(inputs.mouse, 8); // offset 32
}
