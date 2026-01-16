import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  applyTemplate,
  getConfig,
  getCurrentUser,
  getWorkout,
  listExercises,
  listSessionHistory,
  listSounds,
  listTemplates,
  listUsers,
  listWorkouts,
  setAuthHeaderEnabled,
  updateUserName,
} from "./api";

import type {
  CatalogExercise,
  TrainHistoryItem,
  SoundOption,
  Template,
  ThemeMode,
  User,
  Workout,
} from "./types";

import { useSessionTimer } from "./hooks/useSessionTimer";
import { useDialog } from "./hooks/useDialog";
import { useViewState } from "./hooks/useViewState";

import { LoginView } from "./components/pages/LoginPage";
import { AdminView } from "./components/pages/AdminPage";
import { WorkoutsView } from "./components/pages/WorkoutsPage";
import { TrainView } from "./components/pages/TrainPage";
import { TemplatesView } from "./components/pages/TemplatesPage";
import { HistoryView } from "./components/pages/HistoryPage";
import { ProfileView } from "./components/pages/ProfilePage";
import { ExercisesView } from "./components/pages/ExercisesPage";

import { AppShell } from "./components/shell/AppShell";
import DialogModal from "./components/common/DialogModal";

import { isValidEmail } from "./utils/validation";
import { STEP_TYPE_PAUSE } from "./utils/step";

import { useAuthActions } from "./hooks/useAuthActions";
import { useAdminActions } from "./hooks/useAdminActions";
import { useExerciseActions } from "./hooks/useExerciseActions";
import { useProfileActions } from "./hooks/useProfileActions";
import { useSessionActions } from "./hooks/useSessionActions";

import "./styles.css";

type AppConfig = {
  authHeaderEnabled: boolean;
  allowRegistration: boolean;
  version: string;
  commit: string;
};

// useDataLoader wraps async loading with loading/error state.
function useDataLoader<T>(loader: () => Promise<T>, deps: unknown[] = []) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const reload = useCallback(() => {
    let cancelled = false;
    setLoading(true);

    loader()
      .then((res) => {
        if (cancelled) return;
        setData(res);
        setError(null);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err?.message || "load failed");
      })
      .finally(() => {
        if (cancelled) return;
        setLoading(false);
      });

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);

  useEffect(() => {
    const cancel = reload();
    return cancel;
  }, [reload]);

  return { data, loading, error, setData, reload };
}

// resumeMessage formats a resume prompt for an in-progress session.
function resumeMessage(
  session?: ReturnType<typeof useSessionTimer>["session"] | null,
) {
  if (!session) return "";
  const name = session.workoutName || "your workout";
  return `Resume ${name}?`;
}

