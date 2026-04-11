import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
  Typography,
} from "@mui/material";

type TrainingFinishModalProps = {
  summary: string | null;
  onClose: () => void;
  onCopySummary: () => void;
};

const copySummary = (summary: string, onCopySummary: () => void) => {
  if (navigator?.clipboard?.writeText) {
    navigator.clipboard.writeText(summary).catch(() => {});
  }
  onCopySummary();
};

export function TrainingFinishModal({
  summary,
  onClose,
  onCopySummary,
}: TrainingFinishModalProps) {
  if (!summary) return null;

  return (
    <Dialog open onClose={onClose} fullWidth maxWidth="md">
      <DialogTitle>Great job!</DialogTitle>
      <DialogContent dividers>
        <Stack spacing={2}>
          <Typography color="text.secondary">
            Training finished. Copy the summary for AI.
          </Typography>

          <TextField
            value={summary}
            multiline
            minRows={10}
            fullWidth
            InputProps={{ readOnly: true }}
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => copySummary(summary, onCopySummary)}>Copy</Button>
        <Button variant="contained" onClick={onClose}>
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );
}
