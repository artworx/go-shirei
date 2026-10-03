# Daymark fork: v0.7.0 patch audit

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
