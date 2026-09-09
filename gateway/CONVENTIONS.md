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
- Enums live in `model/enums.go`, each with its own named type and a
  `Valid()` method. They are domain vocabulary, not constants: a
  `constants.Role` would split the entity definition across packages and
  invert the dependency. Name the type for what it describes
  (`MessageStatus`, `UserStatus`) so two unrelated status sets cannot be
  assigned to each other.
- **Enum file layout**: all type declarations first, one per line; then one
  `const` block per enum; then the methods. Not type-and-consts interleaved
  per enum.
- When a comment is warranted at all, begin it with the identifier it
  documents: `// MarkComplete settles quota, so a second call would ...`.
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
- **Mapper direction is decided by the service, not by the mapper.** The mapper
  converts to whatever the service's signature takes and from whatever it
  returns, so there is never a choice between model and domain to agonise over.
  A service with several non-entity inputs declares its own command struct in
  its own package.
- **`domain` <-> `model` conversion is not mapping, it is construction**, and it
  lives in `domain` as a constructor (`domain.ThreadFrom`). Putting it in the
  transport mapper would make repositories import transport.
- Name the target in the mapper function: `ToUserModel`, `ToSendCommand`,
  `ToUserResponse`. Mappers move fields and nothing else; validation or
  computation belongs in `model` or `domain`.
- **DRY.** Every rule has one home. Error-to-status mapping lives in one file,
  Persian normalisation in one function, a config key in one struct field, a
  wire format in one client package. If you are about to copy a second copy of
  a rule, move it instead. The exception worth naming: duplication is cheaper
  than the wrong abstraction, so wait for the third occurrence before
  extracting something whose shape you are still guessing at.
- **SOLID, as it applies in Go.**
  - *Single responsibility*: a package changes for one reason. Controllers
    parse, services orchestrate, repositories persist.
  - *Open/closed*: a new provider is a new package satisfying `Provider`, never
    another arm on a switch in `chat`.
  - *Liskov*: every implementation honours the contract exactly. A stream ends
    with exactly one terminal delta regardless of which client produced it.
  - *Interface segregation*: interfaces are one to three methods and named for
    what the caller wants, not what the implementation is.
  - *Dependency inversion*: services depend on interfaces they declare;
    concrete types are chosen only at the composition root.
- Accept interfaces, return structs.
- **One receiver kind per type.** If any method mutates, every method on
  that type takes a pointer. Mixing them keeps the value out of its own
  method set, so `model.Message` would silently fail to satisfy an
  interface that `*model.Message` satisfies. Types with no mutating
  methods use value receivers throughout.
- No interface until there are two reasons for it: a real second
  implementation, or a test that genuinely needs a double.
- Packages are named for the domain, never the layer. No `utils`.
- Every stream terminates with exactly one `DeltaDone` or one `DeltaError`.
  Billing depends on it.
- `http.Server.WriteTimeout` is `0`. Use `http.ResponseController` deadlines.
- The generation context is `context.WithoutCancel` of the request context,
  with its own timeout.

## Testing

**Deferred.** No test files while the package layout is still changing. The
testcontainers harness in `internal/util/testutil` stays for when they return.

When they do: test flows, not functions. A test earns its place by covering a
path a user or a job actually takes — register then log in, send a message and
watch it stream and persist, reserve quota then settle it. Services against
in-memory fakes, repositories against testcontainers Postgres, provider clients
against golden SSE fixtures including frames split mid-line.
