import { Clock } from "../shader/clock.ts";
import { hasErrors, toDiagnostics, type ShaderDiagnostic } from "../shader/diagnostics.ts";
import { INPUTS_BYTE_SIZE, packInputs } from "../shader/inputs.ts";
import { Mouse } from "../shader/mouse.ts";
import { buildProgram, FRAGMENT_ENTRY_POINT, VERTEX_ENTRY_POINT } from "../shader/program.ts";

export class WebGpuUnavailableError extends Error {
  override name = "WebGpuUnavailableError";
}

export interface RendererOptions {
  /** aborts creation, e.g. when the component that started it unmounts */
  readonly signal?: AbortSignal;
  /** called once if the GPU device is lost, e.g. after a driver reset */
  readonly onDeviceLost?: (message: string) => void;
}

/** Renders a user shader to a canvas, every animation frame, with the inputs from Inputs. */
export class ShaderRenderer {
  readonly clock = new Clock();
  readonly mouse = new Mouse();

  readonly #device: GPUDevice;
  readonly #context: GPUCanvasContext;
  readonly #canvas: HTMLCanvasElement;
  readonly #format: GPUTextureFormat;
  readonly #layout: GPUPipelineLayout;
  readonly #bindGroup: GPUBindGroup;
  readonly #inputsBuffer: GPUBuffer;
  readonly #inputs = new ArrayBuffer(INPUTS_BYTE_SIZE);
  readonly #resizeObserver: ResizeObserver;
  #pipeline: GPURenderPipeline | undefined;
  #frameRequest: number | undefined;
  #disposed = false;

  static async create(
    canvas: HTMLCanvasElement,
    options: RendererOptions = {},
  ): Promise<ShaderRenderer> {
    if (!("gpu" in navigator)) {
      throw new WebGpuUnavailableError("WebGPU is not supported by this browser");
    }
    const adapter = await navigator.gpu.requestAdapter();
    if (adapter === null) {
      throw new WebGpuUnavailableError("no suitable GPU adapter found");
    }
    const device = await adapter.requestDevice();
    // checked before configuring the canvas, which a newer renderer may already own
    if (options.signal?.aborted === true) {
      device.destroy();
      throw options.signal.reason;
    }
    const context = canvas.getContext("webgpu");
    if (context === null) {
      device.destroy();
      throw new WebGpuUnavailableError("could not create a WebGPU canvas context");
    }
    return new ShaderRenderer(canvas, device, context, options);
  }

  private constructor(
    canvas: HTMLCanvasElement,
    device: GPUDevice,
    context: GPUCanvasContext,
    options: RendererOptions,
  ) {
    this.#canvas = canvas;
    this.#device = device;
    this.#context = context;
    this.#format = navigator.gpu.getPreferredCanvasFormat();
    context.configure({ device, format: this.#format, alphaMode: "premultiplied" });

    this.#inputsBuffer = device.createBuffer({
      size: INPUTS_BYTE_SIZE,
      usage: GPUBufferUsage.UNIFORM | GPUBufferUsage.COPY_DST,
    });
    const bindGroupLayout = device.createBindGroupLayout({
      entries: [{ binding: 0, visibility: GPUShaderStage.FRAGMENT, buffer: {} }],
    });
    this.#layout = device.createPipelineLayout({ bindGroupLayouts: [bindGroupLayout] });
    this.#bindGroup = device.createBindGroup({
      layout: bindGroupLayout,
      entries: [{ binding: 0, resource: { buffer: this.#inputsBuffer } }],
    });

    void device.lost.then((info) => {
      if (!this.#disposed) {
        options.onDeviceLost?.(info.message);
      }
    });

    this.#resizeObserver = new ResizeObserver(([entry]) => {
      if (entry !== undefined) {
        this.#resize(entry);
      }
    });
    try {
      this.#resizeObserver.observe(canvas, { box: "device-pixel-content-box" });
    } catch {
      // Safari doesn't support device pixel boxes, #resize falls back to the CSS box
      this.#resizeObserver.observe(canvas);
    }
  }

  /** Compiles `userCode` and, if it has no errors, renders it from the next frame on. */
  async compile(userCode: string): Promise<ShaderDiagnostic[]> {
    const program = buildProgram(userCode);
    const module = this.#device.createShaderModule({ code: program.code });
    const info = await module.getCompilationInfo();
    const diagnostics = toDiagnostics(program, info.messages);
    if (hasErrors(diagnostics) || this.#disposed) {
      return diagnostics;
    }

    this.#pipeline = await this.#device.createRenderPipelineAsync({
      layout: this.#layout,
      vertex: { module, entryPoint: VERTEX_ENTRY_POINT },
      fragment: { module, entryPoint: FRAGMENT_ENTRY_POINT, targets: [{ format: this.#format }] },
      primitive: { topology: "triangle-list" },
    });
    this.#requestFrame();
    return diagnostics;
  }

  dispose(): void {
    this.#disposed = true;
    if (this.#frameRequest !== undefined) {
      cancelAnimationFrame(this.#frameRequest);
    }
    this.#resizeObserver.disconnect();
    this.#device.destroy();
  }

  #requestFrame(): void {
    if (this.#frameRequest === undefined && !this.#disposed) {
      this.#frameRequest = requestAnimationFrame((now) => {
        this.#frameRequest = undefined;
        this.#render(now);
        this.#requestFrame();
      });
    }
  }

  #render(nowMs: number): void {
    if (this.#pipeline === undefined) {
      return;
    }
    const { width, height } = this.#canvas;
    const tick = this.clock.tick(nowMs);
    packInputs({ resolution: [width, height, 1], ...tick, mouse: this.mouse.value }, this.#inputs);
    this.#device.queue.writeBuffer(this.#inputsBuffer, 0, this.#inputs);

    const encoder = this.#device.createCommandEncoder();
    const pass = encoder.beginRenderPass({
      colorAttachments: [
        {
          view: this.#context.getCurrentTexture().createView(),
          clearValue: { r: 0, g: 0, b: 0, a: 1 },
          loadOp: "clear",
          storeOp: "store",
        },
      ],
    });
    pass.setPipeline(this.#pipeline);
    pass.setBindGroup(0, this.#bindGroup);
    pass.draw(3);
    pass.end();
    this.#device.queue.submit([encoder.finish()]);
  }

  #resize(entry: ResizeObserverEntry): void {
    const max = this.#device.limits.maxTextureDimension2D;
    const size = entry.devicePixelContentBoxSize[0];
    const fallback = entry.contentBoxSize[0];
    const width = size?.inlineSize ?? (fallback?.inlineSize ?? 0) * devicePixelRatio;
    const height = size?.blockSize ?? (fallback?.blockSize ?? 0) * devicePixelRatio;
    this.#canvas.width = Math.max(1, Math.min(Math.round(width), max));
    this.#canvas.height = Math.max(1, Math.min(Math.round(height), max));
  }
}
