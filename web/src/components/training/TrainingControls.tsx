import type { RefObject } from "react";
import PauseCircleRoundedIcon from "@mui/icons-material/PauseCircleRounded";
import PlayArrowRoundedIcon from "@mui/icons-material/PlayArrowRounded";
import SkipNextRoundedIcon from "@mui/icons-material/SkipNextRounded";
import TaskAltRoundedIcon from "@mui/icons-material/TaskAltRounded";
import { Button, Stack } from "@mui/material";
import type { TrainingState } from "../../types";
import { UI_TEXT } from "../../utils/uiText";

type TrainingControlsProps = {
  training: TrainingState | null;
  done?: boolean;
  running?: boolean;
  startLabel: string;
  isLastStep: boolean;
  mobile: boolean;
  onStart: () => void;
  onPause: () => void;
  onNext: () => void;
  onFinish: () => void;
  runButtonRef?: RefObject<HTMLButtonElement>;
  nextButtonRef?: RefObject<HTMLButtonElement>;
};

export function TrainingControls({
  training,
  done,
  running,
  startLabel,
  isLastStep,
  mobile,
  onStart,
  onPause,
  onNext,
  onFinish,
  runButtonRef,
  nextButtonRef,
}: TrainingControlsProps) {
  const handleNext = () => (isLastStep ? onFinish() : onNext());
  return (
    <Stack spacing={1.25} sx={{ width: "100%", p: mobile ? 1.25 : 0 }}>
      <Button
        ref={runButtonRef}
        variant="contained"
        size="large"
        fullWidth
        startIcon={
          running ? <PauseCircleRoundedIcon /> : <PlayArrowRoundedIcon />
        }
        onClick={running ? onPause : onStart}
        disabled={!training || done}
        sx={{
          minHeight: mobile ? 76 : 58,
          fontSize: mobile ? "1.05rem" : "1rem",
        }}
      >
        {startLabel}
      </Button>
      <Button
        ref={nextButtonRef}
        variant="contained"
        color="secondary"
        size="large"
        fullWidth
        startIcon={
          isLastStep ? <TaskAltRoundedIcon /> : <SkipNextRoundedIcon />
        }
        onClick={handleNext}
        disabled={!training || done || !training.startedAt}
        sx={{
          minHeight: mobile ? 68 : 56,
          fontSize: mobile ? "1rem" : "0.98rem",
          opacity: mobile ? 0.98 : 1,
        }}
      >
        {isLastStep
          ? UI_TEXT.training.nextLabels.finish
          : UI_TEXT.training.nextLabels.next}
      </Button>
    </Stack>
  );
}
