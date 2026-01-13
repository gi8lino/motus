import { useEffect, useMemo, useState, useCallback, useRef } from "react";
import {
  listSessionHistory,
  listSounds,
  listUsers,
  listWorkouts,
  getWorkout,
  listExercises,
  listTemplates,
  getConfig,
  getCurrentUser,
  setAuthHeaderEnabled,
  updateUserName,
} from "./api";
import { useSessionTimer } from "./hooks/useSessionTimer";
import { LoginView } from "./components/auth/LoginView";
import { AdminView } from "./components/admin/AdminView";
import { WorkoutsView } from "./components/workouts/WorkoutsView";
import { SessionsView } from "./components/sessions/SessionsView";
import { TemplatesView } from "./components/templates/TemplatesView";
import { HistoryView } from "./components/history/HistoryView";
import { ProfileView } from "./components/auth/ProfileView";
import { ExercisesView } from "./components/exercises/ExercisesView";
import { BrandHeader } from "./components/common/BrandHeader";
import DialogModal from "./components/common/DialogModal";
import { isValidEmail } from "./utils/validation";
import { useAuthActions } from "./hooks/useAuthActions";
import { useAdminActions } from "./hooks/useAdminActions";
import { useExerciseActions } from "./hooks/useExerciseActions";
import { useWorkoutActions } from "./hooks/useWorkoutActions";
import { useTemplateActions } from "./hooks/useTemplateActions";
import { useProfileActions } from "./hooks/useProfileActions";
import { useWorkoutFormActions } from "./hooks/useWorkoutFormActions";
import { useSessionActions } from "./hooks/useSessionActions";
import { useDialog } from "./hooks/useDialog";
import { STEP_TYPE_PAUSE } from "./utils/step";
import type {
  CatalogExercise,
  SessionHistoryItem,
  SoundOption,
  User,
  Workout,
  Template,
  View,
} from "./types";
import "./styles.css";

type ThemeMode = "auto" | "dark" | "light";

const VIEW_PARAM = "view";

const viewOptions: View[] = [
  "train",
  "login",
  "workouts",
  "profile",
  "history",
  "exercises",
  "templates",
  "admin",
];

// isValidView checks if a string is a known view identifier.
function isValidView(value: string): value is View {
  return viewOptions.includes(value as View);
}

// initialViewFromURL picks the initial view from the URL query parameter.
function initialViewFromURL(): View {
  const params = new URLSearchParams(window.location.search);
  const rawView = params.get(VIEW_PARAM);
  if (rawView && isValidView(rawView)) {
    return rawView;
  }
  return "train";
}

// NavTabs renders the main navigation tabs.
function NavTabs({
  view,
  views,
  onSelect,
}: {
  view: View;
  views: View[];
  onSelect: (next: View) => void;
}) {
  const [open, setOpen] = useState(false);
  const labels: Record<View, string> = {
    login: "Login",
    train: "Train",
    workouts: "Workouts",
    templates: "Templates",
    exercises: "Exercises",
    history: "History",
    profile: "Profile",
    admin: "Admin",
  };
  // Render the main shell with resume prompt, navigation, and active view.
  return (
    <div className="nav-shell">
      <button
        className="btn icon nav-toggle"
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        aria-label="Toggle navigation"
      >
        ☰
      </button>
      <nav className={open ? "nav-menu open" : "nav-menu"}>
        {views.map((v) => (
          <button
            key={v}
            className={view === v ? "tab active" : "tab"}
            onClick={() => {
              onSelect(v);
              setOpen(false);
            }}
          >
            {labels[v]}
          </button>
        ))}
      </nav>
    </div>
  );
}

// resumeMessage formats a resume prompt for an in-progress session.
function resumeMessage(
  session?: ReturnType<typeof useSessionTimer>["session"] | null,
) {
  if (!session) return "";
  const name = session.workoutName || "your workout";
  return `Resume ${name}?`;
}

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
        if (!cancelled) {
          setData(res);
          setError(null);
        }
      })
      .catch((err) => {
        if (!cancelled) setError(err.message || "load failed");
      })
      .finally(() => !cancelled && setLoading(false));
    return () => {
      cancelled = true;
    };
  }, deps); // eslint-disable-line react-hooks/exhaustive-deps

  // Load data when the loader or dependencies change.
  useEffect(() => {
    const cancel = reload();
    return cancel;
  }, [reload]);

  return { data, loading, error, setData, reload };
}