export default function App() {
  const { view, setView } = useViewState("train");

  const [currentUserId, setCurrentUserId] = useState<string | null>(() => {
    const stored = localStorage.getItem("motus:userId");
    if (stored && isValidEmail(stored)) return stored;
    return null;
  });

  const [config, setConfig] = useState<AppConfig | null>(null);
  const [authError, setAuthError] = useState<string | null>(null);
  const [loginError, setLoginError] = useState<string | null>(null);

  const [themeMode, setThemeMode] = useState<ThemeMode>(() => {
    const stored = localStorage.getItem("motus:theme");
    if (stored === "dark" || stored === "light" || stored === "auto")
      return stored;
    return "auto";
  });

  // user defaults
  const [defaultStepSoundKey, setDefaultStepSoundKey] = useState("");
  const [defaultPauseDuration, setDefaultPauseDuration] = useState("");
  const [defaultPauseSoundKey, setDefaultPauseSoundKey] = useState("");
  const [defaultPauseAutoAdvance, setDefaultPauseAutoAdvance] = useState(false);
  const [repeatRestAfterLastDefault, setRepeatRestAfterLastDefault] =
    useState(false);
  const [pauseOnTabHidden, setPauseOnTabHidden] = useState(false);

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

  const allowRegistration = config?.allowRegistration ?? true;
  const authHeaderEnabled = config?.authHeaderEnabled ?? false;
  const appVersion = config?.version || "dev";

  // ---------- data loaders ----------
  const users = useDataLoader<User[]>(listUsers, []);
  const sounds = useDataLoader<SoundOption[]>(listSounds, []);
  const workouts = useDataLoader<Workout[]>(
    () => (currentUserId ? listWorkouts(currentUserId) : Promise.resolve([])),
    [currentUserId],
  );
  const history = useDataLoader<TrainHistoryItem[]>(
    () =>
      currentUserId
        ? listSessionHistory(currentUserId)
        : Promise.resolve([] as TrainHistoryItem[]),
    [currentUserId],
  );
  const templates = useDataLoader<Template[]>(listTemplates, []);

  const activeWorkouts = workouts.data || [];
  const currentUser = useMemo(
    () => users.data?.find((u) => u.id === currentUserId) || null,
    [users.data, currentUserId],
  );

  const currentWorkoutName = useMemo(() => {
    if (!selectedWorkoutId) return "";
    return activeWorkouts.find((w) => w.id === selectedWorkoutId)?.name || "";
  }, [selectedWorkoutId, activeWorkouts]);

  // ---------- config + proxy auth ----------
  useEffect(() => {
    getConfig()
      .then((cfg) => {
        setConfig(cfg);
        setAuthHeaderEnabled(cfg.authHeaderEnabled);

        if (!cfg.authHeaderEnabled) return;

        return getCurrentUser()
          .then((user) => {
            setCurrentUserId(user.id);
            setAuthError(null);
            if (view === "login") setView("train");
          })
          .catch((err: Error) => {
            setAuthError(err.message || "Unable to authenticate user");
          });
      })
      .catch((err: Error) => {
        setAuthError(err.message || "Unable to load configuration");
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // ---------- theme ----------
  useEffect(() => {
    const root = document.documentElement;

    const applyTheme = () => {
      if (themeMode === "auto") {
        const prefersDark = window.matchMedia(
          "(prefers-color-scheme: dark)",
        ).matches;
        root.dataset.theme = prefersDark ? "dark" : "light";
        return;
      }
      root.dataset.theme = themeMode;
    };

    localStorage.setItem("motus:theme", themeMode);
    applyTheme();

    if (themeMode !== "auto") return;
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = () => applyTheme();
    media.addEventListener("change", handler);
    return () => media.removeEventListener("change", handler);
  }, [themeMode]);

  // ---------- hydrate per-user defaults ----------
  useEffect(() => {
    if (!currentUserId) {
      setRepeatRestAfterLastDefault(false);
      setDefaultStepSoundKey("");
      setDefaultPauseDuration("");
      setDefaultPauseSoundKey("");
      setDefaultPauseAutoAdvance(false);
      setPauseOnTabHidden(false);
      return;
    }

    setRepeatRestAfterLastDefault(
      localStorage.getItem(`motus:repeatRestAfterLast:${currentUserId}`) ===
        "true",
    );
    setDefaultStepSoundKey(
      localStorage.getItem(`motus:defaultStepSound:${currentUserId}`) || "",
    );
    setDefaultPauseDuration(
      localStorage.getItem(`motus:defaultPauseDuration:${currentUserId}`) || "",
    );
    setDefaultPauseSoundKey(
      localStorage.getItem(`motus:defaultPauseSound:${currentUserId}`) || "",
    );
    setDefaultPauseAutoAdvance(
      localStorage.getItem(`motus:defaultPauseAuto:${currentUserId}`) ===
        "true",
    );
    setPauseOnTabHidden(
      localStorage.getItem(`motus:pauseOnHidden:${currentUserId}`) === "true",
    );
  }, [currentUserId]);

  // ---------- validate stored user id once local users are known ----------
  useEffect(() => {
    if (authHeaderEnabled) return;
    if (!users.data) return;
    if (currentUserId && users.data.find((u) => u.id === currentUserId)) return;

    if (currentUserId) {
      localStorage.removeItem("motus:userId");
      setCurrentUserId(null);
    }
  }, [authHeaderEnabled, users.data, currentUserId]);

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
    if (!users.data) return;
    if (!currentUserId && view !== "login") setView("login");
  }, [authHeaderEnabled, currentUserId, users.data, view, setView]);

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

  // ---------- session timer ----------
  const {
    session,
    currentStep,
    displayedElapsed,
    restoredFromStorage,
    startFromState,
    startCurrentStep,
    pause,
    nextStep,
    finishAndLog,
    markSoundPlayed,
    clear: clearSession,
  } = useSessionTimer({ currentUserId });

  const [promptedResume, setPromptedResume] = useState(false);
  const [resumeSuppressed, setResumeSuppressed] = useState(false);

  const {
    startSession: handleStartSession,
    finishSession: handleFinishSession,
  } = useSessionActions({
    selectedWorkoutId,
    session,
    currentWorkoutName,
    setTrainView: () => setView("train"),
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
    if (!restoredFromStorage || !session || session.done) return;
    if (promptedResume || resumeSuppressed) return;
    setPromptedResume(true);
  }, [restoredFromStorage, session, promptedResume, resumeSuppressed]);

  // reset resume suppression when session clears
  useEffect(() => {
    if (!session) setResumeSuppressed(false);
  }, [session?.sessionId]);

  // auto-advance for timed steps
  useEffect(() => {
    if (!session || !session.running) return;
    if (!currentStep || !currentStep.estimatedSeconds) return;

    const isAutoAdvance =
      (currentStep.type === STEP_TYPE_PAUSE &&
        currentStep.pauseOptions?.autoAdvance) ||
      Boolean(currentStep.autoAdvance);

    if (!isAutoAdvance) return;

    const threshold = currentStep.estimatedSeconds * 1000;
    if (displayedElapsed < threshold + 200) return;

    const isLast =
      session.currentIndex >=
      (session.steps?.length ? session.steps.length - 1 : 0);

    if (!isLast) {
      nextStep();
      return;
    }

    finishAndLog().then((result) => {
      if (!result?.ok) {
        notify(result?.error || "Unable to save session");
        return;
      }
      history.reload();
    });
  }, [
    session,
    currentStep,
    displayedElapsed,
    nextStep,
    finishAndLog,
    history,
    notify,
  ]);

  // refresh history once when a session logs
  useEffect(() => {
    if (session?.logged && session.sessionId !== historyReloadGuard.current) {
      historyReloadGuard.current = session.sessionId;
      history.reload();
    }
  }, [session?.logged, session?.sessionId, history]);

  // ---------- handlers for defaults ----------
  const handleRepeatRestAfterLastDefault = (value: boolean) => {
    setRepeatRestAfterLastDefault(value);
    if (!currentUserId) return;
    localStorage.setItem(
      `motus:repeatRestAfterLast:${currentUserId}`,
      value ? "true" : "false",
    );
  };

  const handleDefaultStepSound = (value: string) => {
    setDefaultStepSoundKey(value);
    if (!currentUserId) return;
    localStorage.setItem(`motus:defaultStepSound:${currentUserId}`, value);
  };

  const handleDefaultPauseDuration = (value: string) => {
    setDefaultPauseDuration(value);
    if (!currentUserId) return;
    localStorage.setItem(`motus:defaultPauseDuration:${currentUserId}`, value);
  };

  const handleDefaultPauseSound = (value: string) => {
    setDefaultPauseSoundKey(value);
    if (!currentUserId) return;
    localStorage.setItem(`motus:defaultPauseSound:${currentUserId}`, value);
  };

  const handleDefaultPauseAutoAdvance = (value: boolean) => {
    setDefaultPauseAutoAdvance(value);
    if (!currentUserId) return;
    localStorage.setItem(
      `motus:defaultPauseAuto:${currentUserId}`,
      value ? "true" : "false",
    );
  };

  const handlePauseOnTabHidden = (value: boolean) => {
    setPauseOnTabHidden(value);
    if (!currentUserId) return;
    localStorage.setItem(
      `motus:pauseOnHidden:${currentUserId}`,
      value ? "true" : "false",
    );
  };

  // ---------- logout ----------
  const handleLogout = () => {
    localStorage.removeItem("motus:userId");
    setCurrentUserId(null);
    setView("login");
    clearSession();
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
      } catch (err: any) {
        await notify(err?.message || "Unable to apply template");
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

  const resumeOpen = promptedResume && Boolean(session && !session.done);

  return (
    <>
      <AppShell
        view={view}
        onViewChange={setView}
        currentUser={currentUser}
        authHeaderEnabled={authHeaderEnabled}
        onLogout={
          !authHeaderEnabled && currentUserId ? handleLogout : undefined
        }
        resumeOpen={resumeOpen}
        resumeText={resumeMessage(session)}
        onResume={() => {
          setPromptedResume(false);
          setView("train");
          if (!session?.running) startCurrentStep();
          setToast(null);
          setResumeSuppressed(true);
        }}
        onDismissResume={() => {
          setPromptedResume(false);
          clearSession();
          setResumeSuppressed(true);
        }}
        toast={toast}
        appVersion={appVersion}
      >
        {view === "login" && !authHeaderEnabled && (
          <LoginView
            allowRegistration={allowRegistration}
            loginError={loginError}
            onLogin={handleLogin}
            onCreateUser={async (email, password) => {
              try {
                await handleRegister(email, password);
                setView("train");
              } catch (err: any) {
                await notify(err?.message || "Unable to create user");
              }
            }}
            onClearError={() => setLoginError(null)}
          />
        )}

        {view === "admin" && currentUser?.isAdmin && (
          <AdminView
            users={users.data || []}
            loading={users.loading}
            currentUserId={currentUserId}
            allowRegistration={allowRegistration}
            onToggleAdmin={handleToggleAdmin}
            onCreateUser={async (email, password) => {
              try {
                await handleRegister(email, password);
              } catch (err: any) {
                await notify(err?.message || "Unable to create user");
              }
            }}
            onBackfill={backfillCatalog}
          />
        )}

        {view === "workouts" && (
          <WorkoutsView
            workouts={activeWorkouts}
            loading={workouts.loading}
            setWorkouts={(updater) => workouts.setData?.(updater)}
            currentUserId={currentUserId}
            askConfirm={askConfirm}
            askPrompt={askPrompt}
            notifyUser={notify}
            templatesReload={() => templates.reload()}
            sounds={sounds.data || []}
            exerciseCatalog={exerciseCatalog}
            onCreateExercise={createExerciseEntry}
            promptUser={askPrompt}
            defaultStepSoundKey={defaultStepSoundKey}
            defaultPauseDuration={defaultPauseDuration}
            defaultPauseSoundKey={defaultPauseSoundKey}
            defaultPauseAutoAdvance={defaultPauseAutoAdvance}
            repeatRestAfterLastDefault={repeatRestAfterLastDefault}
            onToast={showToast}
          />
        )}

        {view === "train" && (
          <TrainView
            workouts={activeWorkouts}
            selectedWorkoutId={selectedWorkoutId}
            onSelectWorkout={setSelectedWorkoutId}
            onStartSession={handleStartSession}
            startDisabled={!selectedWorkoutId || !currentUserId}
            startTitle={!selectedWorkoutId ? "Select a workout first" : ""}
            session={session}
            currentStep={currentStep}
            elapsed={displayedElapsed}
            workoutName={currentWorkoutName}
            sounds={sounds.data || []}
            markSoundPlayed={markSoundPlayed}
            onStartStep={startCurrentStep}
            onPause={pause}
            onNext={nextStep}
            onFinishSession={handleFinishSession}
            onCopySummary={() => showToast("Copied summary")}
            onToast={showToast}
            pauseOnTabHidden={pauseOnTabHidden}
          />
        )}

        {view === "templates" && (
          <TemplatesView
            templates={templates.data || []}
            loading={templates.loading}
            hasUser={Boolean(currentUserId)}
            onRefresh={() => templates.reload()}
            onApplyTemplate={handleApplyTemplate}
          />
        )}

        {view === "history" && (
          <HistoryView
            items={history.data || []}
            activeSession={session}
            onResume={() => setView("train")}
            loadWorkout={getWorkout}
            onCopySummary={() => showToast("Copied summary")}
          />
        )}

        {view === "profile" && (
          <ProfileView
            profileTab={profileTab}
            onProfileTabChange={setProfileTab}
            currentName={currentUser?.name || ""}
            onUpdateName={handleUpdateName}
            themeMode={themeMode}
            onThemeChange={setThemeMode}
            sounds={sounds.data || []}
            defaultStepSoundKey={defaultStepSoundKey}
            onDefaultStepSoundChange={handleDefaultStepSound}
            defaultPauseDuration={defaultPauseDuration}
            onDefaultPauseDurationChange={handleDefaultPauseDuration}
            defaultPauseSoundKey={defaultPauseSoundKey}
            onDefaultPauseSoundChange={handleDefaultPauseSound}
            defaultPauseAutoAdvance={defaultPauseAutoAdvance}
            onDefaultPauseAutoAdvanceChange={handleDefaultPauseAutoAdvance}
            repeatRestAfterLastDefault={repeatRestAfterLastDefault}
            onRepeatRestAfterLastDefaultChange={
              handleRepeatRestAfterLastDefault
            }
            pauseOnTabHidden={pauseOnTabHidden}
            onPauseOnTabHiddenChange={handlePauseOnTabHidden}
            exportWorkoutId={exportWorkoutId}
            onExportWorkoutChange={setExportWorkoutId}
            activeWorkouts={activeWorkouts}
            onExportWorkout={handleExportSelected}
            onImportWorkout={handleImportSelected}
            onPasswordChange={handlePasswordSubmit}
            importInputRef={importInputRef}
            authHeaderEnabled={authHeaderEnabled}
          />
        )}

        {view === "exercises" && (
          <ExercisesView
            exercises={exerciseCatalog}
            isAdmin={Boolean(currentUser?.isAdmin)}
            onAddExercise={handleAddExercise}
            onAddCoreExercise={handleAddCoreExercise}
            onRenameExercise={handleRenameExercise}
            onDeleteExercise={handleDeleteExercise}
          />
        )}
      </AppShell>

      {dialog && (
        <DialogModal
          dialog={dialog}
          value={dialogValue}
          onValueChange={setDialogValue}
          onClose={closeDialog}
        />
      )}
    </>
  );
}
