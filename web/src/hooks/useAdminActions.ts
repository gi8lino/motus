import { useCallback } from "react";

import {
  backfillExercises,
  linkUserIdentity,
  revokeUserSessions,
  setUserLocalPassword,
  unlinkUserIdentity,
  updateUserAdmin,
} from "../api";
import type { AuthIdentity, User, View } from "../types";
import { MESSAGES, toErrorMessage } from "../utils/messages";
import { UI_TEXT } from "../utils/uiText";

// UseAdminActionsArgs wires user management actions.
type UseAdminActionsArgs = {
  currentUserId: string | null;
  setUsers: (updater: (prev: User[] | null) => User[] | null) => void;
  setView: (view: View) => void;
  notify: (message: string) => Promise<void>;
};

// useAdminActions provides admin-only user management handlers.
export function useAdminActions({
  currentUserId,
  setUsers,
  setView,
  notify,
}: UseAdminActionsArgs) {
  // toggleAdmin flips admin privileges for a user.
  const toggleAdmin = useCallback(
    async (user: User) => {
      try {
        // Persist role change before updating local cache.
        await updateUserAdmin(user.id, !user.isAdmin);
        setUsers(
          (prev) =>
            prev?.map((usr) =>
              usr.id === user.id ? { ...usr, isAdmin: !user.isAdmin } : usr,
            ) || prev,
        );
        if (user.id === currentUserId && !user.isAdmin) {
          setView("admin");
        }
      } catch (err) {
        await notify(toErrorMessage(err, MESSAGES.updateRoleFailed));
      }
    },
    [currentUserId, setUsers, setView, notify],
  );

  // backfillCatalog triggers the exercise catalog backfill.
  const backfillCatalog = useCallback(async () => {
    try {
      await backfillExercises();
      await notify(UI_TEXT.toasts.backfillComplete);
    } catch (err) {
      await notify(toErrorMessage(err, MESSAGES.backfillExercisesFailed));
    }
  }, [notify]);

  // revokeSessions invalidates local sessions for a selected account.
  const revokeSessions = useCallback(async (user: User) => {
    try {
      await revokeUserSessions(user.id);
      await notify(`Sessions revoked for ${user.name}.`);
    } catch (err) {
      await notify(toErrorMessage(err, "Unable to revoke sessions"));
    }
  }, [notify]);

  // setLocalPassword changes an administrator-managed local credential.
  const setLocalPassword = useCallback(async (user: User, password: string) => {
    try {
      await setUserLocalPassword(user.id, password);
      setUsers((prev) => prev?.map((item) => item.id === user.id ? { ...item, hasLocalPassword: Boolean(password) } : item) || prev);
      await notify(password ? `Local password set for ${user.name}.` : `Local password removed for ${user.name}.`);
    } catch (err) {
      await notify(toErrorMessage(err, "Unable to update local password"));
    }
  }, [notify, setUsers]);

  // linkIdentity attaches a provider identity and updates the local user cache.
  const linkIdentity = useCallback(async (user: User, identity: AuthIdentity) => {
    try {
      await linkUserIdentity(user.id, identity);
      setUsers((prev) => prev?.map((item) => item.id === user.id ? { ...item, identities: [...(item.identities || []), identity] } : item) || prev);
      await notify(`Provider identity linked to ${user.name}.`);
    } catch (err) {
      await notify(toErrorMessage(err, "Unable to link provider identity"));
    }
  }, [notify, setUsers]);

  // unlinkIdentity detaches a provider identity and updates the local user cache.
  const unlinkIdentity = useCallback(async (user: User, identity: AuthIdentity) => {
    try {
      await unlinkUserIdentity(user.id, identity.provider, identity.subject);
      setUsers((prev) => prev?.map((item) => item.id === user.id ? { ...item, identities: item.identities?.filter((value) => value.provider !== identity.provider || value.subject !== identity.subject) } : item) || prev);
      await notify(`Provider identity removed from ${user.name}.`);
    } catch (err) {
      await notify(toErrorMessage(err, "Unable to remove provider identity"));
    }
  }, [notify, setUsers]);

  return { toggleAdmin, backfillCatalog, revokeSessions, setLocalPassword, linkIdentity, unlinkIdentity };
}
