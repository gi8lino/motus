import {
  Suspense,
  lazy,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  Box,
  Card,
  CardContent,
  CircularProgress,
  CssBaseline,
  ThemeProvider,
  Typography,
} from "@mui/material";
import {
  applyTemplate,
  getWorkout,
  listExercises,
  updateUserName,
} from "./api";

import type { CatalogExercise, ThemeMode, User } from "./types";

import { useTrainingTimer } from "./hooks/useTrainingTimer";
import { useDialog } from "./hooks/useDialog";
import { useViewState } from "./hooks/useViewState";
import { useAppConfig } from "./hooks/useAppConfig";
import { useWorkoutsData } from "./hooks/useWorkoutsData";
import { useUserDefaults } from "./hooks/useUserDefaults";

import { AppShell } from "./components/shell/AppShell";
import DialogModal from "./components/common/DialogModal";

import { isValidEmail } from "./utils/validation";
import { PROMPTS, toErrorMessage } from "./utils/messages";
import { UI_TEXT } from "./utils/uiText";
import { buildAppTheme } from "./theme";

import { useAuthActions } from "./hooks/useAuthActions";
import { useAdminActions } from "./hooks/useAdminActions";
import { useExerciseActions } from "./hooks/useExerciseActions";
import { useProfileActions } from "./hooks/useProfileActions";
import { useTrainingActions } from "./hooks/useTrainingActions";

import "./styles.css";

const LoginView = lazy(() =>
  import("./components/pages/LoginPage").then((module) => ({
    default: module.LoginView,
  })),
);
const AdminView = lazy(() =>
  import("./components/pages/AdminPage").then((module) => ({
    default: module.AdminView,
  })),
);
const WorkoutsView = lazy(() =>
  import("./components/pages/WorkoutsPage").then((module) => ({
    default: module.WorkoutsView,
  })),
);
const TrainingView = lazy(() =>
  import("./components/pages/TrainingPage").then((module) => ({
    default: module.TrainingView,
  })),
);
const TemplatesView = lazy(() =>
  import("./components/pages/TemplatesPage").then((module) => ({
    default: module.TemplatesView,
  })),
);
const HistoryView = lazy(() =>
  import("./components/pages/HistoryPage").then((module) => ({
    default: module.HistoryView,
  })),
);
const ProfileView = lazy(() =>
  import("./components/pages/ProfilePage").then((module) => ({
    default: module.ProfileView,
  })),
);
const ExercisesView = lazy(() =>
  import("./components/pages/ExercisesPage").then((module) => ({
    default: module.ExercisesView,
  })),
);

// resumeMessage formats a resume prompt for an in-progress training.
function resumeMessage(
  training?: ReturnType<typeof useTrainingTimer>["training"] | null,
) {
  if (!training) return "";
  const name = training.workoutName || "your workout";
  return `Resume ${name}?`;
}

// shouldResetStoredUserID checks whether a loader error means the local user id is invalid.
function shouldResetStoredUserID(error: string | null): boolean {
  const text = (error || "").toLowerCase();
  if (!text) return false;
  return (
    text.includes("unauthorized") ||
    text.includes("not found") ||
    text.includes("auth header is required") ||
    text.includes("userid is required")
  );
}

function PageFallback() {
  return (
    <Card sx={{ maxWidth: 560, mx: "auto" }}>
      <CardContent
        sx={{
          minHeight: 220,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          gap: 2,
        }}
      >
        <CircularProgress size={28} />
        <Box>
          <Typography variant="h6">Loading view</Typography>
          <Typography color="text.secondary">
            Pulling in the next part of the app.
          </Typography>
        </Box>
      </CardContent>
    </Card>
  );
}

