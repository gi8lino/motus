const PREFIX = "[timer]";

// logTimerEvent emits timer debug information to the console.
export const logTimerEvent = (
  event: string,
  details?: Record<string, unknown>,
) => {
  console.log(PREFIX, event, details ?? {});
};
