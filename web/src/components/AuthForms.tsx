import { useState } from "react";

import { isValidEmail } from "../utils/validation";

// UserForm creates a new user.
export function UserForm({
  onCreate,
}: {
  onCreate: (email: string, password: string) => void;
}) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const trimmedEmail = email.trim();
  const emailInvalid = trimmedEmail !== "" && !isValidEmail(trimmedEmail);
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        // Guard: require valid credentials before submit.
        if (!email.trim() || !password.trim()) return;
        if (!isValidEmail(email.trim())) return;
        onCreate(email.trim(), password.trim());
        setEmail("");
        setPassword("");
      }}
      className="stack"
    >
      <div className="field">
        <label>Email</label>
        <input
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="you@example.com"
          className={emailInvalid ? "input-error" : undefined}
          required
        />
        {emailInvalid && (
          <div className="helper error">Enter a valid email address</div>
        )}
      </div>
      <div className="field">
        <label>Password</label>
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Enter a password"
          required
        />
      </div>
      <button className="btn primary" type="submit">
        Create user
      </button>
    </form>
  );
}

// LoginForm validates credentials for local auth.
export function LoginForm({
  onLogin,
  error,
  onClearError,
}: {
  onLogin: (email: string, password: string) => void | Promise<void>;
  error?: string | null;
  onClearError?: () => void;
}) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const trimmedEmail = email.trim();
  const emailInvalid = trimmedEmail !== "" && !isValidEmail(trimmedEmail);
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        // Guard: require valid credentials before submit.
        if (!email.trim() || !password.trim()) return;
        if (!isValidEmail(email.trim())) return;
        onLogin(email.trim(), password.trim());
      }}
      className="stack"
    >
      {error && <p className="muted small">{error}</p>}
      <div className="field">
        <label>Email</label>
        <input
          value={email}
          onChange={(e) => {
            setEmail(e.target.value);
            onClearError?.();
          }}
          placeholder="you@example.com"
          className={emailInvalid ? "input-error" : undefined}
          required
        />
        {emailInvalid && (
          <div className="helper error">Enter a valid email address</div>
        )}
      </div>
      <div className="field">
        <label>Password</label>
        <input
          type="password"
          value={password}
          onChange={(e) => {
            setPassword(e.target.value);
            onClearError?.();
          }}
          placeholder="Your password"
          required
        />
      </div>
      <button
        className="btn primary"
        type="submit"
        disabled={!email.trim() || !password.trim()}
      >
        Log in
      </button>
    </form>
  );
}
