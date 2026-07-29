import { useEffect, useMemo, useState } from "react";
import type { ThemeMode } from "../types";
import { buildAppTheme } from "../theme";
import { resolveThemeMode } from "../utils/themeMode";

export function useThemeController() {
  const [themeMode, setThemeMode] = useState<ThemeMode>(() => {
    const stored = localStorage.getItem("motus:theme");
    return stored === "dark" || stored === "light" || stored === "auto"
      ? stored
      : "auto";
  });
  const [resolvedMode, setResolvedMode] = useState<"dark" | "light">(() =>
    resolveThemeMode(
      themeMode,
      window.matchMedia("(prefers-color-scheme: dark)").matches,
    ),
  );

  useEffect(() => {
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const apply = () => {
      const next = resolveThemeMode(themeMode, media.matches);
      document.documentElement.dataset.theme = next;
      setResolvedMode(next);
      localStorage.setItem("motus:theme", themeMode);
    };
    apply();
    if (themeMode !== "auto") return;
    media.addEventListener("change", apply);
    return () => media.removeEventListener("change", apply);
  }, [themeMode]);

  const theme = useMemo(() => buildAppTheme(resolvedMode), [resolvedMode]);
  return { theme, themeMode, setThemeMode };
}
