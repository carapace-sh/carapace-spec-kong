# AGENTS.md

Guide for agents working in the `carapace-spec-kong` repository.

## What this is

A library that bridges [alecthomas/kong](https://github.com/alecthomas/kong) CLI apps to [carapace-spec](https://github.com/carapace-sh/carapace-spec). It walks a parsed kong CLI tree at runtime and emits a carapace-spec YAML command definition, which downstream tools consume to generate shell completions. It is **not** a standalone binary — it is a small library other kong-based CLIs import.

## Commands

Standard Go library workflow. No Makefile, no scripts.

| Task | Command |
| --- | --- |
| Build | `go build ./...` |
| Test | `go test -v -coverprofile=profile.cov ./...` |
| Format check (must be empty) | `gofmt -d -s .` |
| Static analysis | `go install honnef.co/go/tools/cmd/staticcheck@latest && staticcheck ./...` |
| Coverage report | `go tool cover -html=profile.cov` |

CI (`.github/workflows/go.yml`) runs on Go 1.24, matching the `go 1.24` directive in `go.mod` (carapace-spec v1.8.0 and its transitive dep carapace v1.13.1 both require Go 1.24). The CI formatting check asserts `gofmt -d -s .` produces no diff — run it before committing.

There are currently **no tests** in this repo (`*_test.go` files absent). CI's `go test` still runs and passes trivially. When adding tests, place them alongside `spec.go`.

## Architecture

The entire library is one file: `spec.go` (package `spec`). Control flow:

1. A consuming kong app embeds the exported `Plugin` struct as a kong subcommand tree. This exposes a hidden `_carapace spec` subcommand.
2. When the user runs `<app> _carapace spec`, kong invokes `spec.Run`, which calls `Command(ctx.Model.Node)` to convert the app's own kong node tree into a `command.Command` from `carapace-spec/pkg/command`.
3. The resulting struct is YAML-marshalled and printed to stdout. That YAML is what carapace-spec ingests to generate completion scripts.

`Command(node *kong.Node) command.Command` is the core recursive translator. It maps kong concepts onto the carapace-spec command model:

- **node → command**: `Name`, `Aliases`, `Help` → `Description`, optional `Group` key.
- **flags → flags**: `Longhand = flag.Name` (bare, no dashes), `Shorthand = string(flag.Short)` (bare, no dash) when set, `Value = !IsBool()`, `Repeatable = IsCounter() || IsCumulative()`, `Required`, `Description = flag.Help`, `Default = flag.Default`. Added via `cmd.AddFlag(f)`, which encodes them into the `Flags` map using carapace-spec's flag-string format (see below).
- **flag completions**: kong's `flag.Enum` (comma-separated) → list of choices; otherwise the kong struct tag `Type` is inspected for `path`/`existingfile` → `$files` and `existingdir` → `$directories`. **No other tag types are handled** — adding support for more is a common extension point.
- **subcommands**: `node.Children` are recursed, skipping `Hidden` ones. Hidden commands are dropped entirely (not just marked hidden).

### carapace-spec flag-string format (v1.8.0)

`AddFlag` serializes flags into the `Flags` map as compact keys (e.g. `-v, --verbose`, `--file=`, `--count*`, `--required!`, `--hidden&`). The format, defined in `carapace-spec/pkg/command/flag.go`'s `Flag.format()`, is: `[shorthand, ]longhand` followed by modifier suffixes: `?` optarg, `=` value, `*` repeatable, `!` required, `&` hidden. Keep modifiers in this exact order.

**Critical**: `format()` prepends dashes itself (`-` + Shorthand, `--` + Longhand). Set `Longhand` and `Shorthand` to **bare names** (e.g. `"verbose"`, `"v"`) — never include dashes. Including dashes produces double-dashed keys like `----verbose`. In v1.0.0 `format()` used the field values raw, so dashes were required; this changed in v1.8.0 and is an easy migration mistake.

As of carapace-spec **v1.8.0**, `Flags` and `PersistentFlags` are `FlagSet` (`map[string]Flag`), not the old `map[string]string`. The custom `MarshalYAML` emits a plain string value when a flag has only a description, but switches to an **extended object** form when `Default` or `Nargs` is set:

```yaml
# plain flag (no default/nargs)
--file=: path to file

# flag with default value -> extended notation
-l, --level=:
  description: log level
  default: info
```

The `Default` field on `command.Flag` is a string; populate it from kong's `flag.Default` (the raw string from the `default` struct tag, after variable interpolation). Since kong stores defaults as strings regardless of the underlying Go type (e.g. `default:"true"` for a bool, `default:"42"` for an int), passing `flag.Default` through directly is correct. When no `default` tag is present, both `flag.HasDefault` is false and `flag.Default` is empty, so the extended YAML form is skipped naturally.

## Gotchas

- **`Plugin` is the only exported integration point.** `spec` (lowercase) and `Command` are technically exported but `Plugin` is what consumers embed. Don't rename it without coordination with downstream users.
- **`spec.Run` returns `err` from the named return** — the `yaml.Marshal`/`Fprintln` success path relies on the named return `err` staying its zero value. Preserve this pattern if refactoring.
- **Flag completion keys use kong's `flag.Name`** (no `--` prefix), matching what carapace-spec's `FlagCompletion` expects. Don't switch to the longhand form.
- **Hidden subcommands are silently dropped** in `Command()` (`if !subcmd.Hidden`). This is intentional — kong's hidden commands shouldn't surface in completions.
- **No positional/dash/any completion is generated.** Only flag completions and subcommand structure are produced from the kong tree. kong doesn't expose positional completion metadata the way it exposes flags, so this is a known limitation, not a bug.
- The module path was migrated from `github.com/rsteube/...` to `github.com/carapace-sh/...` (see git history). If you find stale references to the old path, update them.
- **The `command.Flag` field renamed `Usage` → `Description`** in carapace-spec v1.8.0. If you encounter `Usage:` in code or diffs, update it to `Description:`. The kong side still exposes `flag.Help` (kong's own field name), which maps onto `Description`.
- **`command.Flag.Longhand`/`Shorthand` must be bare names** (no `-`/`--` prefixes) in v1.8.0. The `format()` method prepends dashes itself. This changed from v1.0.0 where the raw field value was used. See the flag-string format section above.

## Conventions

- Single-file library; no internal packages. New helpers should stay in `spec.go` unless they clearly warrant splitting.
- Follow existing style: `gofmt -s` compliant, short receiver names (`s spec`, `f flag`), table-driven where tests are added.
- The repo uses squash-merge PR flow (see git log). Dependabot keeps `alecthomas/kong`, `carapace-spec`, and GitHub Actions updated — let it handle routine bumps.
