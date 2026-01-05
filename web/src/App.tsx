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
} from "./api";
import { useSessionTimer } from "./hooks/useSessionTimer";
import { WorkoutForm } from "./components/WorkoutForm";
import { LoginView } from "./components/LoginView";
import { AdminView } from "./components/AdminView";
import { WorkoutsView } from "./components/WorkoutsView";
import { SessionsView } from "./components/SessionsView";
import { TemplatesView } from "./components/TemplatesView";
import { HistoryView } from "./components/HistoryView";
import { ProfileView } from "./components/ProfileView";
import { ExercisesView } from "./components/ExercisesView";
import { useAuthActions } from "./hooks/useAuthActions";
import { useAdminActions } from "./hooks/useAdminActions";
import { useExerciseActions } from "./hooks/useExerciseActions";
import { useWorkoutActions } from "./hooks/useWorkoutActions";
import { useTemplateActions } from "./hooks/useTemplateActions";
import { useProfileActions } from "./hooks/useProfileActions";
import { useWorkoutFormActions } from "./hooks/useWorkoutFormActions";
import { useSessionActions } from "./hooks/useSessionActions";
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
  "sessions",
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
  return "sessions";
}

// isValidEmail checks a basic email pattern for client validation.
function isValidEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
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
  return (
    <nav>
      {views.map((v) => (
        <button
          key={v}
          className={view === v ? "tab active" : "tab"}
          onClick={() => onSelect(v)}
        >
          {v.toUpperCase()}
        </button>
      ))}
    </nav>
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
  const [selectedWorkoutId, setSelectedWorkoutId] = useState<string | null>(
    null,
  );
  const [editingWorkout, setEditingWorkout] = useState<Workout | null>(null);
  const [showWorkoutForm, setShowWorkoutForm] = useState(false);
  const [workoutDirty, setWorkoutDirty] = useState(false);
  const historyReloadGuard = useRef<string | null>(null);
  const [exerciseCatalog, setExerciseCatalog] = useState<CatalogExercise[]>([]);
  const [promptedResume, setPromptedResume] = useState(false);
  const [dialog, setDialog] = useState<{
    type: "alert" | "confirm" | "prompt";
    message: string;
    title?: string;
    defaultValue?: string;
    placeholder?: string;
    resolve: (value: any) => void;
  } | null>(null);
  const [dialogValue, setDialogValue] = useState("");
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

useEffect(() => {
    // Load config and resolve proxy-auth user early.
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
                setView("sessions");
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

useEffect(() => {
    // Apply theme selection and react to system preference changes.
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

useEffect(() => {
    // Validate stored user id once local users are known.
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

useEffect(() => {
    // Persist the current view in the URL for refresh/bookmark.
    const params = new URLSearchParams(window.location.search);
    if (view === "sessions") {
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

  useEffect(() => {
    if (view === "login") setLoginError(null);
  }, [view]);

useEffect(() => {
    // Auto-redirect to sessions when a local user logs in.
    if (authHeaderEnabled) return;
    if (currentUserId && view === "login") {
      setView("sessions");
    }
  }, [authHeaderEnabled, currentUserId, view]);

useEffect(() => {
    // Keep the exercise catalog in sync with auth state.
    if (!authHeaderEnabled && !currentUserId) {
      setExerciseCatalog([]);
      return;
    }
    listExercises()
      .then((items) => setExerciseCatalog(items || []))
      .catch(() => {});
  }, [authHeaderEnabled, currentUserId]);

useEffect(() => {
    // Force login screen when local auth has no user.
    if (authHeaderEnabled) return;
    if (!users.data) return;
    if (!currentUserId && view !== "login") {
      setView("login");
    }
  }, [authHeaderEnabled, currentUserId, users.data, view]);

  // notify shows a basic alert dialog.
  const notify = useCallback(
    (message: string) =>
      new Promise<void>((resolve) => {
        setDialog({
          type: "alert",
          message,
          resolve: () => {
            resolve();
            setDialog(null);
          },
        });
        setDialogValue("");
      }),
    [],
  );

  // askConfirm shows a confirmation dialog.
  const askConfirm = useCallback(
    (message: string) =>
      new Promise<boolean>((resolve) => {
        setDialog({
          type: "confirm",
          message,
          resolve: (val: boolean) => {
            resolve(val);
            setDialog(null);
          },
        });
        setDialogValue("");
      }),
    [],
  );

  // askPrompt shows a prompt dialog.
  const askPrompt = useCallback(
    (message: string, defaultValue = "", placeholder = "") =>
      new Promise<string | null>((resolve) => {
        setDialog({
          type: "prompt",
          message,
          defaultValue,
          placeholder,
          resolve: (val: string | null) => {
            resolve(val);
            setDialog(null);
          },
        });
        setDialogValue(defaultValue);
      }),
    [],
  );

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

  // onLoginSuccess persists the authenticated user for local mode.
  const onLoginSuccess = (user: User) => {
    setCurrentUserId(user.id);
    localStorage.setItem("motus:userId", user.id);
    setView("sessions");
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
    setSessionsView: () => setView("sessions"),
    setPromptedResume,
    setResumeSuppressed,
    startFromState,
    finishAndLog,
    historyReload: () => history.reload(),
    askConfirm,
    notify,
  });

  useEffect(() => {
    // Prompt to resume a restored, unfinished session once.
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

  useEffect(() => {
    // Auto-advance timed or auto-pause steps when their target elapses.
    if (!session || !session.running) return;
    if (!currentStep || !currentStep.estimatedSeconds) return;
    const isTimed = currentStep.type === "timed";
    const isAutoPause =
      currentStep.type === "pause" && currentStep.pauseOptions?.autoAdvance;
    if (!isTimed && !isAutoPause) return;
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

  useEffect(() => {
    // Reset resume suppression when session ends/clears.
    if (!session) {
      setResumeSuppressed(false);
    }
  }, [session?.sessionId]);

  useEffect(() => {
    // Refresh history once when a session logs.
    if (session?.logged && session.sessionId !== historyReloadGuard.current) {
      historyReloadGuard.current = session.sessionId;
      history.reload();
    }
  }, [session?.logged, session?.sessionId, history]);

  if (!config) {
    return (
      <div className="shell">
        <header className="topbar">
          <h1>Motus</h1>
        </header>
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
        <header className="topbar">
          <h1>Motus</h1>
        </header>
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
              <strong>Resume session?</strong>
              <div className="muted small">{resumeMessage(session)}</div>
            </div>
            <div className="btn-group">
              <button
                className="btn primary"
                onClick={() => {
                  setPromptedResume(false);
                  setView("sessions");
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
          <h1>Motus</h1>
          {(authHeaderEnabled || currentUserId) && (
            <div className="topbar-actions">
              <NavTabs
                view={view}
                views={(
                  [
                    "sessions",
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
                  setView("sessions");
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
              currentUserName={currentUser?.name || "Unknown user"}
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
              authHeaderEnabled={authHeaderEnabled}
              currentUserName={currentUser?.name || "Unknown user"}
              hasUser={Boolean(currentUserId)}
              onNewWorkout={handleNewWorkout}
              onEditWorkout={handleEditWorkoutFromList}
              onShareTemplate={handleShareTemplate}
              onDeleteWorkout={handleDeleteWorkout}
            />
          )}

          {view === "sessions" && (
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
            />
          )}

          {view === "templates" && (
            <TemplatesView
              templates={templates.data || []}
              loading={templates.loading}
              currentUserName={currentUser?.name || "Unknown user"}
              authHeaderEnabled={authHeaderEnabled}
              hasUser={Boolean(currentUserId)}
              onRefresh={() => templates.reload()}
              onApplyTemplate={handleApplyTemplate}
            />
          )}

          {view === "history" && (
            <HistoryView
              items={history.data || []}
              activeSession={session}
              onResume={() => setView("sessions")}
              loadWorkout={getWorkout}
              onCopySummary={() => showToast("Copied summary")}
            />
          )}

          {view === "profile" && (
            <ProfileView
              profileTab={profileTab}
              onProfileTabChange={setProfileTab}
              themeMode={themeMode}
              onThemeChange={setThemeMode}
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

      {showWorkoutForm && (
        <div className="modal-overlay" onClick={handleCloseWorkoutModal}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <WorkoutForm
              userId={currentUserId}
              onSave={handleSaveWorkout}
              onUpdate={handleUpdateWorkout}
              editingWorkout={editingWorkout}
              sounds={sounds.data || []}
              exerciseCatalog={exerciseCatalog}
              onCreateExercise={createExerciseEntry}
              onClose={handleCloseWorkoutModal}
              promptUser={askPrompt}
              notifyUser={notify}
              onDirtyChange={setWorkoutDirty}
              onToast={showToast}
            />
          </div>
        </div>
      )}
      {dialog && (
        <div
          className="modal-overlay"
          onClick={() => {
            if (dialog.type === "confirm") {
              dialog.resolve(false);
            } else if (dialog.type === "prompt") {
              dialog.resolve(null);
            } else {
              dialog.resolve(undefined);
            }
            setDialog(null);
          }}
        >
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>
              {dialog.title ||
                (dialog.type === "confirm" ? "Confirm" : "Message")}
            </h3>
            <p className="muted">{dialog.message}</p>
            {dialog.type === "prompt" && (
              <input
                autoFocus
                value={dialogValue}
                placeholder={dialog.placeholder || ""}
                onChange={(e) => setDialogValue(e.target.value)}
              />
            )}
            <div className="btn-group" style={{ justifyContent: "flex-end" }}>
              {dialog.type !== "alert" && (
                <button
                  className="btn subtle"
                  onClick={() => {
                    if (dialog.type === "confirm") {
                      dialog.resolve(false);
                    } else {
                      dialog.resolve(null);
                    }
                    setDialog(null);
                  }}
                >
                  Cancel
                </button>
              )}
              <button
                className="btn primary"
                onClick={() => {
                  if (dialog.type === "confirm") {
                    dialog.resolve(true);
                  } else if (dialog.type === "prompt") {
                    dialog.resolve(dialogValue);
                  } else {
                    dialog.resolve(undefined);
                  }
                  setDialog(null);
                }}
              >
                {dialog.type === "confirm"
                  ? "Confirm"
                  : dialog.type === "prompt"
                    ? "Save"
                    : "OK"}
              </button>
            </div>
          </div>
        </div>
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
