import { useCallback, useEffect, useMemo, useRef, useState } from "react";
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

import { LoginView } from "./components/pages/LoginPage";
import { AdminView } from "./components/pages/AdminPage";
import { WorkoutsView } from "./components/pages/WorkoutsPage";
import { TrainingView } from "./components/pages/TrainingPage";
import { TemplatesView } from "./components/pages/TemplatesPage";
import { HistoryView } from "./components/pages/HistoryPage";
import { ProfileView } from "./components/pages/ProfilePage";
import { ExercisesView } from "./components/pages/ExercisesPage";

import { AppShell } from "./components/shell/AppShell";
import DialogModal from "./components/common/DialogModal";

import { isValidEmail } from "./utils/validation";
import { toErrorMessage } from "./utils/messages";

import { useAuthActions } from "./hooks/useAuthActions";
import { useAdminActions } from "./hooks/useAdminActions";
import { useExerciseActions } from "./hooks/useExerciseActions";
import { useProfileActions } from "./hooks/useProfileActions";
import { useTrainingActions } from "./hooks/useTrainingActions";

import "./styles.css";

// resumeMessage formats a resume prompt for an in-progress training.
function resumeMessage(
  training?: ReturnType<typeof useTrainingTimer>["training"] | null,
) {
  if (!training) return "";
  const name = training.workoutName || "your workout";
  return `Resume ${name}?`;
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
  const {
    users,
    sounds,
    workouts,
    history,
    templates,
    activeWorkouts,
    currentUser,
  } = useWorkoutsData({ currentUserId });
  const {
    defaultStepSoundKey,
    defaultPauseDuration,
    defaultPauseSoundKey,
    defaultPauseAutoAdvance,
    repeatRestAfterLastDefault,
    pauseOnTabHidden,
    updateRepeatRestAfterLastDefault,
    updateDefaultStepSoundKey,
    updateDefaultPauseDuration,
    updateDefaultPauseSoundKey,
    updateDefaultPauseAutoAdvance,
    updatePauseOnTabHidden,
  } = useUserDefaults({ currentUserId });

  const allowRegistration = config?.allowRegistration ?? true;
  const authHeaderEnabled = config?.authHeaderEnabled ?? false;
  const appVersion = config?.version || "dev";

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
        {view === "login" && !authHeaderEnabled && (
          <LoginView
            allowRegistration={allowRegistration}
            loginError={loginError}
            onLogin={handleLogin}
            onCreateUser={async (email, password) => {
              try {
                await handleRegister(email, password);
                setView("train");
              } catch (err) {
                await notify(toErrorMessage(err, "Unable to create user"));
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
              } catch (err) {
                await notify(toErrorMessage(err, "Unable to create user"));
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
          <TrainingView
            workouts={activeWorkouts}
            selectedWorkoutId={selectedWorkoutId}
            onSelectWorkout={setSelectedWorkoutId}
            onStartTraining={handleStartTraining}
            startDisabled={!selectedWorkoutId || !currentUserId}
            startTitle={!selectedWorkoutId ? "Select a workout first" : ""}
            training={training}
            currentStep={currentStep}
            elapsed={displayedElapsed}
            workoutName={currentWorkoutName}
            sounds={sounds.data || []}
            markSoundPlayed={markSoundPlayed}
            onStartStep={startCurrentStep}
            onPause={pause}
            onNext={nextStep}
            onFinishTraining={handleFinishTraining}
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
            activeTraining={training}
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
            onDefaultStepSoundChange={updateDefaultStepSoundKey}
            defaultPauseDuration={defaultPauseDuration}
            onDefaultPauseDurationChange={updateDefaultPauseDuration}
            defaultPauseSoundKey={defaultPauseSoundKey}
            onDefaultPauseSoundChange={updateDefaultPauseSoundKey}
            defaultPauseAutoAdvance={defaultPauseAutoAdvance}
            onDefaultPauseAutoAdvanceChange={updateDefaultPauseAutoAdvance}
            repeatRestAfterLastDefault={repeatRestAfterLastDefault}
            onRepeatRestAfterLastDefaultChange={
              updateRepeatRestAfterLastDefault
            }
            pauseOnTabHidden={pauseOnTabHidden}
            onPauseOnTabHiddenChange={updatePauseOnTabHidden}
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
