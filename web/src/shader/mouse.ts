/**
 * Trying to copy Shadertoy's iMouse, in device pixels from the top left. xy is the position while a button is down
 * and stays at the last position after release. zw is where the button went down, negated after release.
 */
export class Mouse {
  #position: readonly [number, number] = [0, 0];
  #click: readonly [number, number] = [0, 0];
  #down = false;

  down(x: number, y: number): void {
    this.#down = true;
    this.#position = [x, y];
    this.#click = [x, y];
  }

  move(x: number, y: number): void {
    if (this.#down) {
      this.#position = [x, y];
    }
  }

  up(): void {
    this.#down = false;
  }

  get value(): readonly [number, number, number, number] {
    const [x, y] = this.#position;
    const [cx, cy] = this.#click;
    return this.#down ? [x, y, cx, cy] : [x, y, -Math.abs(cx), -Math.abs(cy)];
  }
}
