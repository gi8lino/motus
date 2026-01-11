const PREFIX = "[timer]";

export const logTimerEvent = (
  event: string,
  details?: Record<string, unknown>,
) => {
  console.log(PREFIX, event, details ?? {});
};
