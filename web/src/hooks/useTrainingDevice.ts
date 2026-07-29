import { useEffect, useRef } from "react";

type WakeLockSentinelLike = { release: () => Promise<void> };

export function shouldVibrate(enabled: boolean, supported: boolean): boolean {
  return enabled && supported;
}

export function useTrainingDevice(active: boolean, vibrationEnabled: boolean) {
  const wakeLock = useRef<WakeLockSentinelLike | null>(null);

  useEffect(() => {
    const acquire = async () => {
      if (!active || !("wakeLock" in navigator)) return;
      try {
        wakeLock.current = await (
          navigator as Navigator & {
            wakeLock: {
              request: (type: "screen") => Promise<WakeLockSentinelLike>;
            };
          }
        ).wakeLock.request("screen");
      } catch {
        // The browser may deny wake lock in low-power mode.
      }
    };
    void acquire();
    return () => {
      void wakeLock.current?.release();
      wakeLock.current = null;
    };
  }, [active]);

  useEffect(() => {
    if (
      active &&
      shouldVibrate(vibrationEnabled, typeof navigator.vibrate === "function")
    ) {
      navigator.vibrate(35);
    }
  }, [active, vibrationEnabled]);
}
