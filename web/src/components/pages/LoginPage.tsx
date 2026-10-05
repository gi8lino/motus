import { LoginForm, UserForm } from "./../auth/AuthForm";
import { UI_TEXT } from "../../utils/uiText";
import { Box, Card, CardContent, Chip, Stack, Typography } from "@mui/material";
import { Button } from "@mui/material";
import { withBasePath } from "../../utils/basePath";

export type LoginViewData = {
  allowRegistration: boolean;
  oidcEnabled: boolean;
  setupRequired: boolean;
  loginError: string | null;
};

export type LoginViewActions = {
  onLogin: (email: string, password: string) => void | Promise<void>;
  onCreateUser: (email: string, password: string) => void | Promise<void>;
  onClearError: () => void;
};

// LoginView renders local login and optional registration.
export function LoginView({
  data,
  actions,
}: {
  data: LoginViewData;
  actions: LoginViewActions;
}) {
  const { allowRegistration, oidcEnabled, setupRequired, loginError } = data;
  const { onLogin, onCreateUser, onClearError } = actions;
  return (
    <Box
      sx={{ maxWidth: 520, mx: "auto", width: "100%", py: { xs: 1, md: 5 } }}
    >
      <Card>
        <CardContent
          sx={{ p: { xs: 3, sm: 5 }, "&:last-child": { pb: { xs: 3, sm: 5 } } }}
        >
          <Stack spacing={1} sx={{ mb: 3.5 }}>
            <Chip
              label={setupRequired ? "First-run setup" : "Your training space"}
              color="primary"
              variant="outlined"
              sx={{ alignSelf: "flex-start" }}
            />
            <Typography variant="h3">
              {setupRequired ? "Set up Motus" : UI_TEXT.pages.auth.loginTitle}
            </Typography>
            <Typography color="text.secondary">
              {setupRequired
                ? "Create Motus’s first administrator. You can configure other sign-in methods later."
                : "Pick up your plan, timer, and training history in one place."}
            </Typography>
          </Stack>
          {setupRequired ? (
            <UserForm
              onCreate={onCreateUser}
              submitLabel="Create administrator"
            />
          ) : oidcEnabled ? (
            <Stack spacing={1.5}>
              <Button
                fullWidth
                size="large"
                variant="contained"
                href={withBasePath("/auth/oidc/login")}
              >
                Continue with your identity provider
              </Button>
              <Typography
                variant="body2"
                color="text.secondary"
                sx={{ textAlign: "center" }}
              >
                You will return here after signing in.
              </Typography>
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
        <Card sx={{ mt: 3 }}>
          <CardContent sx={{ p: { xs: 3, sm: 4 } }}>
            <Typography variant="h5" sx={{ mb: 2 }}>
              {UI_TEXT.pages.auth.createUserTitle}
            </Typography>
            <UserForm onCreate={onCreateUser} />
          </CardContent>
        </Card>
      ) : !setupRequired ? (
        <Typography
          variant="body2"
          color="text.secondary"
          sx={{ mt: 2.5, textAlign: "center" }}
        >
          {UI_TEXT.pages.auth.registrationDisabledHint}
        </Typography>
      ) : null}
    </Box>
  );
}
