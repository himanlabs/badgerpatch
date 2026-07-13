# badgerpatch

[![CI](https://github.com/himanlabs/badgerpatch/actions/workflows/ci.yml/badge.svg)](https://github.com/himanlabs/badgerpatch/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/himanlabs/badgerpatch.svg)](https://pkg.go.dev/github.com/himanlabs/badgerpatch)
[![Go Report Card](https://goreportcard.com/badge/github.com/himanlabs/badgerpatch)](https://goreportcard.com/report/github.com/himanlabs/badgerpatch)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`badgerpatch` is a runtime monkey-patching library for Go. It replaces a function's
compiled machine code with a jump to a different function of the same signature,
so calls to it are redirected at the CPU level — no interfaces, no dependency
injection scaffolding required.

```go
patches := badgerpatch.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
    return []byte("fixture data"), nil
})
defer patches.Reset()
```

**This is a test-doubles library, not a production runtime mechanism.** See
[Scope & Intended Use](#scope--intended-use) below — please read it before
adopting this, it will save you a bad on-call night.

## Why this exists

Go's type system and lack of monkeypatching-by-default push you toward
interfaces + DI for testability, which is usually right. But sometimes you're
testing code you don't own, wrapping a third-party SDK, or dealing with a
free function (`time.Now`, `os.ReadFile`, a package-level helper) where adding
an interface is disproportionate. `badgerpatch` lets you stub that one call
for the duration of a test, without restructuring the code under test.

## Install

```sh
go get github.com/himanlabs/badgerpatch
```

Requires Go 1.21+. Supported platforms:

| OS      | Arch    | Mechanism                                   | Status |
|---------|---------|----------------------------------------------|--------|
| Linux   | amd64   | `mprotect` + direct machine-code jump        | ✅ tested in CI |
| Linux   | arm64   | `mprotect` + direct machine-code jump        | ✅ tested in CI (QEMU) |
| Windows | amd64   | `VirtualProtect` + direct machine-code jump  | ✅ tested in CI |
| macOS   | arm64   | `pthread_jit_write_protect_np` (via cgo)     | ⚠️ works, requires cgo, not covered by CI |
| Other   | —       | —                                             | ❌ not supported |

## Usage

### Patch a single function

```go
func TestFetchesUser(t *testing.T) {
    patch := badgerpatch.StaticReturn(fetchUserFromAPI, func(id string) (*User, error) {
        return &User{ID: id, Name: "stub"}, nil
    })
    defer patch.Unpatch()

    u, err := handler.GetUser("42")
    // ...
}
```

### Patch several functions as a group

```go
func TestOrderFlow(t *testing.T) {
    patches := badgerpatch.ApplyFunc(chargeCard, mockCharge)
    badgerpatch.Apply(patches, sendReceipt, mockReceipt)
    defer patches.Reset() // unpatches everything, in one call

    // ...
}
```

### Return a different value on each call

Useful for testing retry logic — first two calls fail, third succeeds:

```go
func TestRetriesThenSucceeds(t *testing.T) {
    seq := badgerpatch.NewSequence[error](
        errors.New("timeout"),
        errors.New("connection reset"),
        nil,
    )
    patches := badgerpatch.ApplyFunc(FetchData, func() (string, error) {
        if err := seq.Next(); err != nil {
            return "", err
        }
        return "ok", nil
    })
    defer patches.Reset()

    // ...
}
```

## Required: mark target functions `//go:noinline`

This is the single most common way to misuse this library, so it gets its own
heading instead of a bullet point.

The Go compiler inlines small functions at their call sites. If it inlines
your target function, the call never touches the function's compiled machine
code — it *is* the machine code at that point — so patching it does nothing,
silently, at that call site. `badgerpatch` cannot detect this from inside
`Patch()`: it only has visibility into the target function's own bytes, not
every place that calls it.

Fix: add `//go:noinline` directly above any function you intend to patch.

```go
//go:noinline
func fetchUserFromAPI(id string) (*User, error) {
    // ...
}
```

If you don't control the function's source (e.g. it's in a third-party
package), you can instead build your test binary with inlining disabled:

```sh
go test -gcflags="all=-l" ./...
```

...but this disables inlining for the *entire* build, which is a blunt
instrument and will make your test suite slower. Prefer `//go:noinline` on
the specific target wherever you can.

`Patch()` does verify that its own write succeeded (see
[How it works](#how-it-works)), so OS-level failures fail loudly. It cannot
detect the inlining case, because there's nothing wrong with the write itself
— the problem is a call site elsewhere that never reads it.

## Scope & intended use

Use `badgerpatch` from `_test.go` files, to replace collaborators for the
duration of a test. Do not import it from non-test code to make runtime
behavior swappable in a live service. Reasons:

- **Not atomic.** A patch is a multi-instruction write to executable memory.
  If another goroutine calls the target function while the write is
  in-flight, it can execute a torn instruction sequence. In a test, you
  control this by patching before spawning any concurrent work. In a live
  service handling real traffic, you generally can't guarantee that.
- **Environment-dependent.** The technique relies on being able to flip a
  code page between writable and executable (`mprotect`/`VirtualProtect`, or
  the macOS JIT toggle). Some hardened kernels, gVisor/container sandboxes,
  or security-conscious deployment targets restrict exactly this kind of
  W^X transition, sometimes exactly *because* it's associated with tricks
  like this one. Code that depends on it can work in your dev/CI environment
  and fail in production.
- **Silent failure mode.** As documented above, an inlined target silently
  doesn't patch. In a test, a bad assertion catches this immediately. In
  production, "the mock override didn't take" isn't a failure mode you get
  a clean error for — you get inconsistent behavior that looks like a bug
  somewhere else entirely.

If what you actually need is swappable behavior in production — feature
flags, environment-specific implementations, gradual rollouts — reach for an
interface + dependency injection, a strategy pattern, or a feature-flag
service instead. Those are boring, well-understood tools, precisely because
they don't have any of the above failure modes.

## How it works

For a target `T` and replacement of the same type:

1. Look up the target's function pointer via `reflect.ValueOf(target).Pointer()`.
2. Save a copy of the first N bytes at that address (N = the length of a
   jump instruction sequence for the current architecture), so patching can
   be reversed.
3. Make that memory page writable (`mprotect`/`VirtualProtect`/JIT toggle).
4. Overwrite those bytes with a direct jump to the replacement function's
   address.
5. Restore the page to execute-only.
6. Read the bytes back and compare — if they don't match what was written,
   panic immediately rather than return a guard that silently isn't patching
   anything.

`Unpatch()` reverses step 4 using the bytes saved in step 2.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Architecture-specific changes
(`jmp_*.go`, `copy_*.go`) require the accompanying disassembly/verification
described there — these are the files most likely to introduce a silent,
hard-to-diagnose bug, so PRs touching them get extra scrutiny.

## Security

This library intentionally makes writable-then-executable memory pages. See
[SECURITY.md](SECURITY.md) for how to report concerns.

## License

MIT — see [LICENSE](LICENSE).