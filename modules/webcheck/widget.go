package webcheck

import (
	"fmt"
	"github.com/rivo/tview"
	"github.com/wtfutil/wtf/view"
	"sort"
	"strings"
)

type Widget struct {
	view.KeyboardWidget
	view.MultiSourceWidget
	view.TextWidget

	settings *Settings
}

// NewWidget creates a new instance of a widget
func NewWidget(app *tview.Application, settings *Settings) *Widget {
	widget := Widget{
		TextWidget: view.NewTextWidget(app, settings.common),
		settings:   settings,
	}

	widget.settings.common.RefreshInterval = 30
	widget.View.SetInputCapture(widget.InputCapture)

	widget.SetDisplayFunction(widget.Refresh)
	widget.View.SetWordWrap(true)
	widget.View.SetWrap(settings.wrapText)

	widget.KeyboardWidget.SetView(widget.View)

	return &widget
}

/* -------------------- Exported Functions -------------------- */

func (widget *Widget) Refresh() {
	if widget.Disabled() {
		return
	}
	widget.Redraw(widget.content)
}

func (widget *Widget) HelpText() string {
	return widget.KeyboardWidget.HelpText()
}

/* -------------------- Unexported Functions -------------------- */

func (widget *Widget) content() (string, string, bool) {
	title := widget.CommonSettings().Title

	client := getClient(widget.settings)

	var outList []string

	var ch = make(chan string)
	for _, url := range widget.settings.urls {
		go func(url string) {
			var o string

			switch checkURL(client, url, widget.settings.warnCodes) {
			case "fail":
				o = fmt.Sprintf("🔴 %s", url)
			case "warn":
				o = fmt.Sprintf("🟠 %s", url)
			case "success":
				o = fmt.Sprintf("🟢 %s", url)
			}
			ch <- o
		}(url)
	}

	var res string

	for i := 1; i <= len(widget.settings.urls); i++ {
		res = <-ch
		outList = append(outList, res)
	}

	sort.Strings(outList)

	output := strings.Join(outList, "\n")

	return title, output, widget.settings.wrapText
}
