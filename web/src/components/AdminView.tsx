import { useState } from "react";
import type { User } from "../types";
import { UserForm } from "./AuthForms";

type AdminTab = "users" | "settings";

// AdminView manages users and admin access.
export function AdminView({
  users,
  loading,
  currentUserId,
  allowRegistration,
  onToggleAdmin,
  onCreateUser,
  onBackfill,
}: {
  users: User[];
  loading: boolean;
  currentUserId: string | null;
  allowRegistration: boolean;
  onToggleAdmin: (user: User) => void | Promise<void>;
  onCreateUser: (email: string, password: string) => void | Promise<void>;
  onBackfill: () => void | Promise<void>;
}) {
  // tab tracks the active admin section.
  const [tab, setTab] = useState<AdminTab>("users");
  // backfilling controls the backfill button state.
  const [backfilling, setBackfilling] = useState(false);

  // handleBackfill triggers the exercise catalog backfill.
  const handleBackfill = async () => {
    setBackfilling(true);
    try {
      await onBackfill();
    } finally {
      setBackfilling(false);
    }
  };

  return (
    <section className="panel">
      <div className="panel-header">
        <div>
          <h3>Admin</h3>
          <p className="muted small">Manage users and maintenance tasks.</p>
        </div>
      </div>
      <div className="profile-layout">
        <div className="profile-content">
          {tab === "users" && (
            <div className="stack">
              <div className="panel">
                <div className="panel-header">
                  <div>
                    <h3>Users</h3>
                    <p className="muted small">Manage roles and switch user.</p>
                  </div>
                </div>
                {loading && <p>Loading users…</p>}
                <ul className="list">
                  {users.map((u) => (
                    <li
                      key={u.id}
                      className={
                        currentUserId === u.id
                          ? "list-item active"
                          : "list-item"
                      }
                    >
                      <div className="list-row">
                        <div>
                          <strong>{u.name}</strong>
                          <div className="muted">
                            {u.isAdmin ? "Admin" : "User"} •{" "}
                            {new Date(u.createdAt).toLocaleDateString()}
                          </div>
                        </div>
                        <div className="btn-group">
                          <button
                            className="btn subtle"
                            onClick={() => onToggleAdmin(u)}
                          >
                            {u.isAdmin ? "Revoke Admin" : "Make Admin"}
                          </button>
                        </div>
                      </div>
                    </li>
                  ))}
                </ul>
              </div>
              <div className="panel">
                <h3>Create user</h3>
                {allowRegistration ? (
                  <UserForm onCreate={onCreateUser} />
                ) : (
                  <p className="muted small">
                    Registration is disabled in this environment.
                  </p>
                )}
              </div>
            </div>
          )}
          {tab === "settings" && (
            <div className="stack">
              <div className="panel">
                <div className="label">Maintenance</div>
                <p className="muted small">
                  Backfill promotes workouts exercises into the Core catalog.
                </p>
                <button
                  className="btn subtle"
                  type="button"
                  onClick={handleBackfill}
                  disabled={backfilling}
                >
                  {backfilling ? "Backfilling…" : "Backfill exercises"}
                </button>
              </div>
            </div>
          )}
        </div>
        <div className="profile-tabs">
          <button
            className={tab === "users" ? "tab active" : "tab"}
            onClick={() => setTab("users")}
          >
            Users
          </button>
          <button
            className={tab === "settings" ? "tab active" : "tab"}
            onClick={() => setTab("settings")}
          >
            Settings
          </button>
        </div>
      </div>
    </section>
  );
}
