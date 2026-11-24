import { useLayoutEffect, type RefObject } from "react";

/**
 * Keeps `--page-top` on `element` set to its distance from the top of the page, so CSS can size it to
 * end at a point in the viewport. Content above it can rewrap, so it's measured again on every resize.
 */
export function usePageTop(element: RefObject<HTMLElement | null>): void {
  useLayoutEffect(() => {
    const target = element.current;
    if (target === null) {
      return;
    }
    const measure = () => {
      const top = target.getBoundingClientRect().top + window.scrollY;
      target.style.setProperty("--page-top", `${String(Math.round(top))}px`);
    };
    measure();
    window.addEventListener("resize", measure);
    return () => {
      window.removeEventListener("resize", measure);
    };
  }, [element]);
}
