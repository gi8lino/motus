import type { User } from "../types";
import { UserForm } from "./AuthForms";

// AdminView manages users and admin access.
export function AdminView({
  users,
  loading,
  currentUserId,
  currentUserName,
  authHeaderEnabled,
  allowRegistration,
  onSelectUser,
  onToggleAdmin,
  onCreateUser,
}: {
  users: User[];
  loading: boolean;
  currentUserId: string | null;
  currentUserName: string;
  authHeaderEnabled: boolean;
  allowRegistration: boolean;
  onSelectUser: (id: string) => void;
  onToggleAdmin: (user: User) => void | Promise<void>;
  onCreateUser: (email: string, password: string) => void | Promise<void>;
}) {
  return (
    <section className="grid two">
      <div className="panel">
        <div className="panel-header">
          <div>
            <h3>Users</h3>
            <p className="muted small">Manage roles and switch user.</p>
          </div>
          {!authHeaderEnabled && (
            <UserSelect
              users={users}
              value={currentUserId}
              onChange={onSelectUser}
            />
          )}
          {authHeaderEnabled && (
            <div className="muted small">{currentUserName}</div>
          )}
        </div>
        {loading && <p>Loading users…</p>}
        <ul className="list">
          {users.map((u) => (
            <li
              key={u.id}
              className={
                currentUserId === u.id ? "list-item active" : "list-item"
              }
            >
              <div className="list-row">
                <div
                  onClick={() => {
                    if (!authHeaderEnabled) {
                      onSelectUser(u.id);
                    }
                  }}
                  style={{
                    cursor: authHeaderEnabled ? "default" : "pointer",
                  }}
                >
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
    </section>
  );
}

// UserSelect renders a dropdown of users.
function UserSelect({
  users,
  value,
  onChange,
}: {
  users: User[];
  value: string | null;
  onChange: (id: string) => void;
}) {
  return (
    <select value={value || ""} onChange={(e) => onChange(e.target.value)}>
      <option value="">Select user</option>
      {users.map((u) => (
        <option key={u.id} value={u.id}>
          {u.name}
        </option>
      ))}
    </select>
  );
}
