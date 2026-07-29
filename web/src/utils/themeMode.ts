import type { ThemeMode } from "../types";

export function resolveThemeMode(
  mode: ThemeMode,
  prefersDark: boolean,
): "dark" | "light" {
  return mode === "auto" ? (prefersDark ? "dark" : "light") : mode;
}
