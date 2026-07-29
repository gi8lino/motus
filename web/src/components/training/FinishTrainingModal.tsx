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
  onSaveFeedback: (notes: string, perceivedEffort?: number) => Promise<void>;
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
  onSaveFeedback,
}: TrainingFinishModalProps) {
  const [notes, setNotes] = useState("");
  const [effort, setEffort] = useState("");
  const [saving, setSaving] = useState(false);
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
            slotProps={{ input: { readOnly: true } }}
          />
          <TextField
            label="How did it feel?"
            value={notes}
            onChange={(event) => setNotes(event.target.value)}
            multiline
            minRows={2}
          />
          <TextField
            label="Perceived effort (1–10)"
            value={effort}
            onChange={(event) => setEffort(event.target.value)}
            type="number"
            slotProps={{ htmlInput: { min: 1, max: 10 } }}
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => copySummary(summary, onCopySummary)}>
          Copy
        </Button>
        <Button
          variant="contained"
          disabled={saving}
          onClick={async () => {
            setSaving(true);
            await onSaveFeedback(
              notes,
              effort ? Number.parseInt(effort, 10) : undefined,
            );
            setSaving(false);
            onClose();
          }}
        >
          Save & close
        </Button>
      </DialogActions>
    </Dialog>
  );
}
import { useState } from "react";
