// now returns the current timestamp in milliseconds.
export function now() {
  return Date.now();
}

// structuredCloneSafe clones state using structuredClone when available.
export function structuredCloneSafe<T>(value: T): T {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const sc = (globalThis as any).structuredClone as
    | undefined
    | ((v: any) => any);
  if (typeof sc === "function") return sc(value) as T;
  return JSON.parse(JSON.stringify(value)) as T;
}
