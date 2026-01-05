// parseDurationSeconds parses a duration string into seconds.
export function parseDurationSeconds(value?: string | number | null): number {
  if (typeof value === "number") return value;
  if (!value) return 0;
  const trimmed = String(value).trim();
  if (!trimmed) return 0;
  const simple = trimmed.match(/^(\d+)(s)?$/i);
  if (simple) {
    return parseInt(simple[1], 10);
  }
  try {
    const totalMs = ms(trimmed);
    return Math.max(0, Math.round(totalMs / 1000));
  } catch {
    return 0;
  }
}

// normalizeTimestamp clears zero or invalid timestamps.
export function normalizeTimestamp(value?: string | null): string | null {
  if (!value) return null;
  if (value.startsWith("0001-01-01")) return null;
  const parsed = Date.parse(value);
  if (Number.isNaN(parsed)) return null;
  return value;
}

// Minimal duration parser supporting h/m/s/ms.
function ms(str: string): number {
  const regex = /(\d+\.?\d*)\s*(ms|s|m|h)?/gi;
  let total = 0;
  let match: RegExpExecArray | null;
  while ((match = regex.exec(str))) {
    const value = parseFloat(match[1]);
    const unit = (match[2] || "ms").toLowerCase();
    switch (unit) {
      case "h":
        total += value * 3600000;
        break;
      case "m":
        total += value * 60000;
        break;
      case "s":
        total += value * 1000;
        break;
      default:
        total += value;
    }
  }
  return total;
}
