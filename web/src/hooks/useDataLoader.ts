import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type SetStateAction,
} from "react";
import { MESSAGES, toErrorMessage } from "../utils/messages";

type DataLoaderOptions = {
  enabled?: boolean;
  cacheKey?: unknown;
};

const defaultCacheKey = Symbol("data-loader");

// useDataLoader wraps async loading with loading/error state.
export function useDataLoader<T>(
  loader: () => Promise<T>,
  deps: unknown[] = [],
  options: DataLoaderOptions = {},
) {
  const enabled = options.enabled ?? true;
  const cacheKey = options.cacheKey ?? defaultCacheKey;
  const [snapshot, setSnapshot] = useState<{
    cacheKey: unknown;
    data: T | null;
  }>({ cacheKey, data: null });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const loaderRef = useRef(loader);
  const cacheKeyRef = useRef(cacheKey);
  const requestIdRef = useRef(0);
  const mountedRef = useRef(true);
  cacheKeyRef.current = cacheKey;

  const data = Object.is(snapshot.cacheKey, cacheKey) ? snapshot.data : null;

  useEffect(() => {
    loaderRef.current = loader;
  }, [loader]);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  const reload = useCallback(() => {
    if (!enabled) return;
    const requestID = ++requestIdRef.current;
    const requestCacheKey = cacheKeyRef.current;
    setLoading(true);

    loaderRef
      .current()
      .then((res) => {
        if (
          !mountedRef.current ||
          requestID !== requestIdRef.current ||
          !Object.is(requestCacheKey, cacheKeyRef.current)
        )
          return;
        setSnapshot({ cacheKey: requestCacheKey, data: res });
        setError(null);
      })
      .catch((err) => {
        if (!mountedRef.current || requestID !== requestIdRef.current) return;
        setError(toErrorMessage(err, MESSAGES.loadFailed));
      })
      .finally(() => {
        if (!mountedRef.current || requestID !== requestIdRef.current) return;
        setLoading(false);
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled, cacheKey, ...deps]);

  useEffect(() => {
    if (enabled) {
      reload();
      return;
    }
    requestIdRef.current += 1;
    setLoading(false);
  }, [enabled, reload]);

  const setData = useCallback(
    (update: SetStateAction<T | null>) => {
      setSnapshot((current) => {
        const currentData = Object.is(current.cacheKey, cacheKey)
          ? current.data
          : null;
        const nextData =
          typeof update === "function"
            ? (update as (previous: T | null) => T | null)(currentData)
            : update;
        return { cacheKey, data: nextData };
      });
    },
    [cacheKey],
  );

  return { data, loading, error, setData, reload };
}
