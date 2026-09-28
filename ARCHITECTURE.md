# Architecture

Three conventions apply across this codebase, from the first commit onward — not just in the
packages where they happen to be visible today.

## `internal/`, not `pkg/`

`pkg/` is an unofficial convention that signals "importable by anyone." Nothing in this module is
meant to be imported by another module. `internal/` is enforced by the Go compiler and correctly
says "application-private."

## Dependency direction

`internal/app` is the composition root: it constructs dependencies and wires them together
explicitly. Only `app` imports `internal/{api,db,clients,tasks}` — never the reverse. Keeping this
one-directional avoids import cycles once handlers need client/config access.

## Dependency injection via constructors, no singletons

Dependencies (the DB `Store`, and later the Kratos/Keto clients) are constructed once in `app` and
passed down explicitly through constructors — not fetched via package-level `GetXClient()`-style
singleton getters. This is what makes each piece testable as a plain Go value instead of requiring
global state to be mocked.
