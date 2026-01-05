import { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import { throttle } from 'lodash-es';

// Throttle interval - updates will be batched to this frequency
const THROTTLE_MS = 250;

export function useWebSocket(url) {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState(null);
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const latestDataRef = useRef(null);
  const connectRef = useRef(null);

  // Throttled state updater - batches rapid updates
  const throttledSetMessage = useMemo(
    () =>
      throttle(
        (data) => {
          setLastMessage(data);
        },
        THROTTLE_MS,
        { leading: true, trailing: true }
      ),
    []
  );

  const connect = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}${url}`;

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setIsConnected(true);
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        // Store latest data immediately (for reference)
        latestDataRef.current = data;
        // Throttle React state updates
        throttledSetMessage(data);
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e);
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      // Reconnect after 1 second using the ref
      reconnectTimeoutRef.current = setTimeout(() => {
        connectRef.current?.();
      }, 1000);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      ws.close();
    };
  }, [url, throttledSetMessage]);

  // Keep connectRef up to date
  useEffect(() => {
    connectRef.current = connect;
  }, [connect]);

  useEffect(() => {
    connect();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      throttledSetMessage.cancel();
    };
  }, [connect, throttledSetMessage]);

  return { isConnected, lastMessage };
}
