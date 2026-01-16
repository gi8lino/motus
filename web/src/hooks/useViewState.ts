import { useEffect, useMemo, useState } from "react";
import type { View } from "../types";

const VIEW_PARAM = "view";

const viewOptions: View[] = [
  "train",
  "login",
  "workouts",
  "profile",
  "history",
  "exercises",
  "templates",
  "admin",
];

function isValidView(value: string): value is View {
  return viewOptions.includes(value as View);
}

function readViewFromURL(fallback: View): View {
  const params = new URLSearchParams(window.location.search);
  const raw = params.get(VIEW_PARAM);
  if (raw && isValidView(raw)) return raw;
  return fallback;
}

function writeViewToURL(view: View) {
  const params = new URLSearchParams(window.location.search);

  // Keep the URL clean for the default view.
  if (view === "train") {
    if (!params.has(VIEW_PARAM)) return;
    params.delete(VIEW_PARAM);
  } else {
    params.set(VIEW_PARAM, view);
  }

  const next = params.toString();
  const url = next
    ? `${window.location.pathname}?${next}${window.location.hash}`
    : `${window.location.pathname}${window.location.hash}`;

  window.history.replaceState({}, "", url);
}

/**
 * useViewState stores the current view in React state and mirrors it to the URL.
 * - Reads initial view from `?view=...`
 * - Writes changes back to `?view=...` (except "train", which removes the param)
 */
export function useViewState(defaultView: View) {
  const initial = useMemo(() => readViewFromURL(defaultView), [defaultView]);
  const [view, setView] = useState<View>(initial);

  useEffect(() => {
    writeViewToURL(view);
  }, [view]);

  return { view, setView };
}
