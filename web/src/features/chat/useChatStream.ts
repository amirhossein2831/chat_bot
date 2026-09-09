import { useCallback, useEffect, useRef, useState } from "react";

import { parseSSE } from "../../lib/sse";
import { stopGeneration } from "./chatApi";

const MAX_BACKOFF_MS = 8_000;

type StreamState = {
  buffer: string;
  streaming: boolean;
  messageId: string | null;
  truncated: boolean;
  reconnecting: boolean;
  failed: boolean;
};

const IDLE: StreamState = {
  buffer: "",
  streaming: false,
  messageId: null,
  truncated: false,
  reconnecting: false,
  failed: false,
};

export function useChatStream(onFinished: () => void, onStarted?: () => void) {
  const [state, setState] = useState<StreamState>(IDLE);

  const controller = useRef<AbortController | null>(null);
  const pending = useRef("");
  const frame = useRef(0);
  const attempt = useRef(0);
  const finished = useRef(onFinished);
  finished.current = onFinished;
  const started = useRef(onStarted);
  started.current = onStarted;

  // Setting state per token re-renders the thread dozens of times a second and
  // visibly stutters a mid-range phone. Accumulate, paint once a frame.
  const schedulePaint = useCallback(() => {
    if (frame.current) return;

    frame.current = requestAnimationFrame(() => {
      frame.current = 0;
      setState((current) => ({ ...current, buffer: pending.current }));
    });
  }, []);

  const consume = useCallback(
    async (body: ReadableStream<Uint8Array>, signal: AbortSignal) => {
      await parseSSE(
        body,
        (event, data) => {
          if (event === "meta") {
            const meta = data as { message_id: string; truncated: boolean };
            setState((current) => ({
              ...current,
              messageId: meta.message_id,
              truncated: meta.truncated,
              reconnecting: false,
            }));
            attempt.current = 0;
            // Both turns exist server-side by the time meta arrives, so the
            // thread is refetched now rather than at the end: otherwise a new
            // conversation shows its empty state for the whole generation.
            started.current?.();
            return;
          }

          if (event === "delta") {
            pending.current += (data as { text: string }).text;
            schedulePaint();
            return;
          }

          if (event === "done" || event === "error") {
            setState((current) => ({
              ...current,
              streaming: false,
              failed: event === "error",
            }));
            finished.current();
          }
        },
        signal,
      );
    },
    [schedulePaint],
  );

  // A dropped connection does not stop the generation: the server keeps going
  // and holds what it has emitted, so reconnecting replays the whole reply.
  const reattach = useCallback(
    async (messageId: string) => {
      for (;;) {
        if (controller.current?.signal.aborted) return;

        try {
          const response = await fetch(`/api/v1/chat/${messageId}/stream`, {
            credentials: "same-origin",
            signal: controller.current?.signal,
          });

          if (!response.ok || !response.body) throw new Error("reattach failed");

          pending.current = "";
          setState((current) => ({ ...current, reconnecting: false, buffer: "" }));
          await consume(response.body, controller.current!.signal);
          return;
        } catch (error) {
          if (controller.current?.signal.aborted) return;
          if (error instanceof DOMException && error.name === "AbortError") return;

          attempt.current += 1;
          setState((current) => ({ ...current, reconnecting: true }));

          const wait = Math.min(1000 * 2 ** (attempt.current - 1), MAX_BACKOFF_MS);
          await new Promise((resolve) => setTimeout(resolve, wait));
        }
      }
    },
    [consume],
  );

  const send = useCallback(
    async (conversationId: string, content: string, model?: string) => {
      controller.current?.abort();
      controller.current = new AbortController();
      pending.current = "";
      attempt.current = 0;
      setState({ ...IDLE, streaming: true });

      try {
        const response = await fetch("/api/v1/chat", {
          method: "POST",
          credentials: "same-origin",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ conversation_id: conversationId, content, model }),
          signal: controller.current.signal,
        });

        if (!response.ok || !response.body) {
          setState((current) => ({ ...current, streaming: false, failed: true }));
          return;
        }

        await consume(response.body, controller.current.signal);
      } catch (error) {
        if (error instanceof DOMException && error.name === "AbortError") return;
        setState((current) => ({ ...current, streaming: false, failed: true }));
      }
    },
    [consume],
  );

  // Aborting only detaches this client. The generation keeps running, and
  // keeps costing, until the server is told.
  const stop = useCallback(async () => {
    const { messageId } = state;
    controller.current?.abort();
    setState((current) => ({ ...current, streaming: false, reconnecting: false }));

    if (messageId) {
      await stopGeneration(messageId).catch(() => undefined);
    }
  }, [state]);

  const resume = useCallback(
    (messageId: string) => {
      controller.current?.abort();
      controller.current = new AbortController();
      pending.current = "";
      attempt.current = 0;
      setState({ ...IDLE, streaming: true, messageId });
      void reattach(messageId);
    },
    [reattach],
  );

  const reset = useCallback(() => {
    controller.current?.abort();
    pending.current = "";
    setState(IDLE);
  }, []);

  useEffect(
    () => () => {
      controller.current?.abort();
      if (frame.current) cancelAnimationFrame(frame.current);
    },
    [],
  );

  return { ...state, send, stop, resume, reset };
}
