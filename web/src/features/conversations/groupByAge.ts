import type { Conversation } from "./conversationsApi";

type Bucket = { key: string; conversations: Conversation[] };

const DAY_MS = 86_400_000;

// A date on every row is noise when most of them say the same thing. One
// heading per span carries the same information and reads faster.
export function groupByAge(conversations: Conversation[], now = Date.now()): Bucket[] {
  const midnight = new Date(now).setHours(0, 0, 0, 0);

  const buckets: Bucket[] = [
    { key: "today", conversations: [] },
    { key: "yesterday", conversations: [] },
    { key: "week", conversations: [] },
    { key: "older", conversations: [] },
  ];

  for (const conversation of conversations) {
    const at = new Date(conversation.createdAt).getTime();
    const bucket =
      at >= midnight ? 0 : at >= midnight - DAY_MS ? 1 : at >= midnight - 6 * DAY_MS ? 2 : 3;
    buckets[bucket].conversations.push(conversation);
  }

  return buckets.filter((bucket) => bucket.conversations.length > 0);
}
