import { Button, Card, CardContent, Stack, Typography } from "@mui/material";
import { LoginForm, SetupForm, UserForm } from "./../auth/AuthForm";
import { withBasePath } from "../../utils/basePath";

export type LoginViewData = {
  allowRegistration: boolean;
  oidcEnabled: boolean;
  localLoginEnabled: boolean;
  localLoginRequested: boolean;
  setupRequired: boolean;
  loginError: string | null;
};

export type LoginViewActions = {
  onLogin: (email: string, password: string) => void | Promise<void>;
  onCreateUser: (email: string, password: string) => void | Promise<void>;
  onClearError: () => void;
};

const authCardSx = {
  borderRadius: "10px",
  boxShadow: "none",
};

// LoginView renders local, OIDC, and first-run authentication in the public auth shell.
export function LoginView({
  data,
  actions,
}: {
  data: LoginViewData;
  actions: LoginViewActions;
}) {
  const {
    allowRegistration,
    oidcEnabled,
    localLoginEnabled,
    localLoginRequested,
    setupRequired,
    loginError,
  } = data;
  const { onLogin, onCreateUser, onClearError } = actions;
  const eyebrow = setupRequired
    ? "First-time setup"
    : oidcEnabled && !localLoginRequested
      ? "External access"
      : "Local access";
  const title = setupRequired
    ? "Create the Motus administrator"
    : "Sign in to Motus";
  const description = setupRequired
    ? "Create Motus’s first administrator. This page disappears after the first account is created."
    : oidcEnabled && !localLoginRequested
      ? "Continue with the identity provider configured for this Motus instance."
      : "Use a Motus-managed account.";

  return (
    <Stack spacing={2.25}>
      <Card sx={authCardSx}>
        <CardContent sx={{ p: 3, "&:last-child": { pb: 3 } }}>
          <Stack spacing={0.5} sx={{ mb: 2.25 }}>
            <Typography
              variant="overline"
              color="primary.main"
              sx={{ fontWeight: 800, lineHeight: 1.4, letterSpacing: "0.08em" }}
            >
              {eyebrow}
            </Typography>
            <Typography
              component="h1"
              sx={{ fontSize: 22, fontWeight: 700, lineHeight: 1.25 }}
            >
              {title}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {description}
            </Typography>
          </Stack>

          {setupRequired ? (
            <SetupForm onCreate={onCreateUser} />
          ) : oidcEnabled && !localLoginRequested ? (
            <Stack spacing={1.5}>
              <Button
                fullWidth
                size="large"
                variant="contained"
                href={withBasePath("/auth/oidc/login")}
              >
                Continue with identity provider
              </Button>
              {localLoginEnabled ? (
                <Button
                  fullWidth
                  size="small"
                  href={withBasePath("/auth/local")}
                >
                  Sign in with local password
                </Button>
              ) : null}
            </Stack>
          ) : (
            <LoginForm
              onLogin={onLogin}
              error={loginError}
              onClearError={onClearError}
            />
          )}
        </CardContent>
      </Card>

      {allowRegistration && !oidcEnabled && !setupRequired ? (
        <Card sx={authCardSx}>
          <CardContent sx={{ p: 3, "&:last-child": { pb: 3 } }}>
            <Stack spacing={0.5} sx={{ mb: 2.25 }}>
              <Typography
                variant="overline"
                color="primary.main"
                sx={{
                  fontWeight: 800,
                  lineHeight: 1.4,
                  letterSpacing: "0.08em",
                }}
              >
                New account
              </Typography>
              <Typography
                component="h2"
                sx={{ fontSize: 20, fontWeight: 700, lineHeight: 1.25 }}
              >
                Create a Motus account
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Registration is enabled for this instance.
              </Typography>
            </Stack>
            <UserForm onCreate={onCreateUser} />
          </CardContent>
        </Card>
      ) : !setupRequired && !oidcEnabled ? (
        <Typography
          variant="body2"
          color="text.secondary"
          sx={{ textAlign: "center" }}
        >
          New-user registration is closed.
        </Typography>
      ) : null}
    </Stack>
  );
}
