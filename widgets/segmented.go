package widgets

// SegmentedControl: adjacent flat option-buttons in a shared hairline frame,
// the shirei take on a radio row. Born in examples/piano's voice picker,
// promoted here once the design settled, then rethemed (2026-07-07) to match
// CheckBox/OptionButton's accent language: an accent-colored frame, and the
// selected segment reading as a solid accent fill (edge to edge, no margin)
// with a bold white label — exactly the filled-vs-outlined convention those
// widgets use.
//
// Interaction is ProcessSegmentEvents per cell (same family as
// ProcessButtonEvents): the caller owns each cell container; the library
// returns a rich snapshot (selected, became-selected, prev value, local
// click, selected-at time) so custom chrome can animate. Default
// SegmentedControl is a thin combination of that analysis + accent chrome.
//
// Dividers between segments are explicit accent-colored elements at the
// same thickness as the outer border (not a padding gap revealing the
// frame's background), so the selected segment's fill can run flush to the
// frame edge with no empty margin around it.

import (
	"time"

	. "go.hasen.dev/shirei"
)

// SegmentState is one cell's interaction snapshot for a segmented control.
// Call ProcessSegmentEvents inside that cell's Container (once per cell per
// frame). The function binds to the current container id.
//
// Typical custom segmented control:
//
//	for _, c := range cells {
//	    ContainerWithKey(c.Value, Attrs(...), func() {
//	        st := ProcessSegmentEvents(target, c.Value, false)
//	        if st.BecameSelected {
//	            // st.Prev is the previous selection; st.Local is click pos
//	            // in this cell; st.SelectedAt is when this value became on
//	        }
//	        // paint from st.Selected / st.Hovered / st.Active
//	    })
//	}
type SegmentState[T comparable] struct {
	Hovered  bool
	Active   bool // pointer captured on this cell
	Clicked  bool // completed press on this cell this frame
	Selected bool // *target == value after this call
	Disabled bool
	HasFocus bool
	// Local is the pointer relative to this cell's screen top-left.
	Local Vec2

	// BecameSelected is true on the frame this cell was chosen (target
	// changed from something else to value). Prev is *target just before
	// that change — useful for slide direction and other transitions.
	BecameSelected bool
	Prev           T

	// SelectedAt is when *target last became value (zero if this cell is
	// not selected, or was never chosen through ProcessSegmentEvents).
	SelectedAt time.Time
}

// ProcessSegmentEvents analyzes pointer and keyboard interaction on the
// current container as one segment of a mutually-exclusive group bound to
// *target. On a completed click (or Space/Enter while focused) that chooses
// value, it writes *target and returns BecameSelected with Prev set to the
// prior value.
//
// Pointer model matches ProcessButtonEvents. Call once per cell, inside that
// cell's container body. Arrow-key movement across the group is the stock
// SegmentedControl's job — custom skins that want it handle Left/Right on
// the frame after the cells build.
func ProcessSegmentEvents[T comparable](target *T, value T, disabled bool) SegmentState[T] {
	var st SegmentState[T]
	bst := ProcessButtonEvents(disabled)
	st.Disabled = bst.Disabled
	st.Hovered = bst.Hovered
	st.Active = bst.Active
	st.Clicked = bst.Clicked
	st.HasFocus = bst.HasFocus
	st.Local = bst.Local

	type hook struct {
		selectedAt time.Time
	}
	h := Use[hook]("seg-cell")

	if st.Clicked && *target != value {
		st.Prev = *target
		*target = value
		st.BecameSelected = true
		h.selectedAt = time.Now()
		RequestNextFrame()
	}
	st.Selected = *target == value
	if st.Selected {
		st.SelectedAt = h.selectedAt
	}
	return st
}

