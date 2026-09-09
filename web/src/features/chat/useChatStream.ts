import { useCallback, useEffect, useRef, useState } from "react";

import { parseSSE } from "../../lib/sse";
import { stopGeneration } from "./chatApi";

const MAX_BACKOFF_MS = 8_000;
const MAX_ATTEMPTS = 3;

type StreamState = {
  sent: string | null;
  sentAt: number | null;
  buffer: string;
  streaming: boolean;
  messageId: string | null;
  truncated: boolean;
  reconnecting: boolean;
  failed: boolean;
  failureCode: string | null;
};

const IDLE: StreamState = {
  sent: null,
  sentAt: null,
  buffer: "",
  streaming: false,
  messageId: null,
  truncated: false,
  reconnecting: false,
  failed: false,
  failureCode: null,
};

export function useChatStream(onFinished: () => void, onStarted?: () => void) {
  const [state, setState] = useState<StreamState>(IDLE);

  const controller = useRef<AbortController | null>(null);
  const abandoned = useRef(new Set<string>());
  const last = useRef<(() => void) | null>(null);
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
      let terminated = false;

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
            terminated = true;
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

      // A body that ends without a terminal event leaves the view streaming
      // forever, so the end of the body is treated as one.
      if (!terminated && !signal.aborted) {
        setState((current) => ({ ...current, streaming: false }));
        finished.current();
      }
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

          // The server holds a generation only while it runs. A 4xx says this
          // one is not running, which usually means it finished while we were
          // away: the reply is already stored, so refetch rather than retry.
          if (response.status >= 400 && response.status < 500) {
            abandoned.current.add(messageId);
            setState((current) => ({ ...current, streaming: false, reconnecting: false }));
            finished.current();
            return;
          }

          if (!response.ok || !response.body) throw new Error("reattach failed");

          pending.current = "";
          setState((current) => ({ ...current, reconnecting: false, buffer: "" }));
          await consume(response.body, controller.current!.signal);
          return;
        } catch (error) {
          if (controller.current?.signal.aborted) return;
          if (error instanceof DOMException && error.name === "AbortError") return;

          attempt.current += 1;
          if (attempt.current > MAX_ATTEMPTS) {
            abandoned.current.add(messageId);
            setState((current) => ({
              ...current,
              streaming: false,
              reconnecting: false,
              failed: true,
              failureCode: "disconnected",
            }));
            finished.current();
            return;
          }

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
      last.current = () => void send(conversationId, content, model);
      controller.current?.abort();
      controller.current = new AbortController();
      pending.current = "";
      attempt.current = 0;
      // The turn is shown from here, not from the server's copy of it: waiting
      // for that leaves the message invisible for as long as the model takes
      // to produce its first token.
      setState({ ...IDLE, sent: content, sentAt: Date.now(), streaming: true });

      try {
        const response = await fetch("/api/v1/chat", {
          method: "POST",
          credentials: "same-origin",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ conversation_id: conversationId, content, model }),
          signal: controller.current.signal,
        });

        if (!response.ok || !response.body) {
          const payload = await response.json().catch(() => ({ code: "unknown" }));
          setState((current) => ({
            ...current,
            streaming: false,
            failed: true,
            failureCode: String(payload.code ?? "unknown"),
          }));
          // The turn was persisted before dispatch failed, so the thread is
          // refetched: otherwise the message the user just sent disappears.
          finished.current();
          return;
        }

        await consume(response.body, controller.current.signal);
      } catch (error) {
        if (error instanceof DOMException && error.name === "AbortError") return;
        setState((current) => ({
          ...current,
          streaming: false,
          failed: true,
          failureCode: "disconnected",
        }));
        finished.current();
      }
    },
    [consume],
  );

  const regenerate = useCallback(
    async (messageId: string, model?: string) => {
      last.current = () => void regenerate(messageId, model);
      controller.current?.abort();
      controller.current = new AbortController();
      pending.current = "";
      attempt.current = 0;
      setState({ ...IDLE, sentAt: Date.now(), streaming: true });

      try {
        const response = await fetch(`/api/v1/chat/${messageId}/regenerate`, {
          method: "POST",
          credentials: "same-origin",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ model }),
          signal: controller.current.signal,
        });

        if (!response.ok || !response.body) {
          const payload = await response.json().catch(() => ({ code: "unknown" }));
          setState((current) => ({
            ...current,
            streaming: false,
            failed: true,
            failureCode: String(payload.code ?? "unknown"),
          }));
          finished.current();
          return;
        }

        await consume(response.body, controller.current.signal);
      } catch (error) {
        if (error instanceof DOMException && error.name === "AbortError") return;
        setState((current) => ({
          ...current,
          streaming: false,
          failed: true,
          failureCode: "disconnected",
        }));
        finished.current();
      }
    },
    [consume],
  );

  const retry = useCallback(() => {
    last.current?.();
  }, []);

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

  // Without this the caller re-resumes the same dead generation on every
  // refetch, which is an endless "reconnecting" the user cannot escape.
  const resume = useCallback(
    (messageId: string) => {
      if (abandoned.current.has(messageId)) return;

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

  return { ...state, send, stop, resume, reset, regenerate, retry };
}
