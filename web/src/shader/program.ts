// A user shader is a `mainImage` function. The prelude provides everything around it: a fullscreen
// triangle, the uniform inputs, and the fragment entry point that calls `mainImage`.
const prelude = `struct Inputs {
  resolution: vec3f,
  time: f32,
  timeDelta: f32,
  frame: u32,
  mouse: vec4f,
}

@group(0) @binding(0) var<uniform> inputs: Inputs;

var<private> iResolution: vec3f;
var<private> iTime: f32;
var<private> iTimeDelta: f32;
var<private> iFrame: u32;
var<private> iMouse: vec4f;

@vertex
fn vertexMain(@builtin(vertex_index) index: u32) -> @builtin(position) vec4f {
  let corners = array(vec2f(-1.0, 3.0), vec2f(-1.0, -1.0), vec2f(3.0, -1.0));
  return vec4f(corners[index], 0.0, 1.0);
}

@fragment
fn fragmentMain(@builtin(position) fragCoord: vec4f) -> @location(0) vec4f {
  iResolution = inputs.resolution;
  iTime = inputs.time;
  iTimeDelta = inputs.timeDelta;
  iFrame = inputs.frame;
  iMouse = inputs.mouse;
  return mainImage(fragCoord.xy);
}

`;

export interface ShaderProgram {
  readonly code: string;
  /** where the user's code starts within `code`, in UTF-16 code units */
  readonly userOffset: number;
  readonly userLength: number;
}

export function buildProgram(userCode: string): ShaderProgram {
  return { code: prelude + userCode, userOffset: prelude.length, userLength: userCode.length };
}

export const VERTEX_ENTRY_POINT = "vertexMain";
export const FRAGMENT_ENTRY_POINT = "fragmentMain";

export const DEFAULT_SHADER = `// Shader inputs
// iResolution: vec3f  viewport resolution in pixels, and pixel aspect ratio
// iTime: f32          playback time in seconds
// iTimeDelta: f32     render time of the previous frame in seconds
// iFrame: u32         frames rendered
// iMouse: vec4f       xy: position while a button is down, zw: click position (negative when released)

fn mainImage(fragCoord: vec2f) -> vec4f {
  // normalized pixel coordinates, from 0 to 1
  let uv = fragCoord / iResolution.xy;

  // time varying pixel color
  let color = 0.5 + 0.5 * cos(iTime + uv.xyx + vec3f(0, 2, 4));

  return vec4f(color, 1.0);
}
`;
