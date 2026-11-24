export interface Tick {
  readonly time: number;
  readonly timeDelta: number;
  readonly frame: number;
}

/** Playback time for a shader, in seconds. Time doesn't advance while paused. */
export class Clock {
  #elapsedMs = 0;
  #lastMs: number | undefined;
  #frame = 0;
  #playing = true;

  get playing(): boolean {
    return this.#playing;
  }

  play(): void {
    this.#playing = true;
  }

  pause(): void {
    this.#playing = false;
    this.#lastMs = undefined;
  }

  reset(): void {
    this.#elapsedMs = 0;
    this.#lastMs = undefined;
    this.#frame = 0;
  }

  /** Advances to `nowMs`, e.g. a requestAnimationFrame timestamp. */
  tick(nowMs: number): Tick {
    const deltaMs = this.#playing && this.#lastMs !== undefined ? nowMs - this.#lastMs : 0;
    this.#lastMs = this.#playing ? nowMs : undefined;
    this.#elapsedMs += deltaMs;
    const tick = { time: this.#elapsedMs / 1000, timeDelta: deltaMs / 1000, frame: this.#frame };
    if (this.#playing) {
      this.#frame += 1;
    }
    return tick;
  }
}
