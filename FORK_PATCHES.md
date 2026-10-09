# Daymark fork: v0.8.0 patch audit

Status: Dated verification record. Decisions refer to the source revisions below;
inspect Git for current publication and branch state.

## October 9, 2026: v0.8.0 rebase

Upstream: `58b0835` (`upstream/master`, containing release tag `v0.8.0` at
`073d468`). Previous fork: `2bf89e0`, based on upstream `53ac833` (v0.7.0).
The previous fork is preserved on `backup/pre-v0.8.0-2026-10-09`. Local `master`
contains the replayed patches on v0.8.0. Daymark integration is committed as
`2e3b27e`. Nothing has been pushed.

Every fork-only commit was compared against the v0.8.0 source and its tests.
None is completely superseded. "Keep" refers to its surviving contribution,
not to reviving implementation removed by the earlier v0.7.0 adaptation.
The v0.8.0 upstream delta leaves fonts, image loading, resources, Cocoa input,
quit handling and software image rendering unchanged from the fork's upstream
base. Its text changes add explicit selection paint; they do not replace the
fork's shaping and viewport optimizations.

| Previous commit | Decision | Evidence and retained contribution |
| --- | --- | --- |
| `aaa33de` resolved-style scans | Keep | Upstream `styleAt` still walks spans; binary paint lookup and adjacent-run coalescing remain useful. Resolved-span tests pass. |
| `230efc4` boundary-event span flattening | Keep | Upstream `flattenStyleSpans` still visits all spans at each breakpoint. Preserve ordered boundary events and overlapping-style tests. |
| `92eaa7f` per-pass shaping lookups | Keep | Upstream family interning does not supply the fork's forward style traversal or per-rune face/glyph memoization. |
| `c99d1d5` reusable shaped segments | Keep | Upstream has wrapped/unwrapped paragraph caches, with no edit-reusable segment cache. Segment reuse tests pass. |
| `a56a3a8` bounded large-document caches | Keep | Upstream puts all paragraphs in the shared caches. Keep two-entry large wrapped/unwrapped LRUs and their eviction tests. |
| `162760c` incremental equal-length edits | Keep | Upstream reshapes after paragraph-key changes. Conservative ASCII overwrite path, immutable unchanged-line sharing and paint-geometry equality tests remain needed. |
| `55e4693` per-line bidi cache | Keep | Upstream `ParagraphBidi` reruns every line after an edit. Keep the bounded per-line direction cache. |
| `ec65f84` resolved-span entry points | Keep | Upstream's flattened shaping path is private. Daymark still consumes the public resolved-style API. |
| `80cf982` viewport layout | Keep | Upstream builds every line. Daymark's editor requires viewport-only layout with full-document height and ink padding. |
| `f0aba45` asynchronous image reads | Keep | Upstream reads whole files on the frame thread. Keep header-only sizing and background file/pixel work. |
| `e7bf0b4` unused decorations | Keep; adapt helper | Upstream still scans advance bands for unused span decorations. Apply the fork checks in the shared aligned/selection-color implementation. |
| `2052a7e` nearest font face | Keep | Upstream family resolution still asks for exact face aspects. Preserve nearest stretch/style/weight matching and registry invalidation. Font-only source overlays attribute showcase/example differences exactly. |
| `f43dbea` opaque blits | Keep surviving portion | Preserve opacity promises and rounded-clip interior row copies. The v0.7.0 integration already removed old channel-conversion assembly; upstream's ordered-pixel cache remains authoritative. |
| `af8bec6` asynchronous small images | Keep | Upstream decodes images below a compressed-byte threshold synchronously. Small compressed size does not bound decode work; headless snapshots remain synchronous. |
| `8aa84b7` v0.7.0 integration | Keep; adapt helper | Per-line paint-span hashing, font-epoch invalidation, safe incremental paint refresh and ordered-image adaptation still apply. Filtering moves inside the shared helper so explicit-color layout also stays fast. |
| `32e77ce` aligned text | Keep; integrate selection paint | Upstream still lacks explicit shaped-line alignment and aligned cursor mapping. Merge alignment with upstream's literal selection-color API and retain stock input defaults. |
| `65d7b7e` leading / Unicode wrapping | Keep | Upstream lacks the fork's explicit `LineHeight` and CJK break behavior. Leading, wrapping and cache-key tests remain required. |
| `5188118` fork agent guidance | Keep | Portable development and consumer verification instructions are fork maintenance requirements. |
| `7a28bcd` oversized focus reveal | Keep | Upstream `revealDelta` still snaps oversized visible content. Preserve scroll position during selection while retaining reveal of offscreen targets. Keep both upstream focus-indicator tests and fork reveal tests. |
| `72de245` Control-Tab | Keep | Upstream Cocoa lacks the native key-equivalent override and pending-key modifier capture. Preserve plain Tab navigation alongside application shortcuts. |
| `f5081f7` quick mouse edges | Keep | Upstream can overwrite press with release before the next display tick. Keep queued edges and native-event regression coverage. |
| `36c9649` deferred quit | Keep | Upstream has immediate native exit. Daymark needs application approval to finish durable saves before closing; generic cleanup still handles accepted exit. |
| `2bf89e0` nonbreaking references | Keep | Upstream segment wrapping does not preserve glue across style/font boundaries. Keep NBSP, narrow NBSP and word-joiner grouping and incremental-edit metadata. |

