# Contributing to BadgerPatch

First off, thank you for considering contributing to **BadgerPatch**! 🎉

BadgerPatch is a low-level monkey patching library that modifies executable memory at runtime. Because it interacts directly with architecture-specific machine code and operating system memory protection APIs, even small changes can have significant effects.

Please read these guidelines before submitting a contribution.

---

# Development Setup

## Requirements

Before getting started, ensure you have:

- Go **1.26.4** or later
- Git
- A supported operating system:
  - Linux
  - Windows
  - macOS (Intel or Apple Silicon)

---

## Clone the Repository

```bash
git clone https://github.com/himanlabs/badgerpatch.git
cd badgerpatch
```

---

## Build

```bash
go build ./...
```

---

## Run Tests

Go aggressively optimizes and inlines functions during compilation.

Because **BadgerPatch** patches compiled machine code directly, all tests **must** be executed with optimizations disabled.

```bash
go test -v -gcflags="all=-N -l" ./...
```

---

# Pull Request Process

Before opening a Pull Request, please ensure the following:

- Fork the repository and create your branch from `main`.
- Add tests for all new features and bug fixes.
- Ensure existing tests continue to pass.
- Format your code using:

```bash
go fmt ./...
```

- Update documentation (`README.md`, examples, GoDoc comments) whenever the public API changes.
- Keep pull requests focused on a single feature or fix whenever possible.

---

# Cross-Platform Compatibility

BadgerPatch supports multiple operating systems and CPU architectures.

If your changes affect:

- executable memory management
- machine code generation
- architecture-specific instructions
- page protection
- runtime patching

please consider compatibility with:

| Operating System | Architectures |
|------------------|--------------|
| Linux | amd64, arm64 |
| Windows | amd64 |
| macOS | amd64, arm64 (Apple Silicon) |

If you cannot verify your changes on every supported platform, mention this clearly in your Pull Request so maintainers can perform additional testing.

---

# Development Guidelines

## Type Safety

BadgerPatch uses **Go Generics** to provide compile-time type safety.

Avoid introducing runtime reflection into the critical patching path unless it is strictly necessary and discarded immediately after extracting the required function pointer.

---

## Memory Handling

For reading and writing executable memory:

- ✅ Use `unsafe.Slice`
- ❌ Do **not** use `reflect.SliceHeader`

`unsafe.Slice` is the modern, supported approach and avoids deprecated runtime behavior.

---

## Performance

Performance is a core design goal.

When contributing:

- Avoid unnecessary allocations.
- Avoid unnecessary reflection.
- Minimize synchronization overhead.
- Keep the patching path as lightweight as possible.

If your change impacts performance, consider including benchmark results in your Pull Request.

---

## Architecture-Specific Code

Files responsible for machine code generation and executable memory handling require extra care.

Examples include:

- `jmp_amd64.go`
- `jmp_arm64.go`
- `copy_windows.go`
- `copy_linux.go`
- `copy_darwin_arm64.go`

When modifying these files:

- Verify generated instruction bytes.
- Test on the target architecture whenever possible.
- Add regression tests for bugs.
- Explain any architecture-specific behavior in the PR description.

---

# Reporting Bugs

When reporting a bug, please include:

- Go version (`go version`)
- Operating system
- CPU architecture
- Minimal reproducible example
- Expected behavior
- Actual behavior
- Stack trace or panic output (if applicable)

---

# Code Style

Please follow standard Go conventions:

- Run `go fmt ./...`
- Keep exported APIs documented.
- Prefer simple, readable implementations.
- Maintain backward compatibility whenever possible.

---

# Questions & Discussions

If you're unsure whether a feature belongs in BadgerPatch or have questions about the implementation, feel free to open a GitHub Discussion or Issue before starting work.

Early discussion often saves time for both contributors and maintainers.

---

# Thank You ❤️

Every contribution—whether it's code, documentation, bug reports, or suggestions—helps make **BadgerPatch** better.

We appreciate your time and effort, and we look forward to your contributions!