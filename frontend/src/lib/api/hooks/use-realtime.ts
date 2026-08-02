"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import { eventClient } from "@/lib/api/clients";
import { isUnauthenticated } from "@/lib/api/errors";
import type { SchemaEvent } from "@/lib/gen/event/v1/event_messages_pb";

export type EventStreamOptions = {
  projectIds?: string[];
  eventTypes?: string[];
  maxEvents?: number;
  reconnectDelayMs?: number;
};

export type EventStreamState = {
  events: SchemaEvent[];
  connected: boolean;
  lastEventId: string | null;
  error: string | null;
};

const DEFAULT_MAX_EVENTS = 200;

/**
 * Real-time event subscription backed by the backend's Redis pub/sub
 * EventService.Subscribe server stream. Auto-reconnects with backoff and
 * resumes from the last received event id.
 */
export function useEventStream(options: EventStreamOptions = {}): EventStreamState {
  const {
    projectIds = [],
    eventTypes = [],
    maxEvents = DEFAULT_MAX_EVENTS,
    reconnectDelayMs = 3000,
  } = options;

  const [events, setEvents] = useState<SchemaEvent[]>([]);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const lastEventIdRef = useRef<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);
  const retryRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const stop = useCallback(() => {
    if (retryRef.current) clearTimeout(retryRef.current);
    abortRef.current?.abort();
    setConnected(false);
  }, []);

  const connect = useCallback(() => {
    abortRef.current = new AbortController();
    const signal = abortRef.current.signal;

    const openStream = async () => {
      try {
        const iterable = eventClient.subscribe(
          {
            projectIds,
            eventTypes,
            lastEventId: lastEventIdRef.current ?? "",
          },
          { signal },
        );

        for await (const event of iterable) {
          if (signal.aborted) return;
          setError(null);
          setConnected(true);
          lastEventIdRef.current = event.id;
          setEvents((prev) => {
            const next = [event, ...prev];
            return next.slice(0, maxEvents);
          });
        }
        // stream ended without abort → schedule reconnect
        if (!signal.aborted) {
          retryRef.current = setTimeout(() => {
            setConnected(false);
            connect();
          }, reconnectDelayMs);
        }
      } catch (err) {
        if (signal.aborted) return;
        if (isUnauthenticated(err)) {
          setError("Authentication required. Please log in.");
          setConnected(false);
          return;
        }
        setConnected(false);
        retryRef.current = setTimeout(connect, reconnectDelayMs);
      }
    };

    void openStream();
  }, [projectIds, eventTypes, reconnectDelayMs, maxEvents]);

  useEffect(() => {
    setEvents([]);
    lastEventIdRef.current = null;
    connect();
    return stop;
  }, [connect, stop]);

  return {
    events,
    connected,
    lastEventId: lastEventIdRef.current,
    error,
  };
}