## v0.8.0 integration changes

The text layout conflict combines line alignment with explicit selection paint.
`ShapedTextLayoutStyled` retains upstream's literal transparent-zero semantics;
`ShapedTextLayoutStyledAligned` permits both capabilities together. Stock text
inputs keep direction-aware default alignment. Per-line span filtering remains
inside the common implementation, including the explicit-color path.

A new selection/alignment regression test exposed a carried-forward bug:
centered/end-aligned glyphs moved while selection/decorations stayed at the line
origin. `emitTextRuns` now shares the horizontal glyph offset with paint bands.
The test covers single/multiple lines, all explicit alignments and transparent
selection. Light and dark captures of these aligned selections were inspected. It failed before the fix and passes afterward.

Daymark publishes precomputed light/dark `CurrentColorScheme` values derived
from its own semantic palette. This replaces the removed `ButtonAccent` global;
app-owned button rendering reads `T().ButtonAccent`. Menus, fields, modal text
and scrim follow the app's appearance. Its module requirement is v0.8.0 with
the existing sibling replacement. Framework stock controls use upstream's new
focus-indicator semantics; editor carets continue to depend on keyboard focus.

## Verification environment and controls

Host: Apple M4 Pro, macOS/arm64. Compiled framework and Daymark benchmark binaries report Go 1.27.2.
All before/after binaries use the same toolchain. Tests use isolated temporary configuration and fixture
data. Baseline framework source was exported from the backup branch. Untouched
v0.8.0 source was exported from `upstream/master`. Temporary workspaces include
all separately versioned framework examples/extensions so they consume these
local sources rather than an installed module.

Full functional suites, targeted race checks, platform builds, snapshot
attribution and measurements are recorded below. Snapshot goldens are unchanged.

### Snapshot attribution

The root framework suite reports eight controls-showcase and six widget snapshot
mismatches. All six widget actuals are byte-identical to untouched v0.8.0 on this
host. All eight showcase actuals are byte-identical to that control when only
nearest-face resolution is disabled through a Go source overlay. Light and dark
showcases were inspected: the retained patch selects the requested bold face
instead of falling back to regular weight.

All nested modules were also tested with the local source. Both fork and pristine
upstream produce snapshot failures on this host. Of 108 comparable actual PNGs,
48 match byte-for-byte and 60 differ with the normal fork. With only nearest-face
resolution disabled, all 94 comparable nested-example PNGs match byte-for-byte
(eight showcase and six widget PNGs are covered separately). Haystack's grouped
search drive test also fails on both fork and untouched upstream. These are
attributed control failures, not blanket exemptions for later changes.

Daymark's main workspace, task dialog, split Markdown editor/viewer and Ask
answer were rendered and inspected in both appearances. Existing scroll,
selection/copy, status-reference, table/cursor and shortcut regressions exercise
interaction beyond the static captures.

### Race-test baseline

The broad framework race run fails on unsynchronized test access to the shared
file caches and on a UDP listener surviving into the following input test.
The original fork reproduces these same race locations. The broad Daymark root
race run exposes status/projection timer callbacks outliving tests and racing
with subsequent global-state changes; the original app/fork reproduces them.
The full viewer race suite exceeds architecture-render timeout/goroutine-settle
limits on both framework revisions. These failures predate this upgrade.

