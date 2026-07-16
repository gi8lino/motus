import { useState } from "react";
import { Alert, Button, Stack, TextField } from "@mui/material";

import { isValidEmail } from "../../utils/validation";
import { UI_TEXT } from "../../utils/uiText";

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
    <Stack
      component="form"
      onSubmit={(e) => {
        e.preventDefault();
        // Guard: require valid credentials before submit.
        if (!email.trim() || !password.trim()) return;
        if (!isValidEmail(email.trim())) return;
        onCreate(email.trim(), password.trim());
        setEmail("");
        setPassword("");
      }}
      spacing={2}
    >
      <TextField
        label="Email"
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="you@example.com"
        error={emailInvalid}
        helperText={emailInvalid ? "Enter a valid email address" : " "}
        autoComplete="email"
        required
      />
      <TextField
        label="Password"
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder={UI_TEXT.auth.enterPassword}
        autoComplete="new-password"
        required
      />
      <Button variant="contained" size="large" type="submit">
        Create user
      </Button>
    </Stack>
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
    <Stack
      component="form"
      onSubmit={(e) => {
        e.preventDefault();
        // Guard: require valid credentials before submit.
        if (!email.trim() || !password.trim()) return;
        if (!isValidEmail(email.trim())) return;
        onLogin(email.trim(), password.trim());
      }}
      spacing={2}
    >
      {error && <Alert severity="error">{error}</Alert>}
      <TextField
        label="Email"
        type="email"
        value={email}
        onChange={(e) => {
          setEmail(e.target.value);
          onClearError?.();
        }}
        placeholder="you@example.com"
        error={emailInvalid}
        helperText={emailInvalid ? "Enter a valid email address" : " "}
        autoComplete="email"
        autoFocus
        required
      />
      <TextField
        label="Password"
        type="password"
        value={password}
        onChange={(e) => {
          setPassword(e.target.value);
          onClearError?.();
        }}
        placeholder={UI_TEXT.auth.yourPassword}
        autoComplete="current-password"
        required
      />
      <Button
        variant="contained"
        size="large"
        type="submit"
        disabled={!email.trim() || !password.trim()}
      >
        Log in
      </Button>
    </Stack>
  );
}
