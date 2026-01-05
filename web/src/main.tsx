import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import "./styles.css";

declare global {
  interface Window {
    __MOTUS_BASE?: string;
  }
}

const readRoutePrefix = (): string => {
  const meta = document.querySelector('meta[name="routePrefix"]');
  if (!meta) return "";
  const value = meta.getAttribute("content")?.trim() ?? "";
  if (!value || value === "/" || value.includes("{{")) return "";
  return value.endsWith("/") ? value.slice(0, -1) : value;
};

window.__MOTUS_BASE = readRoutePrefix();

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