Isolated retained-patch race tests, Cocoa/app race checks, editor race tests,
Daymark appearance/typography/tab/Ask selection/quit checks and Android bridge
race checks pass. The full normal suites exercise the diagram cases without
race-detector instrumentation. No full-suite race-clean claim is made.

### Reproduction

Run from the fork (root and separately versioned modules are distinct):

```sh
go test ./...
go test ./... -skip '^TestSnapshot'
go test -race . -run '^Test(Resolved|Flatten|Shape|Large|Bidi|Calculate|FontEpoch|Nonbreaking|UseOpaque|Blit|LoadImageDoesNot|LookupClosest|SystemFontParser|Quit|FocusReveal|RevealDelta|ControlFocus)'
go test -race ./cocoabackend ./app
go test . -run '^$' -bench 'BenchmarkBlit.*Diagram' -benchmem
```

To test nested modules against this fork, create a temporary `go.work` containing
the root and every directory with a `go.mod`, set `GOWORK` to it, and run
`go test ./...` from each module. An ordinary root `go test ./...` excludes these
nested modules as of v0.8.0.

Run from Daymark:

```sh
go test ./...
go test -race . -run '^Test(SyncTheme|Theme|Typography|WorkspaceTabShortcut|Quit|CmdQ|Hindsight.*(Scroll|Selection|Copy))'
go test -race ./internal/markdowneditor
go test ./internal/markdowneditor -run '^$' -bench '^BenchmarkMarkdownLargeFileFrame$' -benchmem -benchtime=20x
go test ./internal/markdownviewer -run '^$' -bench '^Benchmark(ViewerLargeDocumentFrame|ViewerLargeSingleFenceFrame|ArchitectureExpandedModalScrollPaint)$' -benchmem -benchtime=100x
go test . -run '^$' -bench '^BenchmarkDaymarkFixtureFrame$' -benchmem -benchtime=100x
make build build-darwin build-windows
python3 scripts/check-docs.py --workspace
```

Run `sh scripts/test.sh` from Android for bridge race checks, Kotlin unit tests,
lint and debug APK assembly. Native Android UI was not changed; device tests
and backend integration scenarios were not part of this framework upgrade.

### Performance

Four pairs of runs alternate after/before and before/after order, with no
competing test/build jobs during timing. Medians below use 20 iterations for
each large-editor scenario, 100 for whole-app/viewer/modal scenarios and
100 ms for direct blits. The viewer control waits for the asynchronous system
font scan via an identical temporary benchmark overlay on both revisions.
The whole-app fixture benchmark includes theme publication and app chrome.

| Scenario | Before | v0.8.0 fork | Change |
| --- | ---: | ---: | ---: |
| Editor steady | 7.000 ms | 6.696 ms | -4.3% |
| Editor scroll | 6.344 ms | 6.236 ms | -1.7% |
| Editor equal-length overwrite | 18.618 ms | 19.062 ms | +2.4% |
| Editor insertion | 177.916 ms | 176.888 ms | -0.6% |
| Whole-app fixture (light) | 0.172 ms | 0.168 ms | -2.5% |
| Whole-app fixture (dark) | 0.162 ms | 0.158 ms | -2.3% |
| Opaque diagram blit | 0.146 ms | 0.152 ms | +3.7% |
| Rounded opaque diagram blit | 0.196 ms | 0.201 ms | +2.5% |
| Viewer steady (400 sections) | 0.208 ms | 0.216 ms | +3.5% |
| Viewer scroll (400 sections) | 1.091 ms | 1.126 ms | +3.2% |
| Large code fence steady | 1.398 ms | 1.307 ms | -6.5% |
| Large code fence scroll | 1.354 ms | 1.292 ms | -4.6% |
| Expanded diagram scroll + software paint | 4.646 ms | 4.695 ms | +1.1% |

There is no material slowdown across the tested scenarios. Small timing
differences reverse between runs; the initial single-order run showed a 15%
steady-editor increase that did not reproduce in the alternating measurements.
Large-editor allocations are unchanged: about 32.04 MB/steady frame, 66.55 MB
per overwrite and 305.11 MB per insertion. Overwrites still shape one segment.
The rounded opaque blit remains approximately 0.20 ms versus 3.14 ms for the
general path. Insertions still exceed an interactive frame budget; the upgrade
does not claim to solve that existing limitation. Native compositor FPS and
end-to-end presentation latency were not measured.

