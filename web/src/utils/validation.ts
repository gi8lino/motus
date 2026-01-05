// isValidEmail checks a basic email pattern for client validation.
export function isValidEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
}
