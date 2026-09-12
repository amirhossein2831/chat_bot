import { useCallback, useEffect, useRef, useState } from "react";

import { isStorageUnavailable } from "../../lib/errors";
import {
  confirmUpload,
  fetchAttachment,
  presign,
  putToStorage,
  type Attachment,
} from "./filesApi";

const POLL_INTERVAL_MS = 2_000;

export type UploadItem = Attachment & { uploading?: boolean };

export function useUpload() {
  const [items, setItems] = useState<UploadItem[]>([]);
  const [storageDown, setStorageDown] = useState(false);
  const timers = useRef<Map<string, number>>(new Map());

  const patch = useCallback((id: string, changes: Partial<UploadItem>) => {
    setItems((current) =>
      current.map((item) => (item.id === id ? { ...item, ...changes } : item)),
    );
  }, []);

  // Extraction is a job, so the result arrives later. Polling stops as soon as
  // the status leaves the queue, whatever it lands on.
  const poll = useCallback(
    (id: string) => {
      const timer = window.setInterval(async () => {
        try {
          const attachment = await fetchAttachment(id);
          if (attachment.status !== "queued" && attachment.status !== "pending") {
            window.clearInterval(timer);
            timers.current.delete(id);
            patch(id, { ...attachment, uploading: false });
          }
        } catch {
          window.clearInterval(timer);
          timers.current.delete(id);
        }
      }, POLL_INTERVAL_MS);

      timers.current.set(id, timer);
    },
    [patch],
  );

  const upload = useCallback(
    async (file: File) => {
      try {
        const { id, url } = await presign(file);

        setItems((current) => [
          ...current,
          {
            id,
            filename: file.name,
            mime: file.type,
            size: file.size,
            status: "pending",
            error: null,
            uploading: true,
          },
        ]);

        await putToStorage(url, file);
        const confirmed = await confirmUpload(id);
        patch(id, { ...confirmed, uploading: false });
        poll(id);
      } catch (error) {
        // Storage being down disables uploads and nothing else: the rest of
        // chat keeps working, which is the point of the gateway degrading.
        if (isStorageUnavailable(error)) setStorageDown(true);
      }
    },
    [patch, poll],
  );

  const remove = useCallback((id: string) => {
    const timer = timers.current.get(id);
    if (timer) {
      window.clearInterval(timer);
      timers.current.delete(id);
    }
    setItems((current) => current.filter((item) => item.id !== id));
  }, []);

  const clear = useCallback(() => {
    timers.current.forEach((timer) => window.clearInterval(timer));
    timers.current.clear();
    setItems([]);
  }, []);

  useEffect(() => {
    const active = timers.current;
    return () => active.forEach((timer) => window.clearInterval(timer));
  }, []);

  return { items, upload, remove, clear, storageDown };
}
