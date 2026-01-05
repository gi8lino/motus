import { useCallback } from "react";

import { createUser, loginUser } from "../api";
import type { User } from "../types";

// UseAuthActionsArgs configures login and registration behavior.
type UseAuthActionsArgs = {
  setLoginError: (message: string | null) => void;
  onLoginSuccess: (user: User) => void;
  onRegisterSuccess: (user: User) => void;
};

// useAuthActions provides login and registration handlers.
export function useAuthActions({
  setLoginError,
  onLoginSuccess,
  onRegisterSuccess,
}: UseAuthActionsArgs) {
  // login authenticates a user and forwards the result.
  const login = useCallback(
    async (email: string, password: string) => {
      try {
        setLoginError(null);
        const user = await loginUser(email, password);
        onLoginSuccess(user);
      } catch (err) {
        const message = err instanceof Error ? err.message : "Invalid login";
        setLoginError(message);
      }
    },
    [setLoginError, onLoginSuccess],
  );

  // register creates a new user and forwards the result.
  const register = useCallback(
    async (email: string, password: string) => {
      const user = await createUser(email, password);
      onRegisterSuccess(user);
    },
    [onRegisterSuccess],
  );

  return { login, register };
}