### Final checks

- Daymark `go test ./...`: passes, including the new stock-scheme switch test.
- Framework `go test ./...`: functional checks pass; only the 14 attributed
  root snapshot cases fail (eight controls + six widgets).
- Framework `go test ./... -skip '^TestSnapshot'`: passes, including native
  controls/theme drive tests. Final root/widgets/Cocoa/app checks also pass.
- All 18 separately versioned modules tested against the local fork compile.
  Seven suites pass; eleven retain snapshot failures. Those failures and the
  Haystack grouped-search drive failure reproduce in pristine v0.8.0 controls;
  all extensions and separately versioned demos pass.
- Targeted race checks for retained patches, Cocoa/app, Daymark appearance,
  typography, tab shortcuts, Ask selection and quit: pass. Editor race suite
  and Android bridge race suite pass. Broad baseline races/timeouts are listed
  above rather than hidden by the focused runs.
- Android `sh scripts/test.sh`: passes (bridge race, Kotlin tests, lint, debug APK).
- Daymark Linux/amd64, macOS/arm64 and Windows/amd64 builds: pass.
- Documentation check: 115 Markdown documents across three repositories,
  zero errors. Changed Go files formatted; Git whitespace checks pass.

Local release builds still replace Shirei with the sibling checkout. The
release workflow clones the published fork master, which has not been updated;
publishing requires coordinating these local framework and app commits.

---

# Historical v0.7.0 patch audit

Status: Historical audit of the revisions below. Branch, publication, test counts,
and benchmark results describe that audit; inspect current Git state and rerun
applicable checks for a new change.

Updated 2026-09-19 from `364660b` (v0.6.0) plus the local patch stack through
`8f7ea06` to upstream `53ac833` (latest upstream master, v0.7.0).
The original checkout is preserved on `master`; the updated checkout is on
`upgrade/shirei-v0.7.0`. Nothing has been pushed.

## Original patches

| Original commit | Decision | Reason |
| --- | --- | --- |
| `f086d82` resolved-style scans | Keep partly | Upstream now builds style runs linearly. Retain binary lookup for paint and the coalescing behavior tested by the fork. |
| `20b821d` boundary-event span flattening | Keep | Upstream still scans all spans at each breakpoint. |
| `a35b314` per-pass shaping lookups | Keep partly | Upstream interns/caches font-family resolution. Retain forward span traversal and per-rune glyph/face memoization using interned family IDs. |
| `631fd02` segment shaping cache | Keep | Upstream caches whole paragraphs, not reusable segments across edits. Include the new primary-metrics font in segment keys. |
| `bee5818` bounded large-document cache | Keep and adapt | Both wrapped and upstream's new unwrapped large-document caches are capped at two entries; segment cache stays bounded at 524,288 entries. |
| `90f1f7c` incremental equal-length edits | Keep and adapt | Refresh paint geometry only for changed lines, preserve immutable data for unchanged lines, and invalidate on font-registry changes. Conservative fast path permits same-script ASCII letter/digit overwrites; other edits take the complete path. |
| `22fe2d7` per-line bidi cache | Keep | Paragraph cache hits do not help after an edit. Keep 65,536 cached lines; do not restore the old separate whole-document bidi cache. |
| `5fb76e2` resolved-span shaping | Keep public API | Upstream has an internal flattened-span path; expose the fork's resolved-span entry points through it. |
| `68df1f2` viewport text layout | Keep and adapt | Upstream does not virtualize this app-owned editor. Match upstream's symmetric ink padding and expose full layout height/padding for the app. |
| `c266e4b` asynchronous image reads | Keep | File reads and pixel decoding must stay outside the frame lock. |
| `e5fdeb7` skip unused decorations | Keep partly | Apply the checks to upstream's compact paint bands, not the removed per-glyph container implementation. |
| `82653b3` nearest font face | Keep partly | Retain nearest stretch/style/weight matching and its cache under upstream's registry lock. Upstream already memoizes parser failures and provides a richer fallback system; preserve those implementations. |
| `2be8390` shaped line leading | Drop | Upstream includes leading in the line's minimum height. |
| `a9b4468` opaque image blits | Keep partly | Drop ARM RGBA/BGRA assembly and old per-frame conversion. Upstream caches channel ordering and copies opaque rectangular rows. Retain the caller opacity hint and rounded-clip interior row copies using already ordered pixels. |
| `a9ca24b` Cocoa termination hook | Drop | Upstream calls generic exit cleanup for native termination and `app.Quit`. Daymark registers its cleanup with `generic.AddExitCleanup` on every platform. |
| `8f7ea06` asynchronous decoding of small images | Keep | Compressed size does not bound decode cost. Keep synchronous completed pixels only for headless snapshots. |