// SegmentedControlAttrs configures SegmentedControlExt. Start from
// DefaultSegmentedControlAttrs() and override fields — a zero layout field is
// intentional (e.g. CellPadH: 0, FrameCorners: 0), not “use stock default.”
// SegmentedControl() always passes DefaultSegmentedControlAttrs().
type SegmentedControlAttrs struct {
	Accent Vec4 // zero value: package-level DefaultAccent / Accent

	// Expand makes the control fill available width; each segment Grow(1)
	// so free space is shared evenly (MinCellWidth is still a floor).
	Expand bool

	// CellPadH is horizontal padding inside each segment (design units,
	// then × ComfortScale). DefaultSegmentedControlAttrs uses 12; 0 is flush.
	CellPadH f32

	// FrameCorners is the outer frame corner radius (design units, then ×
	// ComfortScale). End segments use a slightly smaller radius when > 0.
	// DefaultSegmentedControlAttrs uses 6; 0 is square.
	FrameCorners f32

	// MinCellWidth is the minimum width of each segment (design units, then
	// × ComfortScale). DefaultSegmentedControlAttrs uses 56.
	MinCellWidth f32
}

const segmentBorderWidth = 1.5
const segmentHeight = 24

// DefaultSegmentedControlAttrs returns the stock chrome layout (design units
// before ComfortScale). Copy and tweak for SegmentedControlExt.
func DefaultSegmentedControlAttrs() SegmentedControlAttrs {
	return SegmentedControlAttrs{
		CellPadH:     12,
		FrameCorners: 6,
		MinCellWidth: 56,
	}
}

// SegmentedControl renders the segments and keeps *target in sync with the
// clicked one. Returns true when the selection changed this frame (handy
// for reacting to the change, e.g. recomputing derived state). Values must
// be unique — they double as the segments' identity.
//
//	NextAccessName("voice")
//	SegmentedControl(&voice, func() {
//	    NextAccessName("oud")
//	    SegmentedCell("Oud", VoiceOud)
//	    NextAccessName("flute")
//	    SegmentedCell("Flute", VoiceFlute)
//	})
func SegmentedControl[T comparable](target *T, body func()) bool {
	return SegmentedControlExt(target, DefaultSegmentedControlAttrs(), body)
}

// SegmentedControlExt is SegmentedControl with per-instance accent and layout.
// Prefer DefaultSegmentedControlAttrs() as a starting point when overriding
// pad, corners, or Expand.
func SegmentedControlExt[T comparable](target *T, attrs SegmentedControlAttrs, body func()) bool {
	accent := AccentOrFallback(attrs.Accent, DefaultAccent)
	h := comfort(segmentHeight)
	minW := comfort(attrs.MinCellWidth)
	padH := comfort(attrs.CellPadH)
	frameR := comfort(attrs.FrameCorners)
	// End segments sit slightly inside the frame radius when rounded (5 vs 6).
	endR := f32(0)
	if attrs.FrameCorners > 0 {
		endR = comfort(attrs.FrameCorners * 5 / 6)
	}
	labelSize := comfort(12)
	changed := false

	run := &segmentedRun[T]{
		target:    target,
		changed:   &changed,
		accent:    accent,
		h:         h,
		minW:      minW,
		padH:      padH,
		endR:      endR,
		labelSize: labelSize,
		expand:    attrs.Expand,
	}
	prev := currentSegmented
	currentSegmented = run
	defer func() { currentSegmented = prev }()

	frame := Attrs(Row, BorderWidth(segmentBorderWidth), BorderColor(accent[0], accent[1], accent[2], accent[3]), Clip)
	if frameR > 0 {
		frame = AttrsWith(frame, Corners(frameR))
	}
	if attrs.Expand {
		frame = AttrsWith(frame, Expand)
	}

	Container(frame, func() {
		NextAccessRole("radiogroup")
		AssignAccess()

		type hook struct {
			ids    []ContainerId
			values []T
		}
		hids := Use[hook]("seg-ids")
		n := len(hids.ids)
		if n > 0 && len(hids.values) == n {
			idx := -1
			for i, id := range hids.ids {
				if IdHasFocus(id) {
					idx = i
					break
				}
			}
			if idx >= 0 {
				delta := 0
				switch GetFrameInput().Key {
				case KeyLeft, KeyUp:
					delta = -1
				case KeyRight, KeyDown:
					delta = 1
				}
				if delta != 0 {
					next := idx + delta
					if next < 0 {
						next = n - 1
					} else if next >= n {
						next = 0
					}
					if next != idx {
						*target = hids.values[next]
						FocusImmediateOn(hids.ids[next])
						changed = true
						GetFrameInput().Key = KeyCodeNone
						RequestNextFrame()
					}
				}
			}
		}
		run.lastN = n
		if body != nil {
			body()
		}
		hids.ids = append(hids.ids[:0], run.nextIDs...)
		hids.values = append(hids.values[:0], run.nextVals...)
	})
	return changed
}

