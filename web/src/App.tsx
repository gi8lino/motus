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
  getWorkout,
  logoutUser,
  updateTrainingFeedback,
  updateUserName,
} from "./api";

import type { ThemeMode, User, View } from "./types";

import { useTrainingTimer } from "./hooks/useTrainingTimer";
import { useDialog } from "./hooks/useDialog";
import { useViewState } from "./hooks/useViewState";
import { useAppConfig } from "./hooks/useAppConfig";
import { useWorkoutsData } from "./hooks/useWorkoutsData";
import { useUserDefaults } from "./hooks/useUserDefaults";

import { AppShell } from "./components/shell/AppShell";
import DialogModal from "./components/common/DialogModal";

import { PROMPTS, toErrorMessage } from "./utils/messages";
import { UI_TEXT } from "./utils/uiText";
import { buildAppTheme } from "./theme";

import { useAuthActions } from "./hooks/useAuthActions";
import { useAdminActions } from "./hooks/useAdminActions";
import { useExerciseActions } from "./hooks/useExerciseActions";
import { useProfileActions } from "./hooks/useProfileActions";
import { useTrainingActions } from "./hooks/useTrainingActions";

import "./styles.css";

const loadLoginView = () => import("./components/pages/LoginPage");
const loadAdminView = () => import("./components/pages/AdminPage");
const loadWorkoutsView = () => import("./components/pages/WorkoutsPage");
const loadTrainingView = () => import("./components/pages/TrainingPage");
const loadHistoryView = () => import("./components/pages/HistoryPage");
const loadProfileView = () => import("./components/pages/ProfilePage");
const loadExercisesView = () => import("./components/pages/ExercisesPage");

const viewLoaders: Partial<Record<View, () => Promise<unknown>>> = {
  login: loadLoginView,
  admin: loadAdminView,
  workouts: loadWorkoutsView,
  training: loadTrainingView,
  history: loadHistoryView,
  profile: loadProfileView,
  exercises: loadExercisesView,
};

const preloadView = (view: View) => {
  void viewLoaders[view]?.();
};

