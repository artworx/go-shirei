// Theme demo: widget chrome at the default accents, plus a modal focus trap.
//
//	go run .                 # GUI
//	go run . --png out.png   # headless frame
package main

import (
	"flag"
	"fmt"
	"os"

	app "go.hasen.dev/shirei/app"

	. "go.hasen.dev/shirei"
	. "go.hasen.dev/shirei/widgets"
)

func scrollDemoRows() {
	for i := 0; i < 30; i++ {
		Container(Attrs(Pad2(4, 6)), func() {
			Label(fmt.Sprintf("Row %d", i))
		})
	}
}

// filterableMenuItems: open Filterable and type e.g. "piano".
var filterableMenuItems = []string{
	"piano",
	"shirei/examples/piano",
	"git_history",
	"fontviewer",
	"ferry",
	"theme",
	"layout",
	"kanban",
	"color-picker",
	"text-fields",
	"image-viewer",
	"split-panes",
	"toast",
	"synthpad",
}

var idleText = "idle (package default: aqua)"
var focusedText = "focused, per-input meadow accent"
var pathText = "/Users/hasen"
var modalName = ""
var modalEmail = ""
var showModal = false
var checkedAqua = true
var checkedMeadow = true
var uncheckedA = false
var uncheckedB = false
var toggleOff = false
var toggleOn = true
var radioOpt = "A"
var segOpt = 10

