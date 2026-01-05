import { useCallback } from "react";

import { backfillExercises, updateUserAdmin } from "../api";
import type { User, View } from "../types";

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
      } catch (err: any) {
        await notify(err.message || "Unable to update role");
      }
    },
    [currentUserId, setUsers, setView, notify],
  );

  // backfillCatalog triggers the exercise catalog backfill.
  const backfillCatalog = useCallback(async () => {
    try {
      await backfillExercises();
      await notify("Exercise catalog backfill complete.");
    } catch (err: any) {
      await notify(err.message || "Unable to backfill exercises");
    }
  }, [notify]);

  return { toggleAdmin, backfillCatalog };
}