export default function App() {
  const { view, setView } = useViewState("train");

  const [currentUserId, setCurrentUserId] = useState<string | null>(() => {
    const stored = localStorage.getItem("motus:userId");
    if (stored && isValidEmail(stored)) return stored;
    return null;
  });

  const [loginError, setLoginError] = useState<string | null>(null);

  const [themeMode, setThemeMode] = useState<ThemeMode>(() => {
    const stored = localStorage.getItem("motus:theme");
    if (stored === "dark" || stored === "light" || stored === "auto")
      return stored;
    return "auto";
  });
  const [resolvedThemeMode, setResolvedThemeMode] = useState<
    "dark" | "light"
  >(() => {
    if (
      window.matchMedia &&
      window.matchMedia("(prefers-color-scheme: dark)").matches
    ) {
      return "dark";
    }
    return "light";
  });

  // train view state
  const [selectedWorkoutId, setSelectedWorkoutId] = useState<string | null>(
    null,
  );

  // misc
  const [exerciseCatalog, setExerciseCatalog] = useState<CatalogExercise[]>([]);
  const [toast, setToast] = useState<string | null>(null);
  const [profileTab, setProfileTab] = useState<
    "settings" | "password" | "transfer"
  >("settings");
  const [exportWorkoutId, setExportWorkoutId] = useState("");

  const historyReloadGuard = useRef<string | null>(null);

  const {
    dialog,
    dialogValue,
    setDialogValue,
    closeDialog,
    notify,
    askConfirm,
    askPrompt,
  } = useDialog();

  const { config, authError } = useAppConfig({
    view,
    setView,
    setCurrentUserId,
  });
  const allowRegistration = config?.allowRegistration ?? true;
  const authHeaderEnabled = config?.authHeaderEnabled ?? false;
  const appVersion = config?.version || "dev";
  const {
    users,
    sounds,
    workouts,
    history,
    templates,
    activeWorkouts,
    currentUser,
    currentUserLoader,
  } = useWorkoutsData({ currentUserId, authHeaderEnabled });
  const {
    defaultStepSoundKey,
    defaultPauseDuration,
    defaultPauseSoundKey,
    defaultPauseAutoAdvance,
    repeatRestAfterLastDefault,
    pauseOnTabHidden,
    showHours,
    updateRepeatRestAfterLastDefault,
    updateDefaultStepSoundKey,
    updateDefaultPauseDuration,
    updateDefaultPauseSoundKey,
    updateDefaultPauseAutoAdvance,
    updatePauseOnTabHidden,
    updateShowHours,
  } = useUserDefaults({ currentUserId });

  const currentWorkoutName = useMemo(() => {
    if (!selectedWorkoutId) return "";
    return activeWorkouts.find((w) => w.id === selectedWorkoutId)?.name || "";
  }, [selectedWorkoutId, activeWorkouts]);

  // ---------- theme ----------
  useEffect(() => {
    const root = document.documentElement;

    const applyTheme = () => {
      if (themeMode === "auto") {
        const prefersDark = window.matchMedia(
          "(prefers-color-scheme: dark)",
        ).matches;
        const nextMode = prefersDark ? "dark" : "light";
        root.dataset.theme = nextMode;
        setResolvedThemeMode(nextMode);
        return;
      }
      root.dataset.theme = themeMode;
      setResolvedThemeMode(themeMode);
    };

    localStorage.setItem("motus:theme", themeMode);
    applyTheme();

    if (themeMode !== "auto") return;
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = () => applyTheme();
    media.addEventListener("change", handler);
    return () => media.removeEventListener("change", handler);
  }, [themeMode]);

  const theme = useMemo(
    () => buildAppTheme(resolvedThemeMode),
    [resolvedThemeMode],
  );

  // ---------- validate stored user id once local user info is known ----------
  useEffect(() => {
    if (authHeaderEnabled) return;
    if (!currentUserId) return;
    if (currentUser) return;
    if (currentUserLoader.loading) return;
    if (!shouldResetStoredUserID(currentUserLoader.error)) return;

    localStorage.removeItem("motus:userId");
    setCurrentUserId(null);
  }, [
    authHeaderEnabled,
    currentUserId,
    currentUser,
    currentUserLoader.loading,
    currentUserLoader.error,
  ]);

  // ---------- clear login errors when leaving login view ----------
  useEffect(() => {
    if (view === "login") setLoginError(null);
  }, [view]);

  // ---------- auto-redirect to train when local user logs in ----------
  useEffect(() => {
    if (authHeaderEnabled) return;
    if (currentUserId && view === "login") setView("train");
  }, [authHeaderEnabled, currentUserId, view, setView]);

  // ---------- force login when local auth has no user ----------
  useEffect(() => {
    if (authHeaderEnabled) return;
    if (currentUserLoader.loading) return;
    if (!currentUserId && view !== "login") setView("login");
  }, [
    authHeaderEnabled,
    currentUserId,
    currentUserLoader.loading,
    view,
    setView,
  ]);

  // ---------- keep exercise catalog in sync ----------
  useEffect(() => {
    if (!authHeaderEnabled && !currentUserId) {
      setExerciseCatalog([]);
      return;
    }
    listExercises()
      .then((items) => setExerciseCatalog(items || []))
      .catch(() => {});
  }, [authHeaderEnabled, currentUserId]);

  // ---------- toast ----------
  const showToast = useCallback((message: string) => {
    setToast(message);
    setTimeout(() => setToast((t) => (t === message ? null : t)), 1800);
  }, []);

  // ---------- auth actions ----------
  const onLoginSuccess = (user: User) => {
    setCurrentUserId(user.id);
    localStorage.setItem("motus:userId", user.id);
    setView("train");
  };

  const onRegisterSuccess = (user: User) => {
    users.setData?.((prev) => (prev ? [...prev, user] : [user]));
    setCurrentUserId(user.id);
    if (!authHeaderEnabled) localStorage.setItem("motus:userId", user.id);
  };

  const { login: handleLogin, register: handleRegister } = useAuthActions({
    setLoginError,
    onLoginSuccess,
    onRegisterSuccess,
  });

  // ---------- admin actions ----------
  const { toggleAdmin: handleToggleAdmin, backfillCatalog } = useAdminActions({
    currentUserId,
    setUsers: (updater) => users.setData?.(updater),
    setView,
    notify,
  });

  // ---------- exercise actions ----------
  const {
    createExerciseEntry,
    addExercise: handleAddExercise,
    addCoreExercise: handleAddCoreExercise,
    renameExercise: handleRenameExercise,
    deleteExerciseEntry: handleDeleteExercise,
  } = useExerciseActions({
    isAdmin: Boolean(currentUser?.isAdmin),
    setExerciseCatalog,
    askPrompt,
    askConfirm,
    notify,
    showToast,
  });

  // ---------- profile actions ----------
  const importInputRef = useRef<HTMLInputElement | null>(null);

  const {
    exportSelectedWorkout: handleExportSelected,
    importWorkoutFile: handleImportSelected,
    updatePassword: handlePasswordSubmit,
  } = useProfileActions({
    currentUserId,
    exportWorkoutId,
    setSelectedWorkoutId,
    setWorkouts: (updater) => workouts.setData?.(updater),
    showToast,
    notify,
  });

  // ---------- update user name ----------
  const handleUpdateName = useCallback(
    async (name: string) => {
      if (!currentUserId) throw new Error("No active user");
      await updateUserName(name);
      users.setData?.((prev) =>
        prev
          ? prev.map((u) => (u.id === currentUserId ? { ...u, name } : u))
          : prev,
      );
    },
    [currentUserId, users],
  );

  // ---------- training timer ----------
  const {
    training,
    currentStep,
    displayedElapsed,
    restoredFromStorage,
    startFromState,
    startCurrentStep,
    pause,
    nextStep,
    finishAndLog,
    markSoundPlayed,
    clear: clearTraining,
  } = useTrainingTimer({ currentUserId });

  const [promptedResume, setPromptedResume] = useState(false);
  const [resumeSuppressed, setResumeSuppressed] = useState(false);

  const {
    startTraining: handleStartTraining,
    finishTraining: handleFinishTraining,
  } = useTrainingActions({
    selectedWorkoutId,
    training,
    currentWorkoutName,
    setTrainingView: () => setView("train"),
    setPromptedResume,
    setResumeSuppressed,
    startFromState,
    finishAndLog,
    historyReload: () => history.reload(),
    askConfirm,
    notify,
  });

  // prompt resume once
  useEffect(() => {
    if (!restoredFromStorage || !training || training.done) return;
    if (promptedResume || resumeSuppressed) return;
    setPromptedResume(true);
  }, [restoredFromStorage, training, promptedResume, resumeSuppressed]);

  // reset resume suppression when training clears
  useEffect(() => {
    if (!training) setResumeSuppressed(false);
  }, [training?.trainingId]);

  // Auto-advance is handled in the training timer hook to keep timing consistent.

  // refresh history once when a training logs
  useEffect(() => {
    if (
      training?.logged &&
      training.trainingId !== historyReloadGuard.current
    ) {
      historyReloadGuard.current = training.trainingId;
      history.reload();
    }
  }, [training?.logged, training?.trainingId, history]);

  // ---------- logout ----------
  const handleLogout = () => {
    localStorage.removeItem("motus:userId");
    setCurrentUserId(null);
    setView("login");
    clearTraining();
  };

  // ---------- template apply ----------
  const handleApplyTemplate = useCallback(
    async (templateId: string) => {
      if (!currentUserId) {
        await notify("Select a user first.");
        return;
      }
      const name = await askPrompt("Workout name (optional)");
      if (name === null) return;

      try {
        const created = await applyTemplate(templateId, {
          userId: currentUserId,
          name: name.trim() || undefined,
        });

        workouts.setData?.((prev) => (prev ? [created, ...prev] : [created]));
        setView("workouts");
        showToast("Template applied.");
      } catch (err) {
        await notify(toErrorMessage(err, "Unable to apply template"));
      }
    },
    [askPrompt, currentUserId, notify, setView, showToast, workouts],
  );

  // ---------- guards ----------
  if (!config) {
    return (
      <div className="shell">
        <main>
          <section className="panel">
            <p className="muted">Loading configuration…</p>
          </section>
        </main>
      </div>
    );
  }

  if (authHeaderEnabled && authError) {
    return (
      <div className="shell">
        <main>
          <section className="panel">
            <h3>Access denied</h3>
            <p className="muted">{authError}</p>
            <p className="muted small">
              The reverse proxy did not supply a valid user header.
            </p>
          </section>
        </main>
      </div>
    );
  }

  const resumeOpen = promptedResume && Boolean(training && !training.done);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline enableColorScheme />

      <AppShell
        view={view}
        onViewChange={setView}
        currentUser={currentUser}
        authHeaderEnabled={authHeaderEnabled}
        onLogout={
          !authHeaderEnabled && currentUserId ? handleLogout : undefined
        }
        resumeOpen={resumeOpen}
        resumeText={resumeMessage(training)}
        onResume={() => {
          setPromptedResume(false);
          setView("train");
          if (!training?.running) startCurrentStep();
          setToast(null);
          setResumeSuppressed(true);
        }}
        onDismissResume={() => {
          setPromptedResume(false);
          clearTraining();
          setResumeSuppressed(true);
        }}
        toast={toast}
        appVersion={appVersion}
      >
          <Suspense fallback={<PageFallback />}>
            {view === "login" && !authHeaderEnabled && (
              <LoginView
            data={{
              allowRegistration,
              loginError,
            }}
            actions={{
              onLogin: handleLogin,
              onCreateUser: async (email, password) => {
                try {
                  await handleRegister(email, password);
                  setView("train");
                } catch (err) {
                  await notify(toErrorMessage(err, "Unable to create user"));
                }
              },
              onClearError: () => setLoginError(null),
            }}
              />
            )}

            {view === "admin" && currentUser?.isAdmin && (
              <AdminView
            data={{
              users: users.data || [],
              loading: users.loading,
              currentUserId,
              allowRegistration,
            }}
            actions={{
              onToggleAdmin: handleToggleAdmin,
              onCreateUser: async (email, password) => {
                try {
                  await handleRegister(email, password);
                } catch (err) {
                  await notify(toErrorMessage(err, "Unable to create user"));
                }
              },
              onBackfill: backfillCatalog,
            }}
              />
            )}

            {view === "workouts" && (
              <WorkoutsView
            workouts={activeWorkouts}
            loading={workouts.loading}
            setWorkouts={(updater) => workouts.setData?.(updater)}
            currentUserId={currentUserId}
            defaults={{
              defaultStepSoundKey,
              defaultPauseDuration,
              defaultPauseSoundKey,
              defaultPauseAutoAdvance,
              repeatRestAfterLastDefault,
            }}
            formData={{
              sounds: sounds.data || [],
              exerciseCatalog,
            }}
            services={{
              askConfirm,
              askPrompt,
              notifyUser: notify,
              templatesReload: () => templates.reload(),
              onCreateExercise: createExerciseEntry,
              promptUser: askPrompt,
              onToast: showToast,
            }}
              />
            )}

            {view === "train" && (
              <TrainingView
            data={{
              workouts: activeWorkouts,
              selectedWorkoutId,
              startDisabled: !selectedWorkoutId || !currentUserId,
              startTitle: !selectedWorkoutId ? PROMPTS.selectWorkoutFirst : "",
              training,
              currentStep,
              elapsed: displayedElapsed,
              workoutName: currentWorkoutName,
              sounds: sounds.data || [],
              pauseOnTabHidden,
              showHours,
            }}
            actions={{
              onSelectWorkout: setSelectedWorkoutId,
              onStartTraining: handleStartTraining,
              markSoundPlayed,
              onStartStep: startCurrentStep,
              onPause: pause,
              onNext: nextStep,
              onFinishTraining: handleFinishTraining,
              onCopySummary: () => showToast(UI_TEXT.toasts.copiedSummary),
              onToast: showToast,
            }}
              />
            )}

            {view === "templates" && (
              <TemplatesView
            data={{
              templates: templates.data || [],
              loading: templates.loading,
              hasUser: Boolean(currentUserId),
            }}
            actions={{
              onRefresh: () => templates.reload(),
              onApplyTemplate: handleApplyTemplate,
            }}
              />
            )}

            {view === "history" && (
              <HistoryView
            data={{
              items: history.data || [],
              activeTraining: training,
            }}
            actions={{
              onResume: () => setView("train"),
              loadWorkout: getWorkout,
              onCopySummary: () => showToast(UI_TEXT.toasts.copiedSummary),
            }}
              />
            )}

            {view === "profile" && (
              <ProfileView
            data={{
              profileTab,
              currentName: currentUser?.name || "",
              themeMode,
              sounds: sounds.data || [],
              defaultStepSoundKey,
              defaultPauseDuration,
              defaultPauseSoundKey,
              defaultPauseAutoAdvance,
              repeatRestAfterLastDefault,
              pauseOnTabHidden,
              showHours,
              exportWorkoutId,
              activeWorkouts,
              importInputRef,
              authHeaderEnabled,
            }}
            actions={{
              onProfileTabChange: setProfileTab,
              onUpdateName: handleUpdateName,
              onThemeChange: setThemeMode,
              onDefaultStepSoundChange: updateDefaultStepSoundKey,
              onDefaultPauseDurationChange: updateDefaultPauseDuration,
              onDefaultPauseSoundChange: updateDefaultPauseSoundKey,
              onDefaultPauseAutoAdvanceChange: updateDefaultPauseAutoAdvance,
              onRepeatRestAfterLastDefaultChange:
                updateRepeatRestAfterLastDefault,
              onPauseOnTabHiddenChange: updatePauseOnTabHidden,
              onShowHoursChange: updateShowHours,
              onExportWorkoutChange: setExportWorkoutId,
              onExportWorkout: handleExportSelected,
              onImportWorkout: handleImportSelected,
              onPasswordChange: handlePasswordSubmit,
            }}
              />
            )}

            {view === "exercises" && (
              <ExercisesView
            data={{
              exercises: exerciseCatalog,
              isAdmin: Boolean(currentUser?.isAdmin),
            }}
            actions={{
              onAddExercise: handleAddExercise,
              onAddCoreExercise: handleAddCoreExercise,
              onRenameExercise: handleRenameExercise,
              onDeleteExercise: handleDeleteExercise,
            }}
              />
            )}
          </Suspense>
      </AppShell>

      {dialog && (
        <DialogModal
          dialog={dialog}
          value={dialogValue}
          onValueChange={setDialogValue}
          onClose={closeDialog}
        />
      )}
    </ThemeProvider>
  );
}
