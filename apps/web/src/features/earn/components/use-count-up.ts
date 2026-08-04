"use client";

import { useEffect, useRef, useState } from "react";

/** Animated counter — eases from 0 (or previous value) to target over `duration`. */
export function useCountUp(value: number, duration = 700, enabled = true) {
  const [display, setDisplay] = useState(value || 0);
  const displayRef = useRef(display);
  displayRef.current = display;

  useEffect(() => {
    if (!enabled || !Number.isFinite(value)) {
      setDisplay(value || 0);
      return;
    }
    let frame = 0;
    const start = displayRef.current;
    const delta = value - start;
    const started = performance.now();
    const tick = (now: number) => {
      const progress = Math.min(1, (now - started) / duration);
      const eased = 1 - Math.pow(1 - progress, 3);
      setDisplay(start + delta * eased);
      if (progress < 1) frame = requestAnimationFrame(tick);
    };
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [value, duration, enabled]);

  return display;
}