// isValidEmail checks a basic email pattern for client validation.
export function isValidEmail(value: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
}

// isRepRange checks for a single number or a numeric range (e.g. 8-10).
export function isRepRange(value?: string): boolean {
  if (!value) return false;
  const trimmed = value.trim();
  if (!trimmed) return false;
  return /^\d+(-\d+)?$/.test(trimmed);
}
