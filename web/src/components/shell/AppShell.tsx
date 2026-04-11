import type { ReactNode } from "react";
import {
  Alert,
  AppBar,
  Box,
  Button,
  Container,
  Paper,
  Snackbar,
  Stack,
  Toolbar,
  Typography,
} from "@mui/material";
import type { User, View } from "../../types";
import { PROMPTS } from "../../utils/messages";
import { BrandHeader } from "../common/BrandHeader";
import { NavTabs } from "./NavigationTabs";

type AppShellProps = {
  view: View;
  onViewChange: (view: View) => void;
  currentUser: User | null;
  authHeaderEnabled: boolean;
  onLogout?: () => void;
  resumeOpen: boolean;
  resumeText: string;
  onResume: () => void;
  onDismissResume: () => void;
  toast: string | null;
  appVersion: string;
  children: ReactNode;
};

export function AppShell({
  view,
  onViewChange,
  currentUser,
  authHeaderEnabled,
  onLogout,
  resumeOpen,
  resumeText,
  onResume,
  onDismissResume,
  toast,
  appVersion,
  children,
}: AppShellProps) {
  const isAuthed = Boolean(authHeaderEnabled || currentUser);

  const availableViews: View[] = (
    [
      "train",
      "workouts",
      "templates",
      "exercises",
      "history",
      "profile",
      "admin",
    ] as View[]
  ).filter((nextView) =>
    nextView === "admin" ? Boolean(currentUser?.isAdmin) : true,
  );

  return (
    <>
      <Box sx={{ minHeight: "100vh", pb: { xs: 12, md: 10 } }}>
        <AppBar
          position="sticky"
          color="transparent"
          elevation={0}
          sx={{
            borderBottom: 1,
            borderColor: "divider",
            bgcolor: (theme) =>
              theme.palette.mode === "dark"
                ? "rgba(7, 16, 24, 0.78)"
                : "rgba(250, 246, 240, 0.84)",
          }}
        >
          <Toolbar
            sx={{
              minHeight: { xs: 78, md: 88 },
              px: { xs: 2, md: 3 },
              py: { xs: 1.5, md: 2 },
              alignItems: { xs: "flex-start", lg: "center" },
              gap: 2,
              flexDirection: { xs: "column", lg: "row" },
            }}
          >
            <BrandHeader />

            {isAuthed ? (
              <Box
                sx={{
                  ml: { lg: "auto" },
                  width: { xs: "100%", lg: "auto" },
                  display: "flex",
                  alignItems: { xs: "stretch", lg: "center" },
                  justifyContent: "space-between",
                  gap: 1.5,
                  flexWrap: "wrap",
                }}
              >
                <NavTabs
                  view={view}
                  views={availableViews}
                  onSelect={onViewChange}
                />

                {!authHeaderEnabled && currentUser && onLogout ? (
                  <Button
                    color="inherit"
                    variant="outlined"
                    onClick={onLogout}
                    sx={{
                      alignSelf: { xs: "stretch", sm: "center" },
                      borderColor: "divider",
                    }}
                  >
                    Logout
                  </Button>
                ) : null}
              </Box>
            ) : null}
          </Toolbar>
        </AppBar>

        <Container
          maxWidth="xl"
          sx={{
            px: { xs: 2, md: 3 },
            py: { xs: 2.5, md: 4 },
          }}
        >
          {resumeOpen ? (
            <Paper
              elevation={0}
              sx={{
                mb: 3,
                p: { xs: 2, md: 2.5 },
                borderRadius: 4,
                border: 1,
                borderColor: "primary.main",
                bgcolor: (theme) =>
                  theme.palette.mode === "dark"
                    ? "rgba(18, 43, 61, 0.88)"
                    : "rgba(233, 248, 246, 0.94)",
              }}
            >
              <Stack
                direction={{ xs: "column", md: "row" }}
                spacing={2}
                alignItems={{ xs: "flex-start", md: "center" }}
                justifyContent="space-between"
              >
                <Box>
                  <Typography variant="subtitle1" fontWeight={800}>
                    {PROMPTS.resumeTrainingTitle}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    {resumeText}
                  </Typography>
                </Box>

                <Stack
                  direction={{ xs: "column", sm: "row" }}
                  spacing={1.25}
                  sx={{ width: { xs: "100%", md: "auto" } }}
                >
                  <Button
                    variant="contained"
                    onClick={onResume}
                    sx={{ minWidth: 148 }}
                  >
                    {PROMPTS.resumeTrainingConfirm}
                  </Button>
                  <Button
                    variant="text"
                    color="inherit"
                    onClick={onDismissResume}
                  >
                    Dismiss
                  </Button>
                </Stack>
              </Stack>
            </Paper>
          ) : null}

          <Box
            component="main"
            key={view}
            sx={{ display: "flex", flexDirection: "column", gap: 3 }}
          >
            {children}
          </Box>
        </Container>
      </Box>

      <Snackbar
        open={Boolean(toast)}
        autoHideDuration={1800}
        anchorOrigin={{ vertical: "top", horizontal: "right" }}
      >
        <Alert severity="info" variant="filled" sx={{ width: "100%" }}>
          {toast}
        </Alert>
      </Snackbar>

      <Box
        component="footer"
        sx={{
          position: "fixed",
          left: 0,
          right: 0,
          bottom: 0,
          borderTop: 1,
          borderColor: "divider",
          bgcolor: (theme) =>
            theme.palette.mode === "dark"
              ? "rgba(8, 18, 26, 0.92)"
              : "rgba(255, 252, 247, 0.92)",
          backdropFilter: "blur(14px)",
          py: 0.75,
          px: 2,
          zIndex: (theme) => theme.zIndex.appBar - 1,
        }}
      >
        <Typography
          variant="caption"
          display="block"
          textAlign="center"
          color="text.secondary"
        >
          © 2025 Motus | Version: {appVersion}
        </Typography>
      </Box>
    </>
  );
}
