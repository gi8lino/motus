import { useCallback, useState } from "react";
import type { AskConfirmOptions } from "../types";
import type { DialogState } from "../components/common/DialogModal";

type UseDialogResult = {
  dialog: DialogState | null;
  dialogValue: string;
  setDialogValue: (value: string) => void;
  closeDialog: () => void;
  notify: (message: string) => Promise<void>;
  askConfirm: (
    message: string,
    options?: AskConfirmOptions,
  ) => Promise<boolean>;
  askPrompt: (
    message: string,
    defaultValue?: string,
    placeholder?: string,
  ) => Promise<string | null>;
};

// useDialog manages the shared alert/confirm/prompt dialog state.
export const useDialog = (): UseDialogResult => {
  const [dialog, setDialog] = useState<DialogState | null>(null);
  const [dialogValue, setDialogValue] = useState("");

  const closeDialog = useCallback(() => {
    setDialog(null);
  }, []);

  const notify = useCallback(
    (message: string) =>
      new Promise<void>((resolve) => {
        setDialog({
          type: "alert",
          message,
          resolve: () => resolve(),
        });
        setDialogValue("");
      }),
    [],
  );

  const askConfirm = useCallback(
    (message: string, options?: AskConfirmOptions) =>
      new Promise<boolean>((resolve) => {
        setDialog({
          type: "confirm",
          message,
          title: options?.title,
          confirmLabel: options?.confirmLabel,
          cancelLabel: options?.cancelLabel,
          resolve: (val: boolean) => resolve(val),
        });
        setDialogValue("");
      }),
    [],
  );

  const askPrompt = useCallback(
    (message: string, defaultValue = "", placeholder = "") =>
      new Promise<string | null>((resolve) => {
        setDialog({
          type: "prompt",
          message,
          defaultValue,
          placeholder,
          resolve: (val: string | null) => resolve(val),
        });
        setDialogValue(defaultValue);
      }),
    [],
  );

  return {
    dialog,
    dialogValue,
    setDialogValue,
    closeDialog,
    notify,
    askConfirm,
    askPrompt,
  };
};