## Integration fixes

Limit paint/color-cache hashing to spans intersecting each line. Without this,
upstream's new glyph-color cache scanned the entire large document for every
visible line and steady frames regressed to approximately 91 ms. The corrected
path takes approximately 6 ms.

The accompanying Daymark changes use interned font styles, `IconGlyph`, upstream
exit cleanup and focus navigation. Its custom segmented controls remain app-owned.
Translated editor glyph blocks have explicit full-content bounds and no inner
clip; their viewport still clips them. Selection uses a stable outer container
when upstream switches between single-line and selected-text layout structures.

## September 30, 2026: focus reveal for oversized content

Focus reveal preserves the current scroll offset on an axis when a focus target
is at least as large as the scroll port and intersects its content clip. This
prevents clicking a long Daymark Ask Hindsight answer from snapping its top into
view and changing the text under a selection drag. Small controls and entirely
off-screen targets retain the existing reveal behavior. Framework regression
coverage checks visible and off-screen bounds; the desktop regression exercises
scrolling, selection and copy through the Ask answer card in both themes.

Validated on macOS/arm64 with Go 1.27.0: desktop and Android bridge suites pass;
the framework suite passes except for 13 snapshot mismatches. Each mismatch
reproduces with byte-identical PNGs using an overlay of the unchanged framework
source. Light and dark selected-answer captures were inspected.

## October 1, 2026: Control-Tab delivery and shortcut focus

The Cocoa view now handles Ctrl+Tab and Ctrl+Shift+Tab in
`performKeyEquivalent:` and forwards them to the normal key-down path. AppKit
can otherwise consume these keys in its native key-view loop before `keyDown:`
reaches Shirei. Other key equivalents continue through the superclass.

Shirei's own control-focus ring also runs only for plain Tab and Shift+Tab.
Modified Tab shortcuts reach the application without scheduling a focus change
before its frame builder can consume the key.

Cocoa retains the modifier flags from the pending `keyDown` while delivering
that key to the application frame. Previously a `flagsChanged` release before
the display tick replaced Ctrl+Tab or Ctrl+Shift+Tab with plain Tab. The frame
temporarily uses the captured key modifiers, then restores the live held-key
state; this does not change the backend's existing single pending-key policy.

Regression coverage constructs native NSEvents and dispatches them through the
registered view's `performKeyEquivalent:` selector, then delivers the frame
after modifier release. Both shortcuts fail with the previous AppKit source
and pass with the override. Separate tests cover key down/up modifier retention,
unhandled key equivalents, and control-focus preservation when the application
consumes modified Tab. Plain Tab and Shift+Tab retain their existing behavior.

`go test ./cocoabackend`, `go test -race ./cocoabackend`, and all 1,889 Daymark
desktop tests pass. The full framework suite has 13 snapshot differences;
their actual images are byte-identical with all three keyboard changes removed
through a source overlay. The native-selector tests do not open a window;
end-to-end verification in the user's `go run .` window remains outstanding.

## Validation of the September 19 upgrade

- Daymark: `go test ./...` passes, 1,701 tests across 36 packages.
- Daymark: `make build build-darwin build-windows` succeeds for Linux/amd64,
  macOS/arm64 and Windows/amd64.
- Light and dark headless previews rendered and visually inspected.
- Fork: 502 tests pass, 10 skip, and 13 snapshot comparisons fail. An untouched
  upstream `53ac833` checkout has the same 13 failures on this macOS host.
  Eleven actual PNGs are byte-identical to that control. The two piano images
  retain the fork's nearest-weight font behavior (bold key labels); both versions
  were inspected. Snapshot baselines were not rewritten.
- Regression tests verify one-segment edits, equality with full layout/paint
  geometry, unchanged-line sharing, font-epoch invalidation, image blending,
  viewport height, long-field horizontal scrolling, caret alignment and selection.

