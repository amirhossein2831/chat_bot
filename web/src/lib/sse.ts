export type FrameHandler = (event: string, data: unknown) => void;

// Frames arrive split across TCP reads, so the buffer is what makes this
// correct. Kept free of React and fetch so it can be reasoned about alone.
export async function parseSSE(
  stream: ReadableStream<Uint8Array>,
  onFrame: FrameHandler,
  signal?: AbortSignal,
): Promise<void> {
  const reader = stream.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  try {
    for (;;) {
      if (signal?.aborted) return;

      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });

      let boundary = buffer.indexOf("\n\n");
      while (boundary !== -1) {
        emit(buffer.slice(0, boundary), onFrame);
        buffer = buffer.slice(boundary + 2);
        boundary = buffer.indexOf("\n\n");
      }
    }
  } finally {
    reader.releaseLock();
  }
}

function emit(frame: string, onFrame: FrameHandler): void {
  let event = "message";
  let data = "";

  for (const line of frame.split("\n")) {
    if (line.startsWith("event:")) event = line.slice(6).trim();
    else if (line.startsWith("data:")) data += line.slice(5).trim();
  }

  if (!data) return;

  try {
    onFrame(event, JSON.parse(data));
  } catch {
    return;
  }
}
