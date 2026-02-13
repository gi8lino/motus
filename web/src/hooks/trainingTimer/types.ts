import type { TrainingState } from "../../types";

// NormalizedState extends training state with local persistence metadata.
export type NormalizedState = TrainingState & { lastUpdatedAt: number };