// App orchestrates the SPA data and views.
export default function App() {
  const [view, setView] = useState<View>(() => initialViewFromURL());
  const [currentUserId, setCurrentUserId] = useState<string | null>(() => {
    const stored = localStorage.getItem("motus:userId");
    if (stored && isValidEmail(stored)) {
      return stored;
    }
    return null;
  });
  const [config, setConfig] = useState<{
    authHeaderEnabled: boolean;
    allowRegistration: boolean;
    version: string;
    commit: string;
  } | null>(null);
  const [authError, setAuthError] = useState<string | null>(null);
  const [loginError, setLoginError] = useState<string | null>(null);
  const [themeMode, setThemeMode] = useState<ThemeMode>(() => {
    const stored = localStorage.getItem("motus:theme");
    if (stored === "dark" || stored === "light" || stored === "auto") {
      return stored;
    }
    return "auto";
  });
  const [defaultStepSoundKey, setDefaultStepSoundKey] = useState("");
  const [defaultPauseDuration, setDefaultPauseDuration] = useState("");
  const [defaultPauseSoundKey, setDefaultPauseSoundKey] = useState("");
  const [defaultPauseAutoAdvance, setDefaultPauseAutoAdvance] = useState(false);
  const [repeatRestAfterLastDefault, setRepeatRestAfterLastDefault] =
    useState(false);
  const [pauseOnTabHidden, setPauseOnTabHidden] = useState(false);
  const [selectedWorkoutId, setSelectedWorkoutId] = useState<string | null>(
    null,
  );
  const [editingWorkout, setEditingWorkout] = useState<Workout | null>(null);
  const [showWorkoutForm, setShowWorkoutForm] = useState(false);
  const [workoutDirty, setWorkoutDirty] = useState(false);
  const historyReloadGuard = useRef<string | null>(null);
  const [exerciseCatalog, setExerciseCatalog] = useState<CatalogExercise[]>([]);
  const [promptedResume, setPromptedResume] = useState(false);
  const {
    dialog,
    dialogValue,
    setDialogValue,
    closeDialog,
    notify,
    askConfirm,
    askPrompt,
  } = useDialog();
  const [toast, setToast] = useState<string | null>(null);
  const [resumeSuppressed, setResumeSuppressed] = useState(false);
  const importInputRef = useRef<HTMLInputElement | null>(null);
  const [exportWorkoutId, setExportWorkoutId] = useState("");
  const [profileTab, setProfileTab] = useState<
    "settings" | "password" | "transfer"
  >("settings");

  const allowRegistration = config?.allowRegistration ?? true;
  const authHeaderEnabled = config?.authHeaderEnabled ?? false;
  const appVersion = config?.version || "dev";

  const users = useDataLoader<User[]>(listUsers, []);
  const sounds = useDataLoader<SoundOption[]>(listSounds, []);
  const workouts = useDataLoader<Workout[]>(
    () => (currentUserId ? listWorkouts(currentUserId) : Promise.resolve([])),
    [currentUserId],
  );
  const history = useDataLoader<SessionHistoryItem[]>(
    () =>
      currentUserId
        ? listSessionHistory(currentUserId)
        : Promise.resolve([] as SessionHistoryItem[]),
    [currentUserId],
  );
  const templates = useDataLoader<Template[]>(listTemplates, []);

  // Load config and resolve proxy-auth user early.
  useEffect(() => {
    getConfig()
      .then((cfg) => {
        setConfig(cfg);
        setAuthHeaderEnabled(cfg.authHeaderEnabled);
        if (cfg.authHeaderEnabled) {
          return getCurrentUser()
            .then((user) => {
              setCurrentUserId(user.id);
              setAuthError(null);
              if (view === "login") {
                setView("train");
              }
            })
            .catch((err: Error) => {
              setAuthError(err.message || "Unable to authenticate user");
            });
        }
        return undefined;
      })
      .catch((err: Error) => {
        setAuthError(err.message || "Unable to load configuration");
      });
  }, []);

  // Apply theme selection and react to system preference changes.
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
    return () => {
      media.removeEventListener("change", handler);
    };
  }, [themeMode]);

  // Hydrate user-specific defaults from local storage.
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
    const stored = localStorage.getItem(
      `motus:repeatRestAfterLast:${currentUserId}`,
    );
    setRepeatRestAfterLastDefault(stored === "true");
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

  // handleRepeatRestAfterLastDefault persists the profile default for rest-after-last.
  const handleRepeatRestAfterLastDefault = (value: boolean) => {
    setRepeatRestAfterLastDefault(value);
    if (!currentUserId) return;
    localStorage.setItem(
      `motus:repeatRestAfterLast:${currentUserId}`,
      value ? "true" : "false",
    );
  };

  // handleDefaultStepSound persists the default sound for new steps.
  const handleDefaultStepSound = (value: string) => {
    setDefaultStepSoundKey(value);
    if (!currentUserId) return;
    localStorage.setItem(`motus:defaultStepSound:${currentUserId}`, value);
  };

  // handleDefaultPauseDuration persists the default pause duration string.
  const handleDefaultPauseDuration = (value: string) => {
    setDefaultPauseDuration(value);
    if (!currentUserId) return;
    localStorage.setItem(`motus:defaultPauseDuration:${currentUserId}`, value);
  };

  // handleDefaultPauseSound persists the default pause sound selection.
  const handleDefaultPauseSound = (value: string) => {
    setDefaultPauseSoundKey(value);
    if (!currentUserId) return;
    localStorage.setItem(`motus:defaultPauseSound:${currentUserId}`, value);
  };

  // handleDefaultPauseAutoAdvance persists the default pause auto-advance.
  const handleDefaultPauseAutoAdvance = (value: boolean) => {
    setDefaultPauseAutoAdvance(value);
    if (!currentUserId) return;
    localStorage.setItem(
      `motus:defaultPauseAuto:${currentUserId}`,
      value ? "true" : "false",
    );
  };

  // handlePauseOnTabHidden persists the session pause-on-hidden preference.
  const handlePauseOnTabHidden = (value: boolean) => {
    setPauseOnTabHidden(value);
    if (!currentUserId) return;
    localStorage.setItem(
      `motus:pauseOnHidden:${currentUserId}`,
      value ? "true" : "false",
    );
  };

  // Validate stored user id once local users are known.
  useEffect(() => {
    if (authHeaderEnabled) return;
    if (!users.data) return;
    if (currentUserId && users.data.find((u) => u.id === currentUserId)) {
      return;
    }
    if (currentUserId) {
      localStorage.removeItem("motus:userId");
      setCurrentUserId(null);
    }
  }, [authHeaderEnabled, users.data, currentUserId]);

  // Persist the current view in the URL for refresh/bookmark.
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (view === "train") {
      if (!params.has(VIEW_PARAM)) return;
      params.delete(VIEW_PARAM);
    } else {
      params.set(VIEW_PARAM, view);
    }
    const next = params.toString();
    const url = next
      ? `${window.location.pathname}?${next}${window.location.hash}`
      : `${window.location.pathname}${window.location.hash}`;
    window.history.replaceState({}, "", url);
  }, [view]);

  // Clear login errors when leaving the login view.
  useEffect(() => {
    if (view === "login") setLoginError(null);
  }, [view]);

  // Auto-redirect to sessions when a local user logs in.
  useEffect(() => {
    if (authHeaderEnabled) return;
    if (currentUserId && view === "login") {
      setView("train");
    }
  }, [authHeaderEnabled, currentUserId, view]);

  // Keep the exercise catalog in sync with auth state.
  useEffect(() => {
    if (!authHeaderEnabled && !currentUserId) {
      setExerciseCatalog([]);
      return;
    }
    listExercises()
      .then((items) => setExerciseCatalog(items || []))
      .catch(() => {});
  }, [authHeaderEnabled, currentUserId]);

  // Force login screen when local auth has no user.
  useEffect(() => {
    if (authHeaderEnabled) return;
    if (!users.data) return;
    if (!currentUserId && view !== "login") {
      setView("login");
    }
  }, [authHeaderEnabled, currentUserId, users.data, view]);

  // showToast shows a toast notification.
  const showToast = useCallback((message: string) => {
    setToast(message);
    setTimeout(() => setToast((t) => (t === message ? null : t)), 1800);
  }, []);

  // handleLogout clears local auth state and returns to login.
  const handleLogout = () => {
    localStorage.removeItem("motus:userId");
    setCurrentUserId(null);
    setView("login");
    clear();
  };

  const activeWorkouts = workouts.data || [];
  const currentUser = useMemo(
    // Resolve the selected user object from the list.
    () => users.data?.find((u) => u.id === currentUserId) || null,
    [users.data, currentUserId],
  );

  // handleUpdateName saves the display name for the signed-in user.
  const handleUpdateName = useCallback(
    async (name: string) => {
      if (!currentUserId) {
        throw new Error("No active user");
      }
      await updateUserName(name);
      users.setData?.((prev) =>
        prev
          ? prev.map((u) => (u.id === currentUserId ? { ...u, name } : u))
          : prev,
      );
    },
    [currentUserId, users],
  );

  // onLoginSuccess persists the authenticated user for local mode.
  const onLoginSuccess = (user: User) => {
    setCurrentUserId(user.id);
    localStorage.setItem("motus:userId", user.id);
    setView("train");
  };

  // onRegisterSuccess stores the newly created local user.
  const onRegisterSuccess = (user: User) => {
    users.setData?.((prev) => (prev ? [...prev, user] : [user]));
    setCurrentUserId(user.id);
    if (!authHeaderEnabled) {
      localStorage.setItem("motus:userId", user.id);
    }
  };

  const { login: handleLogin, register: handleRegister } = useAuthActions({
    setLoginError,
    onLoginSuccess,
    onRegisterSuccess,
  });

  const { toggleAdmin: handleToggleAdmin, backfillCatalog } = useAdminActions({
    currentUserId,
    setUsers: (updater) => users.setData?.(updater),
    setView,
    notify,
  });

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

  const {
    newWorkout: handleNewWorkout,
    editWorkoutFromList: handleEditWorkoutFromList,
    removeWorkout: handleDeleteWorkout,
    shareWorkout: handleShareTemplate,
  } = useWorkoutActions({
    workouts: activeWorkouts,
    editingWorkout,
    selectedWorkoutId,
    setEditingWorkout,
    setShowWorkoutForm,
    setSelectedWorkoutId,
    setWorkouts: (updater) => workouts.setData?.(updater),
    askConfirm,
    askPrompt,
    notify,
    templatesReload: () => templates.reload(),
  });

  const { applyTemplateToUser: handleApplyTemplate } = useTemplateActions({
    currentUserId,
    setWorkouts: (updater) => workouts.setData?.(updater),
    setSelectedWorkoutId,
    setShowWorkoutForm,
    setView,
    askPrompt,
    notify,
  });

  const {
    exportSelectedWorkout: handleExportSelected,
    importWorkoutFile: handleImportSelected,
    updatePassword: handlePasswordSubmit,
  } = useProfileActions({
    currentUserId,
    exportWorkoutId,
    setSelectedWorkoutId: (id) => setSelectedWorkoutId(id),
    setWorkouts: (updater) => workouts.setData?.(updater),
    showToast,
    notify,
  });

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
    clear,
  } = useSessionTimer({
    currentUserId,
  });

  const currentWorkoutName = useMemo(() => {
    if (!selectedWorkoutId) return "";
    return activeWorkouts.find((w) => w.id === selectedWorkoutId)?.name || "";
  }, [selectedWorkoutId, activeWorkouts]);

  const {
    saveWorkout: handleSaveWorkout,
    updateWorkout: handleUpdateWorkout,
    closeWorkoutModal: handleCloseWorkoutModal,
  } = useWorkoutFormActions({
    currentUserId,
    workoutDirty,
    setSelectedWorkoutId,
    setEditingWorkout,
    setWorkoutDirty,
    setShowWorkoutForm,
    setWorkouts: (updater) => workouts.setData?.(updater),
    reloadWorkouts: () => workouts.reload?.(),
    askConfirm,
  });

  const {
    startSession: handleStartSession,
    finishSession: handleFinishSession,
  } = useSessionActions({
    selectedWorkoutId,
    session,
    currentWorkoutName,
    setSessionsView: () => setView("train"),
    setPromptedResume,
    setResumeSuppressed,
    startFromState,
    finishAndLog,
    historyReload: () => history.reload(),
    askConfirm,
    notify,
  });

  // Prompt to resume a restored, unfinished session once.
  useEffect(() => {
    if (
      !restoredFromStorage ||
      !session ||
      session.done ||
      promptedResume ||
      resumeSuppressed
    )
      return;
    setPromptedResume(true);
  }, [restoredFromStorage, session, promptedResume, resumeSuppressed]);

  // Auto-advance timed exercises or auto-pause steps when their target elapses.
  useEffect(() => {
    if (!session || !session.running) return;
    if (!currentStep || !currentStep.estimatedSeconds) return;
    const isAutoAdvance =
      (currentStep.type === STEP_TYPE_PAUSE &&
        currentStep.pauseOptions?.autoAdvance) ||
      Boolean(currentStep.autoAdvance);
    if (!isAutoAdvance) return;
    const threshold = currentStep.estimatedSeconds * 1000;
    if (displayedElapsed >= threshold + 200) {
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
    }
  }, [
    session,
    currentStep,
    displayedElapsed,
    nextStep,
    finishAndLog,
    history,
    notify,
  ]);

  // Reset resume suppression when session ends/clears.
  useEffect(() => {
    if (!session) {
      setResumeSuppressed(false);
    }
  }, [session?.sessionId]);

  // Refresh history once when a session logs.
  useEffect(() => {
    if (session?.logged && session.sessionId !== historyReloadGuard.current) {
      historyReloadGuard.current = session.sessionId;
      history.reload();
    }
  }, [session?.logged, session?.sessionId, history]);

  // Guard: wait for config to avoid rendering with missing auth/base settings.
  if (!config) {
    return (
      <div className="shell">
        <BrandHeader />
        <main>
          <section className="panel">
            <p className="muted">Loading configuration…</p>
          </section>
        </main>
      </div>
    );
  }

  // Guard: show an explicit error page when auth proxy headers are missing.
  if (authHeaderEnabled && authError) {
    return (
      <div className="shell">
        <BrandHeader />
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

  return (
    <>
      <div className="shell">
        {promptedResume && session && !session.done && (
          <div className="toast">
            <div>
              <strong>Resume training?</strong>
              <div className="muted small">{resumeMessage(session)}</div>
            </div>
            <div className="btn-group">
              <button
                className="btn primary"
                onClick={() => {
                  setPromptedResume(false);
                  setView("train");
                  if (!session.running) {
                    startCurrentStep();
                  }
                  setToast(null);
                  setResumeSuppressed(true);
                }}
              >
                Resume
              </button>
              <button
                className="btn subtle"
                onClick={() => {
                  setPromptedResume(false);
                  clear();
                  setResumeSuppressed(true);
                }}
              >
                Dismiss
              </button>
            </div>
          </div>
        )}
        <header className="topbar">
          <BrandHeader />
          {(authHeaderEnabled || currentUserId) && (
            <div className="topbar-actions">
              <NavTabs
                view={view}
                views={(
                  [
                    "train",
                    "workouts",
                    "templates",
                    "exercises",
                    "history",
                    "profile",
                    "admin",
                  ] as View[]
                ).filter((v) => (v === "admin" ? currentUser?.isAdmin : true))}
                onSelect={setView}
              />
              {!authHeaderEnabled && currentUserId && (
                <button className="btn subtle" onClick={handleLogout}>
                  Logout
                </button>
              )}
            </div>
          )}
        </header>

        <main>
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
                  await notify(err.message || "Unable to create user");
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
                  await notify(err.message || "Unable to create user");
                }
              }}
              onBackfill={backfillCatalog}
            />
          )}

          {view === "workouts" && (
            <WorkoutsView
              workouts={activeWorkouts}
              loading={workouts.loading}
              hasUser={Boolean(currentUserId)}
              onNewWorkout={handleNewWorkout}
              onEditWorkout={handleEditWorkoutFromList}
              onShareTemplate={handleShareTemplate}
              onDeleteWorkout={handleDeleteWorkout}
              workoutForm={{
                open: showWorkoutForm,
                onClose: handleCloseWorkoutModal,
                userId: currentUserId,
                onSave: handleSaveWorkout,
                onUpdate: handleUpdateWorkout,
                editingWorkout,
                sounds: sounds.data || [],
                exerciseCatalog,
                onCreateExercise: createExerciseEntry,
                promptUser: askPrompt,
                notifyUser: notify,
                defaultStepSoundKey,
                defaultPauseDuration,
                defaultPauseSoundKey,
                defaultPauseAutoAdvance,
                repeatRestAfterLastDefault,
                onDirtyChange: setWorkoutDirty,
                onToast: showToast,
              }}
            />
          )}

          {view === "train" && (
            <SessionsView
              workouts={activeWorkouts}
              selectedWorkoutId={selectedWorkoutId}
              onSelectWorkout={(id) => setSelectedWorkoutId(id)}
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
        </main>
      </div>

      {dialog && (
        <DialogModal
          dialog={dialog}
          value={dialogValue}
          onValueChange={setDialogValue}
          onClose={closeDialog}
        />
      )}
      {toast && <div className="toast-floating">{toast}</div>}
      <footer className="app-footer">
        <div className="app-footer-inner">
          <span>© 2025 Motus | Version: {appVersion}</span>
        </div>
      </footer>
    </>
  );
}
