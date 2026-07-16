import { LoginForm, UserForm } from "./../auth/AuthForm";
import { UI_TEXT } from "../../utils/uiText";
import { Box, Card, CardContent, Chip, Stack, Typography } from "@mui/material";

export type LoginViewData = {
  allowRegistration: boolean;
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
  const { allowRegistration, loginError } = data;
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
              label="Your training space"
              color="primary"
              variant="outlined"
              sx={{ alignSelf: "flex-start" }}
            />
            <Typography variant="h3">
              {UI_TEXT.pages.auth.loginTitle}
            </Typography>
            <Typography color="text.secondary">
              Pick up your plan, timer, and training history in one place.
            </Typography>
          </Stack>
          {/* Local login form */}
          <LoginForm
            onLogin={onLogin}
            error={loginError}
            onClearError={onClearError}
          />
        </CardContent>
      </Card>
      {allowRegistration ? (
        <Card sx={{ mt: 3 }}>
          <CardContent sx={{ p: { xs: 3, sm: 4 } }}>
            <Typography variant="h5" sx={{ mb: 2 }}>
              {UI_TEXT.pages.auth.createUserTitle}
            </Typography>
            <UserForm onCreate={onCreateUser} />
          </CardContent>
        </Card>
      ) : (
        <Typography
          variant="body2"
          color="text.secondary"
          sx={{ mt: 2.5, textAlign: "center" }}
        >
          {UI_TEXT.pages.auth.registrationDisabledHint}
        </Typography>
      )}
    </Box>
  );
}
