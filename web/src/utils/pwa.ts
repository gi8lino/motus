export function serviceWorkerUrl(basePath: string): string {
  const base = basePath.replace(/\/+$/, "");
  return `${base}/sw.js`;
}
