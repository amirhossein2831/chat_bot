import { api } from "../../lib/apiClient";

type PresignWire = {
  attachment_id: string;
  url: string;
  expires_at: string;
};

type AttachmentWire = {
  id: string;
  filename: string;
  mime: string;
  size: number;
  status: string;
  error?: string | null;
};

export type AttachmentStatus =
  | "pending"
  | "queued"
  | "done"
  | "failed"
  | "unsupported";

export type Attachment = {
  id: string;
  filename: string;
  mime: string;
  size: number;
  status: AttachmentStatus;
  error: string | null;
};

const toAttachment = (wire: AttachmentWire): Attachment => ({
  id: wire.id,
  filename: wire.filename,
  mime: wire.mime,
  size: wire.size,
  status: wire.status as AttachmentStatus,
  error: wire.error ?? null,
});

export async function presign(file: File): Promise<{ id: string; url: string }> {
  const wire = await api.post<PresignWire>("/files/presign", {
    filename: file.name,
    mime: file.type || "application/octet-stream",
    size: file.size,
  });
  return { id: wire.attachment_id, url: wire.url };
}

// The browser PUTs straight to storage, so file bytes never occupy a gateway
// connection and a slow upload cannot tie one up.
export async function putToStorage(url: string, file: File): Promise<void> {
  const response = await fetch(url, {
    method: "PUT",
    body: file,
    headers: { "Content-Type": file.type || "application/octet-stream" },
  });
  if (!response.ok) throw new Error(`upload failed: ${response.status}`);
}

export async function confirmUpload(id: string): Promise<Attachment> {
  return toAttachment(await api.post<AttachmentWire>(`/files/${id}/confirm`));
}

export async function fetchAttachment(id: string): Promise<Attachment> {
  return toAttachment(await api.get<AttachmentWire>(`/files/${id}`));
}
