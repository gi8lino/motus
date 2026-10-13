import type { ReactNode } from "react";
import {
  Alert,
  AppBar,
  BottomNavigation,
  BottomNavigationAction,
  Box,
  Button,
  Container,
  Paper,
  Snackbar,
  Stack,
  Toolbar,
  Typography,
} from "@mui/material";
import DirectionsRunRoundedIcon from "@mui/icons-material/DirectionsRunRounded";
import FitnessCenterRoundedIcon from "@mui/icons-material/FitnessCenterRounded";
import HistoryRoundedIcon from "@mui/icons-material/HistoryRounded";
import PersonRoundedIcon from "@mui/icons-material/PersonRounded";
import TuneRoundedIcon from "@mui/icons-material/TuneRounded";
import type { User, View } from "../../types";
import { PROMPTS } from "../../utils/messages";
import { BrandHeader } from "../common/BrandHeader";
import { NavTabs } from "./NavigationTabs";

type AppShellProps = {
  view: View;
  onViewChange: (view: View) => void;
  onViewPreload: (view: View) => void;
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

type ToastProps = {
  message: string | null;
};

// Toast renders short application feedback above either shell.
function Toast({ message }: ToastProps) {
  const severity = /error|failed|unable|denied/i.test(message ?? "")
    ? "error"
    : "success";

  return (
    <Snackbar
      open={Boolean(message)}
      autoHideDuration={1800}
      anchorOrigin={{ vertical: "top", horizontal: "right" }}
    >
      <Alert severity={severity} variant="filled" sx={{ width: "100%" }}>
        {message}
      </Alert>
    </Snackbar>
  );
}

// AppFooter renders the compact Motus version footer.
function AppFooter({
  appVersion,
  auth = false,
}: {
  appVersion: string;
  auth?: boolean;
}) {
  return (
    <Box
      component="footer"
      sx={{
        borderTop: auth ? 0 : 1,
        borderColor: "divider",
        bgcolor: auth
          ? "transparent"
          : (theme) =>
              theme.palette.mode === "dark"
                ? "rgba(8, 18, 26, 0.92)"
                : "rgba(255, 252, 247, 0.92)",
        backdropFilter: auth ? "none" : "blur(14px)",
        py: auth ? 2 : 0.75,
        px: 2,
        zIndex: (theme) => theme.zIndex.appBar - 1,
      }}
    >
      <Typography
        variant="caption"
        color="text.secondary"
        sx={{ display: "block", textAlign: "center" }}
      >
        {auth
          ? `Motus ${appVersion} · Training companion`
          : `© ${new Date().getFullYear()} Motus · Version ${appVersion}`}
      </Typography>
    </Box>
  );
}

// AppShell renders the authenticated application chrome or the public auth shell.
export function AppShell({
  view,
  onViewChange,
  onViewPreload,
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

  if (!isAuthed) {
    return (
      <Box
        sx={{
          minHeight: "100vh",
          display: "grid",
          gridTemplateRows: "1fr auto",
          bgcolor: "background.default",
        }}
      >
        <Box
          component="main"
          key={view}
          sx={{
            width: "min(440px, calc(100vw - 32px))",
            placeSelf: "center",
            display: "grid",
            gap: 2.25,
            py: 5,
          }}
        >
          <BrandHeader auth />
          {children}
        </Box>
        <AppFooter appVersion={appVersion} auth />
        <Toast message={toast} />
      </Box>
    );
  }

  const availableViews: View[] = (
    [
      "training",
      "workouts",
      "exercises",
      "history",
      "profile",
      "admin",
    ] as View[]
  ).filter((nextView) =>
    nextView === "admin" ? Boolean(currentUser?.isAdmin) : true,
  );
  const mobileViews = availableViews.filter(
    (nextView): nextView is Exclude<View, "admin" | "login"> =>
      nextView !== "admin" && nextView !== "login",
  );
  const mobileNavigation: Record<
    Exclude<View, "admin" | "login">,
    ReactNode
  > = {
    training: <DirectionsRunRoundedIcon />,
    workouts: <FitnessCenterRoundedIcon />,
    exercises: <TuneRoundedIcon />,
    history: <HistoryRoundedIcon />,
    profile: <PersonRoundedIcon />,
  } as const;
  const mobileView = mobileViews.find((nextView) => nextView === view);

  return (
    <>
      <Box sx={{ minHeight: "calc(100vh - 34px)" }}>
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
            <BrandHeader onClick={() => onViewChange("training")} />

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
                onPreload={onViewPreload}
              />

              {currentUser?.isAdmin ? (
                <Button
                  variant={view === "admin" ? "contained" : "text"}
                  onClick={() => onViewChange("admin")}
                  sx={{ display: { xs: "inline-flex", lg: "none" } }}
                >
                  Admin
                </Button>
              ) : null}

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
          </Toolbar>
        </AppBar>

        <Container
          maxWidth="xl"
          sx={{
            px: { xs: 2, md: 3 },
            pt: { xs: 2.5, md: 4 },
            pb: { xs: 11, md: 4 },
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
                sx={{
                  alignItems: { xs: "flex-start", md: "center" },
                  justifyContent: "space-between",
                }}
              >
                <Box>
                  <Typography variant="subtitle1" sx={{ fontWeight: 800 }}>
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

      <Toast message={toast} />
      <AppFooter appVersion={appVersion} />

      <Paper
        elevation={0}
        sx={{
          display: { xs: "block", lg: "none" },
          position: "fixed",
          bottom: 0,
          left: 0,
          right: 0,
          zIndex: (theme) => theme.zIndex.appBar,
          borderTop: 1,
          borderColor: "divider",
          borderRadius: 0,
          bgcolor: (theme) =>
            theme.palette.mode === "dark"
              ? "rgba(8, 16, 25, 0.94)"
              : "rgba(255, 252, 247, 0.96)",
          backdropFilter: "blur(18px)",
        }}
      >
        <BottomNavigation
          showLabels
          value={mobileView ?? false}
          onChange={(_event, nextView: View) => onViewChange(nextView)}
          sx={{ height: 70 }}
        >
          {mobileViews.map((nextView) => (
            <BottomNavigationAction
              key={nextView}
              value={nextView}
              label={nextView === "training" ? "Train" : nextView}
              icon={mobileNavigation[nextView]}
              onFocus={() => onViewPreload(nextView)}
            />
          ))}
        </BottomNavigation>
      </Paper>
    </>
  );
}
