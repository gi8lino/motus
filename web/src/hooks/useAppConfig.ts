import { useCallback, useEffect, useRef, useState } from "react";
import { getConfig, getCurrentUser } from "../api";
import { resolveAuthEntry } from "../utils/authEntry";
import { MESSAGES, toErrorMessage } from "../utils/messages";
import type { View } from "../types";

// AppConfig describes runtime settings exposed to the SPA.
export type AppConfig = {
  authHeaderEnabled: boolean;
  allowRegistration: boolean;
  oidcEnabled: boolean;
  localLoginEnabled: boolean;
  setupRequired: boolean;
  version: string;
  commit: string;
};

type UseAppConfigArgs = {
  view: View;
  setView: (view: View) => void;
  setCurrentUserId: (id: string | null) => void;
};

// useAppConfig loads config and resolves proxy-auth users when enabled.
export function useAppConfig({
  view,
  setView,
  setCurrentUserId,
}: UseAppConfigArgs) {
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [authError, setAuthError] = useState<string | null>(null);
  const viewRef = useRef<View>(view);

  useEffect(() => {
    viewRef.current = view;
  }, [view]);

  // Refresh after bootstrap so a stale setupRequired flag cannot keep the SPA
  // on setup after the first administrator has been persisted.
  const refreshConfig = useCallback(async () => {
    const cfg = await getConfig();
    setConfig(cfg);

    if (resolveAuthEntry(cfg, null) === "setup") {
      // Setup is public and precedes local, proxy, OIDC, or no-auth login.
      // Avoid requesting /api/me while the backend explicitly denies it.
      setCurrentUserId(null);
      setAuthError(null);
      setView("login");
      return;
    }

    try {
      const user = await getCurrentUser();
      setCurrentUserId(user.id);
      setAuthError(null);
      if (viewRef.current === "login") setView("training");
    } catch (err) {
      setCurrentUserId(null);
      setAuthError(
        cfg.authHeaderEnabled ? toErrorMessage(err, MESSAGES.authFailed) : null,
      );
    }
  }, [setCurrentUserId, setView]);

  useEffect(() => {
    void refreshConfig().catch((err: Error) => {
      setAuthError(toErrorMessage(err, MESSAGES.configFailed));
    });
  }, [refreshConfig]);

  return { config, authError, refreshConfig };
}
