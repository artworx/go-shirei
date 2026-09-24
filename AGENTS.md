# Daymark Shirei fork

This is the Shirei framework checkout consumed by the sibling Daymark desktop.
Read [README.md](README.md) for framework concepts and
[FORK_PATCHES.md](FORK_PATCHES.md) before changing patched rendering, shaping,
image loading, or platform lifecycle behavior.

The fork audit records a dated comparison, not current branch or publication
state. Inspect Git and the consumer's module/build configuration for those facts.
For Daymark integration, read [its agent guide](../daymark-ui/AGENTS.md) and
[architecture map](../daymark-ui/docs/architecture.md).

Preserve app-owned versus framework-owned responsibilities. Keep expensive
image/font/layout preparation outside the interactive frame lock where the
existing async pipeline supports it; preserve cancellation and cache invalidation.

Format changed Go files and run affected package tests. Rendering or shaping
changes also need the corresponding snapshots/benchmarks and affected Daymark
tests plus visual inspection. For shared framework changes, run `go test ./...`
and report platform prerequisites or failures separately.

When snapshots differ, inspect the images and establish whether the difference
comes from the patch or the host/upstream baseline before updating expected output.
A dated baseline failure list does not exempt a new failure. Record new fork
tradeoffs and reproducible verification in the patch audit when modifying the patch stack.

Docs-only changes require link and consistency checks, not rendering tests.