var currentSegmented any

type segmentedRun[T comparable] struct {
	target                         *T
	changed                        *bool
	accent                         Vec4
	h, minW, padH, endR, labelSize f32
	expand                         bool
	lastN                          int
	index                          int
	nextIDs                        []ContainerId
	nextVals                       []T
}

// SegmentedCell paints one segment inside the current SegmentedControl body.
// Other widgets in that body become extra row children and scramble dividers
// and end radii.
func SegmentedCell[T comparable](label string, value T) {
	run, ok := currentSegmented.(*segmentedRun[T])
	if !ok || run == nil || run.target == nil {
		panic("widgets: SegmentedCell must be called from SegmentedControl")
	}
	if run.index > 0 {
		Element(Attrs(FixWidth(segmentBorderWidth), FixHeight(run.h), BackgroundVec(run.accent)))
	}
	var rl, rr f32
	if run.index == 0 {
		rl = run.endR
	}
	if run.lastN > 0 && run.index == run.lastN-1 {
		rr = run.endR
	}
	ch, id := segmentOption(run.accent, run.target, value, label, rl, rr, run.h, run.minW, run.padH, run.labelSize, run.expand)
	if ch {
		*run.changed = true
	}
	run.nextIDs = append(run.nextIDs, id)
	run.nextVals = append(run.nextVals, value)
	run.index++
}

// segmentOption is one segment; rl/rr round the outer corners of the end
// segments so they follow the frame's radius. Returns whether this cell
// became the selection this frame, and the cell's identity.
func segmentOption[T comparable](accent Vec4, target *T, value T, label string, rl, rr, h, minW, padH, labelSize f32, expand bool) (bool, ContainerId) {
	changed := false
	var id ContainerId
	cell := Attrs(FixHeight(h), MinWidth(minW), CrossAlign(AlignMiddle), Pad2(0, padH), Corners4(rl, rr, rr, rl))
	if expand {
		cell = AttrsWith(cell, Grow(1))
	}
	ContainerWithKey(value, cell, func() {
		id = CurrentId()
		st := ProcessSegmentEvents(target, value, false)
		NextAccessRole("radio")
		NextAccessChecked(st.Selected)
		AssignAccess()
		changed = st.BecameSelected

		bg := Vec4{0, 0, 100, 1}
		grad := Vec4{0, 0, -12, 0}
		textClr := TextColor(0, 0, 25, 1)
		weight := WeightNormal
		if st.Hovered && !st.Selected {
			bg = Vec4{accent[0], accent[1] * 0.3, 96, 1}
		}
		if st.Selected {
			bg = accent
			grad[2] = 12
			textClr = TextColor(0, 0, 100, 1)
			weight = WeightBold
			if st.Hovered {
				bg[2] += 5
			}
		}
		ModAttrs(BackgroundVec(bg), GradVec(grad))
		if st.HasFocus {
			ModAttrs(BorderWidth(2), BorderColorVec(FocusRing))
		}

		Filler(1)
		Label(label, FontSize(labelSize), textClr, FontWeight(weight))
		Filler(1)
	})
	return changed, id
}
