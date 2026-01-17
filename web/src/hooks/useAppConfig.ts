import { useEffect, useState } from "react";
import { getConfig, getCurrentUser, setAuthHeaderEnabled } from "../api";
import { MESSAGES, toErrorMessage } from "../utils/messages";
import type { View } from "../types";

// AppConfig describes runtime settings exposed to the SPA.
export type AppConfig = {
  authHeaderEnabled: boolean;
  allowRegistration: boolean;
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

  useEffect(() => {
    getConfig()
      .then((cfg) => {
        setConfig(cfg);
        setAuthHeaderEnabled(cfg.authHeaderEnabled);

        if (!cfg.authHeaderEnabled) return;

        return getCurrentUser()
          .then((user) => {
            setCurrentUserId(user.id);
            setAuthError(null);
            if (view === "login") setView("train");
          })
          .catch((err: Error) => {
            setAuthError(toErrorMessage(err, MESSAGES.authFailed));
          });
      })
      .catch((err: Error) => {
        setAuthError(toErrorMessage(err, MESSAGES.configFailed));
      });
  }, [setCurrentUserId, setView, view]);

  return { config, authError };
}
