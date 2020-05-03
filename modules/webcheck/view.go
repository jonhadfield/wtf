package webcheck

import "fmt"

type results []string

func formatOutput(widget *Widget, url string, result string) string {
	var o string

	switch result {
	case "fail":
		if widget.settings.useEmoji {
			o = fmt.Sprintf("🔴 %s", url)
		} else {
			o = fmt.Sprintf("[red]%s", url)
		}
	case "warn":
		if widget.settings.useEmoji {
			o = fmt.Sprintf("🟠 %s", url)
		} else {
			o = fmt.Sprintf("[yellow]%s", url)
		}
	case "success":
		if widget.settings.useEmoji {
			o = fmt.Sprintf("🟢 %s", url)
		} else {
			o = fmt.Sprintf("[green]%s", url)
		}
	}

	return o
}

func (r results) Len() int {
	return len(r)
}
func (r results) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r results) Less(i, j int) bool {
	return r[j] < r[i]
}
