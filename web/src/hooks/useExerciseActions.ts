import { useCallback } from "react";
import { MESSAGES, toErrorMessage } from "../utils/messages";
import { UI_TEXT } from "../utils/uiText";
import type { Dispatch, SetStateAction } from "react";

import { createExercise, deleteExercise, updateExercise } from "../api";
import type { AskConfirmOptions, CatalogExercise } from "../types";

// UseExerciseActionsArgs wires exercise management actions.
type UseExerciseActionsArgs = {
  isAdmin: boolean;
  setExerciseCatalog: Dispatch<SetStateAction<CatalogExercise[]>>;
  askPrompt: (message: string, defaultValue?: string) => Promise<string | null>;
  askConfirm: (
    message: string,
    options?: AskConfirmOptions,
  ) => Promise<boolean>;
  notify: (message: string) => Promise<void>;
  showToast: (message: string) => void;
};

// useExerciseActions provides exercise CRUD handlers.
export function useExerciseActions({
  isAdmin,
  setExerciseCatalog,
  askPrompt,
  askConfirm,
  notify,
  showToast,
}: UseExerciseActionsArgs) {
  // createCatalogEntry adds a new exercise and stores it in state.
  const createCatalogEntry = useCallback(
    async (name: string, isCore: boolean) => {
      // Guard: require a non-empty name.
      if (!name.trim()) throw new Error(UI_TEXT.errors.exerciseNameRequired);
      const created = await createExercise(name.trim(), isCore);
      setExerciseCatalog((prev) => {
        // Avoid duplicates when optimistic UI already has the entry.
        const exists = prev.find((e) => e.id === created.id);
        if (exists) return prev;
        return [...prev, created];
      });
      return created;
    },
    [setExerciseCatalog],
  );

  // addExercise prompts for a new personal exercise.
  const addExercise = useCallback(async () => {
    const name = await askPrompt(UI_TEXT.prompts.exerciseName);
    if (!name) return;
    try {
      await createCatalogEntry(name, false);
    } catch (err) {
      await notify(toErrorMessage(err, MESSAGES.createExerciseFailed));
    }
  }, [askPrompt, createCatalogEntry, notify]);

  // addCoreExercise prompts for a new core exercise.
  const addCoreExercise = useCallback(async () => {
    const name = await askPrompt(UI_TEXT.prompts.coreExerciseName);
    if (!name) return;
    try {
      await createCatalogEntry(name, true);
    } catch (err) {
      await notify(toErrorMessage(err, MESSAGES.createCoreExerciseFailed));
    }
  }, [askPrompt, createCatalogEntry, notify]);

  // renameExercise renames or copies an exercise.
  const renameExercise = useCallback(
    async (ex: CatalogExercise) => {
      const name = await askPrompt(
        ex.isCore && !isAdmin
          ? UI_TEXT.exercises.createCopy
          : UI_TEXT.exercises.rename,
        ex.name,
      );
      if (!name || name.trim() === ex.name) return;
      try {
        const updated = await updateExercise(ex.id, name.trim());
        setExerciseCatalog((prev) => {
          // Copy-on-write for core exercises when needed.
          const existing = prev.find((entry) => entry.id === updated.id);
          if (updated.id !== ex.id && !existing) {
            return [...prev, updated];
          }
          return prev.map((entry) =>
            entry.id === updated.id ? { ...entry, ...updated } : entry,
          );
        });
        if (updated.id !== ex.id) {
          showToast(UI_TEXT.toasts.createdPersonalCopy);
        }
      } catch (err) {
        await notify(toErrorMessage(err, MESSAGES.renameExerciseFailed));
      }
    },
    [askPrompt, isAdmin, notify, setExerciseCatalog, showToast],
  );

  // deleteExerciseEntry removes an exercise after confirmation.
  const deleteExerciseEntry = useCallback(
    async (ex: CatalogExercise) => {
      const ok = await askConfirm(UI_TEXT.prompts.deleteExerciseConfirm);
      if (!ok) return;
      try {
        await deleteExercise(ex.id);
        setExerciseCatalog((prev) => prev.filter((e) => e.id !== ex.id));
      } catch (err) {
        await notify(toErrorMessage(err, MESSAGES.deleteExerciseFailed));
      }
    },
    [askConfirm, notify, setExerciseCatalog],
  );

  // createExerciseEntry creates a personal exercise entry.
  const createExerciseEntry = useCallback(
    (name: string) => createCatalogEntry(name, false),
    [createCatalogEntry],
  );

  return {
    createExerciseEntry,
    addExercise,
    addCoreExercise,
    renameExercise,
    deleteExerciseEntry,
  };
}