const LoginView = lazy(() =>
  loadLoginView().then((module) => ({
    default: module.LoginView,
  })),
);
const AdminView = lazy(() =>
  loadAdminView().then((module) => ({
    default: module.AdminView,
  })),
);
const WorkoutsView = lazy(() =>
  loadWorkoutsView().then((module) => ({
    default: module.WorkoutsView,
  })),
);
const TrainingView = lazy(() =>
  loadTrainingView().then((module) => ({
    default: module.TrainingView,
  })),
);
const HistoryView = lazy(() =>
  loadHistoryView().then((module) => ({
    default: module.HistoryView,
  })),
);
const ProfileView = lazy(() =>
  loadProfileView().then((module) => ({
    default: module.ProfileView,
  })),
);
const ExercisesView = lazy(() =>
  loadExercisesView().then((module) => ({
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
  const { view, setView } = useViewState("training");

  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const [prefetchedViews, setPrefetchedViews] = useState<ReadonlySet<View>>(
    () => new Set(),
  );

  const [loginError, setLoginError] = useState<string | null>(null);

  const [themeMode, setThemeMode] = useState<ThemeMode>(() => {
    const stored = localStorage.getItem("motus:theme");
    if (stored === "dark" || stored === "light" || stored === "auto")
      return stored;
    return "auto";
  });
  const [resolvedThemeMode, setResolvedThemeMode] = useState<"dark" | "light">(
    () => {
      if (
        window.matchMedia &&
        window.matchMedia("(prefers-color-scheme: dark)").matches
      ) {
        return "dark";
      }
      return "light";
    },
  );

  useEffect(() => {
    document.documentElement.dataset.theme = resolvedThemeMode;
  }, [resolvedThemeMode]);

  // training view state
  const [selectedWorkoutId, setSelectedWorkoutId] = useState<string | null>(
    null,
  );

  // misc
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
    exercises,
    history,
    activeWorkouts,
    currentUser,
    currentUserLoader,
  } = useWorkoutsData({
    currentUserId,
    authHeaderEnabled,
    view,
    prefetchedViews,
  });
  const exerciseCatalog = exercises.data || [];

  const handleViewPreload = useCallback((nextView: View) => {
    preloadView(nextView);
    setPrefetchedViews((current) => {
      if (current.has(nextView)) return current;
      const next = new Set(current);
      next.add(nextView);
      return next;
    });
  }, []);
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

  // ---------- clear login errors when leaving login view ----------
  useEffect(() => {
    if (view === "login") setLoginError(null);
  }, [view]);

  // ---------- auto-redirect to training when local user logs in ----------
  useEffect(() => {
    if (authHeaderEnabled) return;
    if (currentUserId && view === "login") setView("training");
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

  // Warm route chunks after authentication so navigation does not wait on a
  // network round-trip. Pointer/focus preloading below remains the fast path.
  useEffect(() => {
    if (!currentUserId && !authHeaderEnabled) return;
    const preload = () => {
      (
        [
          "workouts",
          "exercises",
          "history",
          "profile",
          ...(currentUser?.isAdmin ? (["admin"] as const) : []),
        ] as View[]
      ).forEach(preloadView);
    };
    if ("requestIdleCallback" in window) {
      const id = window.requestIdleCallback(preload, { timeout: 1500 });
      return () => window.cancelIdleCallback(id);
    }
    const id = globalThis.setTimeout(preload, 250);
    return () => globalThis.clearTimeout(id);
  }, [
    authHeaderEnabled,
    currentUser?.isAdmin,
    currentUserId,
    handleViewPreload,
  ]);

  // ---------- toast ----------
  const showToast = useCallback((message: string) => {
    setToast(message);
    setTimeout(() => setToast((t) => (t === message ? null : t)), 1800);
  }, []);

  // ---------- auth actions ----------
  const onLoginSuccess = (user: User) => {
    setCurrentUserId(user.id);
    setView("training");
  };

  const onRegisterSuccess = (user: User) => {
    users.setData?.((prev) => (prev ? [...prev, user] : [user]));
    setCurrentUserId(user.id);
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
    toggleExerciseSides: handleToggleExerciseSides,
  } = useExerciseActions({
    isAdmin: Boolean(currentUser?.isAdmin),
    setExerciseCatalog: (update) =>
      exercises.setData((previous) =>
        typeof update === "function" ? update(previous || []) : update,
      ),
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
    updateExercisePerformance,
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
    setTrainingView: () => setView("training"),
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
  const handleLogout = async () => {
    try {
      await logoutUser();
    } catch {
      // Clear local state even if the session already expired.
    }
    setCurrentUserId(null);
    setPrefetchedViews(new Set());
    setView("login");
    clearTraining();
  };

  useEffect(() => {
    const handleUnauthorized = () => {
      setCurrentUserId(null);
      setPrefetchedViews(new Set());
      clearTraining();
      setView("login");
    };
    window.addEventListener("motus:unauthorized", handleUnauthorized);
    return () =>
      window.removeEventListener("motus:unauthorized", handleUnauthorized);
  }, [clearTraining, setView]);

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
        onViewPreload={handleViewPreload}
        currentUser={currentUser}
        authHeaderEnabled={authHeaderEnabled}
        onLogout={
          !authHeaderEnabled && currentUserId ? handleLogout : undefined
        }
        resumeOpen={resumeOpen}
        resumeText={resumeMessage(training)}
        onResume={() => {
          setPromptedResume(false);
          setView("training");
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
                    setView("training");
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
                onCreateExercise: createExerciseEntry,
                promptUser: askPrompt,
                onToast: showToast,
              }}
            />
          )}

          {view === "training" && (
            <TrainingView
              data={{
                workouts: activeWorkouts,
                selectedWorkoutId,
                startDisabled: !selectedWorkoutId || !currentUserId,
                startTitle: !selectedWorkoutId
                  ? PROMPTS.selectWorkoutFirst
                  : "",
                training,
                currentStep,
                elapsed: displayedElapsed,
                workoutName: currentWorkoutName,
                sounds: sounds.data || [],
                pauseOnTabHidden,
                showHours,
                history: history.data || [],
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
                onSaveFeedback: async (trainingId, notes, perceivedEffort) => {
                  if (!currentUserId) return;
                  await updateTrainingFeedback(trainingId, {
                    userId: currentUserId,
                    notes,
                    perceivedEffort,
                  });
                  history.reload();
                },
                onUpdateExercisePerformance: updateExercisePerformance,
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
                onResume: () => setView("training"),
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
                loading: exercises.loading,
                isAdmin: Boolean(currentUser?.isAdmin),
              }}
              actions={{
                onAddExercise: handleAddExercise,
                onAddCoreExercise: handleAddCoreExercise,
                onRenameExercise: handleRenameExercise,
                onDeleteExercise: handleDeleteExercise,
                onToggleSides: handleToggleExerciseSides,
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