func RootView() {

	Container(Attrs(Viewport, Background(220, 10, 97, 1), Pad(30), Gap(20)), func() {
		width := GetContentWidth()
		Container(Attrs(Row, Wrap, CrossMid, Gap(14), MaxWidth(width)), func() {
			NextAccessName("btn_lightsteel")
			ButtonWithAccent(NoIcon, "LightSteel", AccentLightSteel)
			NextAccessName("btn_slateblue")
			ButtonWithAccent(NoIcon, "SlateBlue", AccentSlateBlue)
			NextAccessName("btn_blue")
			ButtonWithAccent(NoIcon, "Blue", AccentBlue)
			NextAccessName("btn_meadow")
			ButtonWithAccent(NoIcon, "Meadow", AccentMeadow)
			NextAccessName("btn_sunshine")
			ButtonWithAccent(NoIcon, "Sunshine", AccentSunshine)
			NextAccessName("btn_plastic")
			ButtonWithAccent(NoIcon, "Plastic", AccentPlastic)

			NextAccessName("btn_disabled")
			ButtonExt("Disabled", ButtonAttrs{Disabled: true}, DefaultButtonLook())
			NextAccessName("btn_disabled_blue")
			ButtonExt("Disabled Blue", ButtonAttrs{Disabled: true, Accent: AccentBlue}, DefaultButtonLook())
			NextAccessName("btn_disabled_meadow")
			ButtonExt("Disabled Meadow", ButtonAttrs{Disabled: true, Accent: AccentMeadow}, DefaultButtonLook())
		})

		Container(Attrs(Row, CrossMid, Gap(20)), func() {
			NextAccessName("menu")
			MenuButton(MenuIcon, "Menu", func() {
				MenuItem(SymRefresh, "Refresh")
				MenuItem(SymCopy, "Copy")
				MenuItem(SymSearch, "Search")
			})
			NextAccessName("menu_list")
			MenuButtonExt("List", ButtonAttrs{Accent: AccentBlue, Icon: SymMenu}, DefaultButtonLook(), func() {
				MenuItemExt("Refresh", ButtonAttrs{Icon: SymRefresh, Accent: AccentMeadow})
				MenuItemExt("Copy", ButtonAttrs{Icon: SymCopy, Accent: AccentMeadow})
			})
			NextAccessName("menu_filter")
			MenuButton(MenuIcon, "Filterable", func() {
				_ = MenuFilterQuery()
				for _, name := range filterableMenuItems {
					if !MenuFilterMatches(name) {
						continue
					}
					MenuItem(NoIcon, name)
				}
			})

			NextAccessName("btn_data")
			CtrlButton(SymGrid, "Data", true)
			NextAccessName("btn_info")
			CtrlButton(SymInfo, "Info", true)
			NextAccessName("btn_enable")
			CtrlButton(NoIcon, "Enable", true)
		})

		idleAttrs := DefaultTextInputAttrs()
		idleAttrs.NoAutoFocus = true
		NextAccessName("field_idle")
		TextInputExt(&idleText, idleAttrs)

		focusedAttrs := DefaultTextInputAttrs()
		focusedAttrs.Accent = AccentMeadow
		NextAccessName("field_focused")
		TextInputExt(&focusedText, focusedAttrs)

		// height parity check: button and input at default sizes, side by side.
		// "Open modal" is also the focus-trap demo: Tab among the fields/buttons
		// above, open the modal, confirm focus jumps in and Tab stays inside.
		Container(Attrs(Row, CrossMid, Gap(10)), func() {
			NextAccessName("field_path")
			TextInput(&pathText)
			NextAccessName("open_modal")
			if Button(NoIcon, "Open modal") {
				modalName = ""
				modalEmail = ""
				showModal = true
			}
		})

		if showModal {
			const modalInner = 320 // Modal(360) minus 2×20 pad
			Modal(360, func() { showModal = false }, func() {
				Label("Focus trap", FontSize(14), FontWeight(WeightBold), TextColor(220, 25, 25, 1))
				Container(Attrs(MaxWidth(modalInner)), func() {
					Label("Tab should stay in this card; Escape or outside click dismisses.", FontSize(12), TextColor(220, 10, 45, 1))
				})
				attrs := DefaultTextInputAttrs()
				attrs.MinWidth = modalInner
				Label("Your name", FontSize(11), TextColor(220, 10, 45, 1))
				NextAccessName("modal_name")
				TextInputExt(&modalName, attrs)
				Label("Email address", FontSize(11), TextColor(220, 10, 45, 1))
				NextAccessName("modal_email")
				TextInputExt(&modalEmail, attrs)
				Container(Attrs(Row, CrossMid, Gap(8)), func() {
					NextAccessName("modal_cancel")
					if Button(NoIcon, "Cancel") {
						showModal = false
					}
					NextAccessName("modal_ok")
					if ButtonExt("OK", ButtonAttrs{Accent: AccentMeadow}, DefaultButtonLook()) {
						showModal = false
					}
				})
			})
		}

		Container(Attrs(Row, CrossMid, Gap(20)), func() {
			NextAccessName("check_aqua")
			CheckBoxExt(&checkedAqua, "", CheckBoxAttrs{Accent: AccentBlue, Size: 28})
			NextAccessName("check_meadow")
			CheckBoxExt(&checkedMeadow, "", CheckBoxAttrs{Accent: AccentMeadow, Size: 28})
			NextAccessName("check_a")
			CheckBoxExt(&uncheckedA, "", CheckBoxAttrs{Accent: Vec4{265, 60, 75, 1}, Size: 28})
			NextAccessName("check_b")
			CheckBoxExt(&uncheckedB, "", CheckBoxAttrs{Accent: Vec4{5, 70, 70, 1}, Size: 28})
		})

		Container(Attrs(Row, CrossMid, Gap(20)), func() {
			NextAccessName("toggle_off")
			ToggleSwitch(&toggleOff)
			NextAccessName("toggle_on")
			ToggleSwitchExt(&toggleOn, ToggleSwitchAttrs{Accent: AccentMeadow})
		})

		Container(Attrs(Row, CrossMid, Gap(30)), func() {
			OptionGroup(&radioOpt, func() {
				ModAttrs(Spacing(10))
				NextAccessName("radio_a")
				OptionButton("First", "A")
				NextAccessName("radio_b")
				OptionButton("Second", "B")
			})
			NextAccessName("seg")
			SegmentedControl(&segOpt, func() {
				NextAccessName("seg_x")
				SegmentedCell("X", 10)
				NextAccessName("seg_y")
				SegmentedCell("Y", 20)
				NextAccessName("seg_z")
				SegmentedCell("Z", 30)
			})
		})

		Container(Attrs(FixHeight(120), Expand, Clip, Background(0, 0, 100, 1), BorderWidth(1), BorderColor(0, 0, 88, 1)), func() {
			ScrollOnInput()
			ScrollBars()
			scrollDemoRows()
		})
	})
}

func main() {
	pngPath := flag.String("png", "", "write one settled frame to PATH and exit")
	flag.Parse()

	if *pngPath != "" {
		if err := RenderToPNG(*pngPath, 500, 570, RootView); err != nil {
			fmt.Println("render failed:", err)
			os.Exit(1)
		}
		return
	}

	app.SetupWindow("Theme demo", 540, 640)
	app.SetupDrive()
	app.Run(RootView)
}
