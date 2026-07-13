# Changelog

## v0.1.0

Initial public release.

### Fixed
- **arm64: corrected jump instruction generation.** The previous internal
  version built the target address into `x26` correctly, then executed
  `ldr x10, [x26]` followed by `br x10` — which dereferences the target
  address as a pointer and branches to whatever bytes happen to live there,
  instead of branching to the target address itself. This produced
  undefined/crashing behavior on all arm64 targets (Apple Silicon, arm64
  Linux). Fixed to branch directly (`br x26`).
- **Silent no-op on inlined targets.** Small target functions could be
  inlined by the compiler at their call sites, meaning a "successful" patch
  had no effect where it was actually called. This isn't fixable from inside
  `Patch()` — it now requires `//go:noinline` on target functions (documented
  in the README), and `Patch()` verifies its own write and panics loudly on
  any other failure instead of returning a guard that silently did nothing.

### Added
- Write-back verification after every patch.
- CI: linux/amd64, linux/arm64 (QEMU), windows/amd64, darwin/arm64 (build only).
- README, CONTRIBUTING, SECURITY docs.