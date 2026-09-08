# Web conventions (React)

Rationale lives in spec section 11. Normal React practice — deliberately not a
port of the backend layering.

## Stack

React 19 + Vite + TypeScript strict · Tailwind v4 · shadcn/ui (Radix) ·
TanStack Query v5 · Zustand · React Router v7 · i18next · native `Intl`

Material UI was considered and rejected: Material Design's look is wrong for a
chat product, its bundle cost is real on Iranian mobile networks, and its RTL
support is a transform plugin rather than native.

## Structure: colocate by feature

Everything about a feature lives in one folder — components, hooks, API calls,
types, tests — because those files change together.

    src/features/chat/
      ChatPage.tsx  MessageList.tsx  Composer.tsx  BranchSwitcher.tsx
      useChatStream.ts  useConversation.ts  chatApi.ts  types.ts
      Composer.test.tsx

## Rules

- **Custom hooks are the unit of abstraction.** Not services, not use-cases, no
  classes.
- **TanStack Query owns server state.** No manual caching, no mirroring server
  data into a store. Query keys namespaced per feature.
- **Zustand holds only client state**: live token buffer, composer draft,
  theme, sidebar.
- One typed fetch wrapper in `lib/apiClient.ts`; feature API functions are thin
  and live with their feature.
- API shapes are converted where they arrive, in `*Api.ts` — not in components.
- **No barrel files.** They hurt Vite dev performance and tree-shaking.
- Colocate tests.
- **Extract on the second use, not the first.** The React failure mode is
  premature abstraction, not missing layers.
- Props down, events up. Composition over configuration props.

## RTL — non-negotiable

- **Logical properties only.** `ms-*`, `me-*`, `ps-*`, `pe-*`, `start`, `end`.
  An ESLint rule bans `ml-*`, `mr-*`, `pl-*`, `pr-*`, `left-*`, `right-*`,
  `text-left`, `text-right` from day one.
- `dir="ltr"` on `<pre>` and inline code.
- Inline Latin runs wrapped in `<bdi>`.
- Jalali dates and Persian digits via `Intl`, display-only. Never stored.
- Vazirmatn self-hosted and subset.

## Streaming

- `fetch` + `ReadableStream`, never `EventSource` — it is GET-only and cannot
  send headers or a body.
- **Batch renders on `requestAnimationFrame`.** Per-token `setState` visibly
  janks on mid-range Android.
- Defer syntax highlighting until the stream completes.
- Close unbalanced markdown fences before rendering partial output.
- On `done`, invalidate the query. The local buffer is never the source of
  truth.

## Kept out of React on purpose

Pure logic that is painful to test through a component:

- `lib/sse.ts` — frame parser, tested against recorded fixtures with
  mid-line splits.
- `lib/textnorm.ts` — port of the Go normaliser. If they diverge, client
  filtering contradicts server search.
