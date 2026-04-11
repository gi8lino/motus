import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Typography,
} from "@mui/material";
import { formatCountdownMillis } from "../../utils/format";
import { UI_TEXT } from "../../utils/uiText";

type TrainingOverrunModalProps = {
  show: boolean;
  countdown: number;
  onPause: () => void;
  onPostpone: () => void;
  showHours?: boolean;
};

export function TrainingOverrunModal({
  show,
  countdown,
  onPause,
  onPostpone,
  showHours,
}: TrainingOverrunModalProps) {
  if (!show) return null;

  return (
    <Dialog open onClose={onPause} maxWidth="xs" fullWidth>
      <DialogTitle>{UI_TEXT.pages.training.overrunTitle}</DialogTitle>
      <DialogContent dividers>
        <Typography color="text.secondary">
          {UI_TEXT.pages.training.overrunMessage}{" "}
          {formatCountdownMillis(countdown, { showHours })}.
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={onPostpone}>{UI_TEXT.pages.training.overrunPostpone}</Button>
        <Button variant="contained" onClick={onPause}>
          {UI_TEXT.actions.pause}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
