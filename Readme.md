#  BadgerPatch

> A modern, high-performance, and strictly type-safe monkey patching library for Go.

`badgerpatch` is a modern monkey patching library designed specifically for **Go unit tests**. It leverages **Go Generics**, modern runtime techniques, and architecture-aware machine code generation to provide a safer and more reliable alternative to traditional monkey patching libraries.

It focuses on:

- Compile-time type safety
- High performance
- Modern Go ABI compatibility
- Cross-platform support
- Minimal runtime overhead

---

##  Features

- **Compile-Time Type Safety**
    - Mismatched mock signatures fail immediately during compilation.
    - Eliminates confusing runtime reflection panics.

- **Zero-Allocation Engine**
    - Uses modern `unsafe.Slice` instead of deprecated `reflect.SliceHeader`.

- **Modern ABI Compliance**
    - Uses architecture-safe scratch registers (such as **R11** on AMD64) to avoid register corruption.

- **Apple Silicon Support**
    - Native macOS ARM64 support using `pthread_jit_write_protect_np` through CGO.

- **Stateful Mock Sequences**
    - Return different values across consecutive function calls.

- **Cross Platform**
  - Linux
  - Windows
  - macOS
  - Apple Silicon

---

# ⚠️ Important

## Disable Function Inlining

Go aggressively inlines functions during compilation.

If a function gets inlined, there is **no function call left to intercept**, so monkey patching cannot work.

Always run your test suite with optimizations disabled:

```bash
go test -gcflags="all=-N -l" ./...
```

or

```bash
go test -v -gcflags="all=-N -l" ./...
```

---

# 📦 Installation

```bash
go get github.com/himanlabs/badgerpatch
```

---

# 🚀 Quick Start

## Basic Function Patching

Use `ApplyFunc` to patch a function and automatically restore it after the test.

```go
package main

import (
    "testing"

    "github.com/himanlabs/badgerpatch"
)

func FetchData() string {
    return "Real Database Hit"
}

func TestFetchData(t *testing.T) {
    patches := badgerpatch.ApplyFunc(FetchData, func() string {
        return "Mocked Data"
    })

    defer patches.Restore()

    result := FetchData()

    if result != "Mocked Data" {
        t.Fatalf("Expected 'Mocked Data', got %s", result)
    }
}
```

---

# Returning Static Values

If you only want a function to return a fixed value, use `Return`.

```go
func TestWithReturn(t *testing.T) {

    patches := badgerpatch.NewSet()
    defer patches.Restore()

    badgerpatch.Return(
        patches,
        ValidateUser,
        func(int) bool {
            return true
        },
    )
}
```

---

# Stateful Sequences

Create mocks that return different values on successive calls.

```go
func TestWithSequence(t *testing.T) {

    seq := badgerpatch.NewSequence(
        errors.New("timeout"),
        errors.New("connection reset"),
        nil,
    )

    patches := badgerpatch.ApplyFunc(FetchData, func() (string, error) {

        err := seq.Next()

        if err != nil {
            return "", err
        }

        return "Success", nil
    })

    defer patches.Restore()

    _, err1 := FetchData()
    _, err2 := FetchData()
    res, err3 := FetchData()

    _ = err1
    _ = err2
    _ = err3
    _ = res
}
```

---

# 🚫 The Closure Trap

## Why it happens

Do **not** capture local variables inside your mock function.

When a mock captures local state, the Go compiler transforms it into a closure that requires a hidden context pointer.

Because `badgerpatch` performs a direct jump into machine code, that context register does not exist, causing a segmentation fault.

---

## ❌ Bad (Will Panic)

```go
func TestBad(t *testing.T) {

    counter := 0

    patches := badgerpatch.ApplyFunc(DoWork, func() {
        counter++
    })

    defer patches.Restore()
}
```

---

## ✅ Good (Safe)

```go
var testCounter int

func TestGood(t *testing.T) {

    testCounter = 0

    patches := badgerpatch.ApplyFunc(DoWork, func() {
        testCounter++
    })

    defer patches.Restore()
}
```

---

# 📊 Comparison

| Feature | BadgerPatch | gomonkey |
|----------|-------------|-----------|
| Type Safety | ✅ Compile-Time (Generics) | ❌ Runtime Reflection |
| Memory API | ✅ `unsafe.Slice` | ❌ `reflect.SliceHeader` |
| Apple Silicon | ✅ Native Support | ⚠ Requires Workarounds |
| AMD64 ABI | ✅ Safe Scratch Registers | ❌ Register Conflicts |
| Performance | 🚀 Zero Allocation | 🐢 Reflection Based |
| Stateful Sequences | ✅ Yes | ⚠ Limited |
| Compile-Time Validation | ✅ Yes | ❌ No |

---

# 🤝 Contributing

Contributions are welcome!

Please read the project's **CONTRIBUTING.md** before submitting pull requests.

When contributing:

- Run all tests.
- Ensure all examples compile.
- Maintain compatibility across supported operating systems.
- Include regression tests for bug fixes.

---

# 📜 License

Licensed under the **Apache License 2.0**.

See the **LICENSE** file for details.