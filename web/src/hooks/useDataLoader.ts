import { useCallback, useEffect, useRef, useState } from "react";
import { MESSAGES, toErrorMessage } from "../utils/messages";

// useDataLoader wraps async loading with loading/error state.
export function useDataLoader<T>(
  loader: () => Promise<T>,
  deps: unknown[] = [],
) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const loaderRef = useRef(loader);
  const requestIdRef = useRef(0);
  const mountedRef = useRef(true);

  useEffect(() => {
    loaderRef.current = loader;
  }, [loader]);

  useEffect(() => {
    return () => {
      mountedRef.current = false;
    };
  }, []);

  const reload = useCallback(() => {
    const requestID = ++requestIdRef.current;
    setLoading(true);

    loaderRef.current()
      .then((res) => {
        if (!mountedRef.current || requestID !== requestIdRef.current) return;
        setData(res);
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
  }, deps);

  useEffect(() => {
    reload();
  }, [reload]);

  return { data, loading, error, setData, reload };
}
