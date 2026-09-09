# Gateway conventions (Go)

Rationale lives in spec sections 5 and 6. These are the enforceable rules.

## Layering

    model  <- domain <- dto/mapper <- controller
      ^        ^
    service (declares ports) <- repository, client (implement implicitly)
                                        ^
                                  infrastructure

| Layer | May contain | Must never contain |
|---|---|---|
| Controller | decode DTO, call mapper, call service, map error, write | business branching, SQL, provider calls |
| Mapper | DTO <-> model/domain, pure functions | I/O |
| Service | orchestration, transactions, port calls | SQL, `net/http`, wire types |
| Repository | SQL, scanning into model | business rules, defaulting |
| Model | entity + logic about that entity | driver types, JSON tags, sentinel errors |
| Domain | non-persisted types, cross-entity logic | persistence concerns |

- `model` has a table. `domain` does not.
- `domain` may import `model`. `model` imports nothing from this codebase.
- `repository` and `client` never import `service` — Go's implicit interface
  satisfaction means they do not have to. fx binds them.
- Only `cmd/` imports everything.

## One model, two boundaries

A `model.Message` travels unchanged from repository to service to mapper.
Mapping exists in exactly two places and nowhere else:

1. DTO <-> model, in `internal/mapper`.
2. Wire format <-> `provider.ChatRequest`, inside each client package. Wire
   types are unexported and never escape their package.

## Hard rules

- **No comments unless removing one would let someone break the code.** Names
  and structure carry the meaning. Write a comment only when it is one of:
  a non-obvious constraint where the obvious-looking edit is wrong (`WriteTimeout`
  must stay 0, every config key needs a default), an external quirk that cannot
  be inferred from the code, or a command someone must run. Never restate what
  the code does, never add package or section banners, never leave a comment
  that would still be true if the code changed.
- `context.Context` is the first parameter, never a struct field.
- No globals, no `init()` side effects. Constructors take dependencies.
- `viper` is imported only by `internal/config`; everything downstream receives
  a typed `*config.Config`, validated at load.
- **Config layering**, lowest precedence first: code defaults ->
  `configs/config.yaml` -> `configs/config.<APP_ENV>.yaml` -> `APP_*` variables.
  `APP_ENV` is `local` (default), `staging` or `production`.
- **Override naming**: `APP_` + the key path in upper case with every `.`
  replaced by `_`. `http.shutdown_timeout` becomes `APP_HTTP_SHUTDOWN_TIMEOUT`.
- **Adding a field to `Config` requires adding its default** in `setDefaults`.
  Viper never consults the environment for a key it has not seen, so a field
  without a default silently ignores its `APP_*` override.
- **Secrets never live in a config file.** Per-environment files carry tuning
  only; DSNs and credentials arrive through `APP_*`. Production deliberately
  omits `postgres.dsn` so a misconfigured process fails validation instead of
  falling back to a development database.
- Errors are sentinels in `internal/constants`, wrapped with `%w`. HTTP status
  mapping lives only in `internal/transport/http/error.go`.
- `internal/constants` holds shared constants and sentinel errors. It imports
  nothing; `model` imports it, never the reverse.
- Unmapped errors return a generic message. Never leak DSNs, table names or
  upstream provider detail to a client.
- Ownership is enforced in the SQL `WHERE` clause, not in the service. A missed
  check in a service is an IDOR; a missed join returns zero rows.
- Accept interfaces, return structs.
- No interface until there are two reasons for it: a real second
  implementation, or a test that genuinely needs a double.
- Packages are named for the domain, never the layer. No `utils`.
- Every stream terminates with exactly one `DeltaDone` or one `DeltaError`.
  Billing depends on it.
- `http.Server.WriteTimeout` is `0`. Use `http.ResponseController` deadlines.
- The generation context is `context.WithoutCancel` of the request context,
  with its own timeout.

## Testing

Test flows, not functions. A test earns its place by covering a path a user or
a job actually takes: register then log in, send a message and watch it stream
and persist, reserve quota then settle it, upload then extract. Do not write a
test whose only subject is one small function's return value — it restates the
implementation, and the flow test that matters would catch the same break.

- Services against in-memory fakes, no database. No mocking library.
- Repositories against testcontainers Postgres, exercising real SQL.
- Provider clients against golden SSE fixtures, including frames split
  mid-line. That is the highest-value test in the codebase.
- Pure logic gets a test only when it encodes rules that are easy to get wrong
  and invisible when broken, such as Persian normalisation.
