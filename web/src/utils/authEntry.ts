import type { View } from "../types";

// AuthEntry identifies the frontend gate to apply after loading public configuration.
export type AuthEntry = "setup" | "proxy-denied" | "ready";

// resolveAuthEntry prioritizes initial setup over external authentication failures.
// The backend still owns setup authorization and all subsequent identity checks.
export function resolveAuthEntry(
  config: { setupRequired: boolean; authHeaderEnabled: boolean },
  authError: string | null,
): AuthEntry {
  if (config.setupRequired) return "setup";
  if (config.authHeaderEnabled && authError) return "proxy-denied";
  return "ready";
}

// resolveAuthView shows the setup screen immediately, even when a saved URL
// points to another application view. The state hook then synchronizes the URL.
export function resolveAuthView(entry: AuthEntry, requestedView: View): View {
  return entry === "setup" ? "login" : requestedView;
}
