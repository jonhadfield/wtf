package awsbilling

import (
	"strings"

	"github.com/rivo/tview"
	"github.com/wtfutil/wtf/view"
)

type Widget struct {
	view.KeyboardWidget
	view.MultiSourceWidget
	view.TextWidget

	settings *Settings
}

// NewWidget creates a new instance of a widget
func NewWidget(app *tview.Application, pages *tview.Pages, settings *Settings) *Widget {
	widget := Widget{
		KeyboardWidget: view.NewKeyboardWidget(app, pages, settings.common),
		TextWidget:     view.NewTextWidget(app, settings.common),
		settings:       settings,
	}

	widget.settings.common.RefreshInterval = 30
	widget.View.SetInputCapture(widget.InputCapture)
	widget.initializeKeyboardControls()
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

	var sb strings.Builder

	for _, i := range widget.settings.accounts {
		acc, err := parseAccountSettings(i)
		if err != nil {
			return title, err.Error(), widget.settings.wrapText
		}

		var accOut string

		accOut, err = getAccountOutput(widget, acc)
		if err != nil {
			return title, err.Error(), widget.settings.wrapText
		}

		sb.WriteString(accOut + "\n")
	}

	return title, sb.String(), widget.settings.wrapText
}
