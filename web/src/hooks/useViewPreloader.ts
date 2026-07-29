import { useCallback, useEffect, useState } from "react";
import type { User, View } from "../types";

export function useViewPreloader({
  preload,
  authenticated,
  currentUser,
}: {
  preload: (view: View) => void;
  authenticated: boolean;
  currentUser: User | null;
}) {
  const [prefetchedViews, setPrefetchedViews] = useState<ReadonlySet<View>>(
    () => new Set(),
  );
  const preloadTracked = useCallback(
    (view: View) => {
      preload(view);
      setPrefetchedViews((current) => new Set(current).add(view));
    },
    [preload],
  );

  useEffect(() => {
    if (!authenticated) return;
    const warm = () =>
      (
        [
          "workouts",
          "exercises",
          "history",
          "profile",
          ...(currentUser?.isAdmin ? (["admin"] as const) : []),
        ] as View[]
      ).forEach(preloadTracked);
    if ("requestIdleCallback" in window) {
      const id = window.requestIdleCallback(warm, { timeout: 1500 });
      return () => window.cancelIdleCallback(id);
    }
    const id = globalThis.setTimeout(warm, 250);
    return () => globalThis.clearTimeout(id);
  }, [authenticated, currentUser?.isAdmin, preloadTracked]);

  return { prefetchedViews, preloadView: preloadTracked, setPrefetchedViews };
}