## Performance

Apple M4 Pro, Go 1.27.0, macOS/arm64. Same 621,581-byte Markdown fixture;
five iterations per scenario, three repetitions; medians below. Updated
benchmarks wait for the new asynchronous system-font scan to finish, matching
the old fork's synchronous startup. These are headless frame-production timings,
not native GPU FPS measurements.

| Scenario | Original fork | Updated fork |
| --- | ---: | ---: |
| Steady frame | 8.25 ms | 5.95 ms |
| Scrolling frame | 8.24 ms | 5.38 ms |
| Equal-length visible edit | 26.20 ms | 19.05 ms |
| Length-changing insertion | 197.74 ms | 156.93 ms |
| Segments shaped per visible overwrite | 1 | 1 |

Steady-frame allocation drops from 37.95 MB to 32.04 MB. Insertion allocation
increases from 255.65 MB to about 295.82 MB because the updated pipeline retains
additional unwrapped and paint geometry; both large-document caches remain
bounded. Arbitrary insertion still exceeds a 16.7 ms frame budget.

For 1600×1000 opaque images, direct painting remains around 0.15–0.16 ms.
The retained rounded-clip path takes about 0.21 ms versus 3.21 ms through the
general blend path. The new renderer expects preordered pixels, so these direct
blit measurements exclude upstream's one-time channel conversion.

Reproduce from Daymark:

```sh
go test ./internal/markdowneditor -run '^$' \
  -bench '^BenchmarkMarkdownLargeFileFrame$' -benchmem -benchtime=5x -count=3
```

Reproduce from this fork:

```sh
go test . -run '^$' -bench 'BenchmarkBlit.*Diagram' \
  -benchmem -benchtime=100ms -count=3
```

Daymark's release workflow still clones the published fork's `master`. Publishing
this local update requires coordinating the fork and app changes; the local
build uses the updated sibling checkout through its existing module replacement.

## October 1, 2026: nonbreaking status references

Nonbreaking spaces, narrow nonbreaking spaces and word joiners preserve their
no-break boundary across font and style segments. Wrapping fits each glued
group as a unit, so a Daymark task status icon cannot stay on the previous line
when its first label word wraps. Hard newlines retain their existing behavior;
ordinary segment wrapping is unchanged. The scan is linear and adds no paragraph
buffer. Incremental ASCII edits retain the segment's glue metadata.

Desktop regression coverage checks equal icon slots, shared baselines, wrapping,
click targets and copying authored text without generated status decorations.
The framework regression fails with the unchanged source and passes with this
patch. Desktop package tests, focused race checks and Android bridge race tests
pass. All 13 framework snapshot mismatches reproduce with byte-identical actual
PNGs under an unchanged-source overlay; snapshot baselines were left intact.

## October 2, 2026: Cocoa quick-click edge delivery

The Cocoa backend queues mouse button edges and delivers one edge per frame.
AppKit can report both mouse-down and mouse-up for a trackpad tap before the
display tick; storing a single transient action allowed the release to replace
the press, so `PressAction` never completed. Ordered delivery makes tap-to-click
and physical clicks equivalent for primary and secondary buttons while leaving
macOS responsible for gesture recognition.

Regression coverage submits down and up without an intervening frame, verifies
primary and secondary edge order, completes one ordinary `PressAction`, and
checks that single edges do not repeat. On macOS/arm64 with Go 1.27.0, the Cocoa
package and race suites pass, as do all 1,946 Daymark desktop tests. A native
window smoke verifies primary click, secondary click, and drag-release. The full
framework run retains 13 pre-existing snapshot mismatches; its unrelated Metal
glyph-clip test also reproduces a Go runtime “pointer to free object” failure
when run alone.

## October 3, 2026: deferred application quit

`SetQuitHandler` lets Daymark retain the window while its shared object store
flushes memory-only edits. Native Cocoa close/menu termination, X11 close,
Wayland close and client decoration, Win32 close, and `app.Quit` delegate requests
without stopping the event loop. Applications without a handler retain upstream
exit cleanup. The application owns saving, conflict recovery and final exit.

Verification is recorded in [Daymark's shared-object verification report](../daymark-ui/docs/application-state-verification.md). The
existing rendering/text patch stack and snapshot baselines remain unchanged.
