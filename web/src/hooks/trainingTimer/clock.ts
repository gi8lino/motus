// now returns the current timestamp in milliseconds.
export function now(): number {
  return Date.now();
}

type CloneCapableGlobal = typeof globalThis & {
  structuredClone?: <T>(value: T) => T;
};

// structuredCloneSafe clones data with structuredClone when available.
export function structuredCloneSafe<T>(value: T): T {
  const clone = (globalThis as CloneCapableGlobal).structuredClone;
  if (typeof clone === "function") return clone(value);
  return JSON.parse(JSON.stringify(value)) as T;
}
