# Security Policy

## What this library does, security-wise

`badgerpatch` deliberately makes a page of your process's own executable
memory writable, writes to it, then makes it execute-only again. That is,
by nature, the same category of operation used by legitimate JIT compilers
and by some code-injection techniques. A few concrete implications:

- **This should never run against untrusted input.** `Patch()`/`ApplyFunc()`
  should only ever be called with functions and addresses known at compile
  time in your own test code — never with an address or function value
  derived from user input, config, network data, etc. There is no sandboxing
  here; a target address you don't control is a code-execution primitive,
  not a bug in the traditional sense.
- **It will trip up hardened runtime environments on purpose.** Some
  container runtimes, macOS's hardened runtime, and some Linux LSM/seccomp
  configurations restrict W^X page transitions specifically to make this
  class of technique harder to exploit. If `badgerpatch` fails in such an
  environment, that's the security control working as intended, not a bug
  to route around.
- **Do not use this outside test binaries.** See the README's
  [Scope & Intended Use](README.md#scope--intended-use) section. A
  production service that imports this to make behavior swappable at
  runtime has, by construction, a mechanism inside it capable of rewriting
  its own code — which is a very large attack surface increase if any part
  of the address/function selection is ever influenced by anything other
  than a compile-time constant in your own source.

## Reporting a vulnerability

If you find a way to make this library do something worse than "patch the
wrong test double" — e.g. a way it could be coerced into writing to memory
outside the intended target, or a way its verification step could be
bypassed — please report it privately rather than opening a public issue:

- Preferred: open a private GitHub Security Advisory on this repo (Settings → Security → "Report a vulnerability")
- Please include: Go version, OS/arch, a minimal repro, and your assessment
  of impact.

We aim to acknowledge reports within 5 business days.

## Supported versions

Only the latest tagged release receives security fixes.