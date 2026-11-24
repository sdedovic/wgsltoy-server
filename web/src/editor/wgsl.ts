import { StreamLanguage, type StreamParser, type StringStream } from "@codemirror/language";

// https://www.w3.org/TR/WGSL/#keyword-summary
const keywords = new Set([
  "alias",
  "break",
  "case",
  "const",
  "const_assert",
  "continue",
  "continuing",
  "default",
  "diagnostic",
  "discard",
  "else",
  "enable",
  "fn",
  "for",
  "if",
  "let",
  "loop",
  "override",
  "requires",
  "return",
  "struct",
  "switch",
  "var",
  "while",
]);

const atoms = new Set(["true", "false"]);

const types = new Set([
  "bool",
  "f16",
  "f32",
  "i32",
  "u32",
  "array",
  "atomic",
  "ptr",
  "sampler",
  "sampler_comparison",
  "vec2",
  "vec3",
  "vec4",
  "mat2x2",
  "mat2x3",
  "mat2x4",
  "mat3x2",
  "mat3x3",
  "mat3x4",
  "mat4x2",
  "mat4x3",
  "mat4x4",
]);

// predeclared aliases such as vec3f, mat4x4h, and texture types such as texture_2d
const typePattern = /^(?:vec[234][fhiu]|mat[234]x[234][fh]|texture_\w+)$/;

// https://www.w3.org/TR/WGSL/#builtin-functions, the common ones
const builtins = new Set([
  "abs",
  "acos",
  "all",
  "any",
  "asin",
  "atan",
  "atan2",
  "ceil",
  "clamp",
  "cos",
  "cosh",
  "cross",
  "degrees",
  "determinant",
  "distance",
  "dot",
  "dpdx",
  "dpdy",
  "exp",
  "exp2",
  "faceForward",
  "floor",
  "fma",
  "fract",
  "fwidth",
  "inverseSqrt",
  "length",
  "log",
  "log2",
  "max",
  "min",
  "mix",
  "modf",
  "normalize",
  "pow",
  "radians",
  "reflect",
  "refract",
  "round",
  "saturate",
  "select",
  "sign",
  "sin",
  "sinh",
  "smoothstep",
  "sqrt",
  "step",
  "tan",
  "tanh",
  "transpose",
  "trunc",
  "textureSample",
  "textureSampleLevel",
  "textureLoad",
  "textureStore",
  "textureDimensions",
  "arrayLength",
  "bitcast",
]);

const inputs = new Set(["iResolution", "iTime", "iTimeDelta", "iFrame", "iMouse"]);

const number = /^(?:0[xX][\da-fA-F]+[iu]?|(?:\d+\.\d*|\.\d+|\d+)(?:[eE][+-]?\d+)?[fhiu]?)/;

interface State {
  /** depth of nested block comments, which WGSL allows */
  commentDepth: number;
}

// StringStream.match returns a boolean for strings and a match or null for patterns
function eat(stream: StringStream, pattern: string | RegExp): boolean {
  const matched = typeof pattern === "string" ? stream.match(pattern) : stream.match(pattern);
  return matched !== false && matched !== null;
}

function blockComment(stream: StringStream, state: State): string {
  while (!stream.eol()) {
    if (eat(stream, "/*")) {
      state.commentDepth += 1;
    } else if (eat(stream, "*/")) {
      state.commentDepth -= 1;
      if (state.commentDepth === 0) {
        break;
      }
    } else {
      stream.next();
    }
  }
  return "comment";
}

export const wgslParser: StreamParser<State> = {
  name: "wgsl",
  startState: () => ({ commentDepth: 0 }),
  copyState: (state) => ({ ...state }),
  token(stream, state) {
    if (state.commentDepth > 0) {
      return blockComment(stream, state);
    }
    if (stream.eatSpace()) {
      return null;
    }
    if (eat(stream, "//")) {
      stream.skipToEnd();
      return "comment";
    }
    if (eat(stream, "/*")) {
      state.commentDepth = 1;
      return blockComment(stream, state);
    }
    if (eat(stream, /^@\w+/)) {
      return "meta";
    }
    if (eat(stream, number)) {
      return "number";
    }
    const word = stream.match(/^[A-Za-z_]\w*/);
    if (Array.isArray(word)) {
      const name = word[0];
      if (keywords.has(name)) return "keyword";
      if (atoms.has(name)) return "atom";
      if (types.has(name) || typePattern.test(name)) return "typeName";
      if (builtins.has(name)) return "builtin";
      if (inputs.has(name)) return "variableName.special";
      return "variableName";
    }
    stream.next();
    return null;
  },
  languageData: {
    commentTokens: { line: "//", block: { open: "/*", close: "*/" } },
  },
};

export const wgsl = StreamLanguage.define(wgslParser);
