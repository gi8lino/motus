import { useState } from "react";
import { Button, Stack, TextField } from "@mui/material";
import type { User } from "../../types";
import { UserForm } from "../auth/AuthForm";
import { UI_TEXT } from "../../utils/uiText";

type AdminTab = "users" | "settings";

export type AdminViewData = {
  users: User[];
  loading: boolean;
  currentUserId: string | null;
  allowRegistration: boolean;
};

export type AdminViewActions = {
  onToggleAdmin: (user: User) => void | Promise<void>;
  onRevokeSessions: (user: User) => void | Promise<void>;
  onSetLocalPassword: (user: User, password: string) => void | Promise<void>;
  onCreateUser: (email: string, password: string) => void | Promise<void>;
  onBackfill: () => void | Promise<void>;
};

// AdminView manages users and admin access.
export function AdminView({
  data,
  actions,
}: {
  data: AdminViewData;
  actions: AdminViewActions;
}) {
  const { users, loading, currentUserId, allowRegistration } = data;
  const {
    onToggleAdmin,
    onRevokeSessions,
    onSetLocalPassword,
    onCreateUser,
    onBackfill,
  } = actions;
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
          <h3>{UI_TEXT.pages.admin.title}</h3>
          <p className="muted small hint">{UI_TEXT.pages.admin.hint}</p>
        </div>
      </div>
      <div className="profile-layout">
        <div className="profile-content">
          {/* Tab content */}
          {tab === "users" && (
            <div className="stack">
              <div className="panel">
                <div className="panel-header">
                  <div>
                    <h3>{UI_TEXT.pages.admin.usersTitle}</h3>
                    <p className="muted small hint">
                      {UI_TEXT.pages.admin.usersHint}
                    </p>
                  </div>
                </div>
                {loading && <p>Loading users…</p>}
                {/* User list */}
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
                            {u.isAdmin
                              ? UI_TEXT.roles.admin
                              : UI_TEXT.roles.user}{" "}
                            • {new Date(u.createdAt).toLocaleDateString()}
                          </div>
                          <div className="muted small">
                            <UserAuthenticationSummary user={u} />
                          </div>
                        </div>
                        <div className="btn-group">
                          <button
                            className="btn subtle"
                            onClick={() => onToggleAdmin(u)}
                          >
                            {u.isAdmin
                              ? UI_TEXT.admin.revokeAdmin
                              : UI_TEXT.admin.makeAdmin}
                          </button>
                          <button
                            className="btn subtle"
                            onClick={() => onRevokeSessions(u)}
                          >
                            Revoke sessions
                          </button>
                        </div>
                      </div>
                      <UserAccessControls
                        user={u}
                        onSetLocalPassword={onSetLocalPassword}
                      />
                    </li>
                  ))}
                </ul>
              </div>
              <div className="panel">
                <h3>{UI_TEXT.pages.admin.createUserTitle}</h3>
                {allowRegistration ? (
                  <UserForm onCreate={onCreateUser} />
                ) : (
                  <p className="muted small hint">
                    {UI_TEXT.pages.auth.registrationDisabledHint}
                  </p>
                )}
              </div>
            </div>
          )}
          {tab === "settings" && (
            <div className="stack">
              <div className="panel">
                <div className="label">
                  {UI_TEXT.pages.admin.maintenanceLabel}
                </div>
                <p className="muted small hint">
                  {UI_TEXT.pages.admin.backfillHint}
                </p>
                <button
                  className="btn subtle"
                  type="button"
                  onClick={handleBackfill}
                  disabled={backfilling}
                >
                  {backfilling
                    ? UI_TEXT.admin.backfill.working
                    : UI_TEXT.admin.backfill.action}
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

// UserAuthenticationSummary presents the account's configured sign-in methods.
function UserAuthenticationSummary({ user }: { user: User }) {
  const methods = [
    ...(user.hasLocalPassword ? ["Local password"] : []),
    ...new Set(
      user.identities?.map((identity) =>
        authenticationProvider(identity.provider),
      ) || [],
    ),
  ];

  return methods.length > 0
    ? `Authentication: ${methods.join(" · ")}`
    : "No sign-in method configured";
}

// authenticationProvider translates persisted provider keys into user-facing labels.
function authenticationProvider(provider: string): string {
  if (provider === "trusted-proxy") return "Trusted proxy";
  if (provider.startsWith("oidc:")) return "OpenID Connect";
  return provider;
}

// UserAccessControls manages the local recovery credential for one user.
function UserAccessControls({
  user,
  onSetLocalPassword,
}: {
  user: User;
  onSetLocalPassword: (user: User, password: string) => void | Promise<void>;
}) {
  const [password, setPassword] = useState("");
  return (
    <Stack spacing={1} sx={{ mt: 1.5, maxWidth: 620 }}>
      <Stack direction="row" spacing={1}>
        <TextField
          size="small"
          label="Local password"
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
        />
        <Button
          size="small"
          onClick={() => {
            onSetLocalPassword(user, password);
            setPassword("");
          }}
        >
          {user.hasLocalPassword ? "Replace" : "Set"}
        </Button>
        {user.hasLocalPassword ? (
          <Button
            size="small"
            color="warning"
            onClick={() => onSetLocalPassword(user, "")}
          >
            Remove
          </Button>
        ) : null}
      </Stack>
    </Stack>
  );
}
