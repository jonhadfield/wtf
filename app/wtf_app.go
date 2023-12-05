package app

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	_ "github.com/gdamore/tcell/terminfo/extended"
	"github.com/gdamore/tcell/v2"
	"github.com/olebedev/config"
	"github.com/rivo/tview"
	"github.com/wtfutil/wtf/cfg"
	"github.com/wtfutil/wtf/support"
	"github.com/wtfutil/wtf/utils"
	"github.com/wtfutil/wtf/wtf"
	"log"
	"os"
	"time"
)

// WtfApp is the container for a collection of widgets that are all constructed from a single
// configuration file and displayed together
type WtfApp struct {
	TViewApp *tview.Application

	config         *config.Config
	configFilePath string
	display        *Display
	focusTracker   FocusTracker
	ghUser         *support.GitHubUser
	pages          *tview.Pages
	validator      *ModuleValidator
	widgets        []wtf.Wtfable

	// The redrawChan channel is used to allow modules to signal back to the main loop that
	// the screen needs to be explicitly redrawn, instead of waiting for tcell to redraw
	// on a user event, because something has visually changed
	redrawChan chan bool
}

// NewWtfApp creates and returns an instance of WtfApp
func NewWtfApp(tviewApp *tview.Application, config *config.Config, configFilePath string) *WtfApp {
	wtfApp := &WtfApp{
		TViewApp: tviewApp,

		config:         config,
		configFilePath: configFilePath,
		pages:          tview.NewPages(),

		redrawChan: make(chan bool, 1),
	}

	wtfApp.TViewApp.SetBeforeDrawFunc(func(s tcell.Screen) bool {
		s.Clear()
		return false
	})

	wtfApp.widgets = MakeWidgets(wtfApp.TViewApp, wtfApp.pages, wtfApp.config, wtfApp.redrawChan)
	if len(wtfApp.widgets) == 0 {
		fmt.Println("No modules were defined. Make sure you have at least one properly defined widget")
		os.Exit(1)
	}

	wtfApp.display = NewDisplay(wtfApp.widgets, wtfApp.config)
	wtfApp.focusTracker = NewFocusTracker(wtfApp.TViewApp, wtfApp.widgets, wtfApp.config)
	wtfApp.validator = NewModuleValidator()

	githubAPIKey := readGitHubAPIKey(wtfApp.config)
	wtfApp.ghUser = support.NewGitHubUser(githubAPIKey)

	wtfApp.pages.AddPage("grid", wtfApp.display.Grid, true, true)

	wtfApp.validator.Validate(wtfApp.widgets)

	firstWidget := wtfApp.widgets[0]
	wtfApp.pages.Box.SetBackgroundColor(
		wtf.ColorFor(
			firstWidget.CommonSettings().Colors.WidgetTheme.Background,
		),
	)

	wtfApp.TViewApp.SetInputCapture(wtfApp.keyboardIntercept)
	wtfApp.TViewApp.SetRoot(wtfApp.pages, true)

	// Create a watcher to handle calls to redraw the screen
	go handleRedraws(wtfApp.TViewApp, wtfApp.redrawChan)

	return wtfApp
}

func handleRedraws(tviewApp *tview.Application, redrawChan chan bool) {
	if redrawChan == nil {
		return
	}

	for {
		data := <-redrawChan

		if data {
			tviewApp.Draw()
		}
	}
}

/* -------------------- Exported Functions -------------------- */

// Exit quits the app
func (wtfApp *WtfApp) Exit() {
	wtfApp.Stop()
	wtfApp.TViewApp.Stop()
	wtfApp.DisplayExitMessage()
	os.Exit(0)
}

// Execute starts the underlying tview app
func (wtfApp *WtfApp) Execute() error {
	if err := wtfApp.TViewApp.Run(); err != nil {
		return err
	}

	return nil
}

// Start initializes the app
func (wtfApp *WtfApp) Start() {
	go wtfApp.scheduleWidgets()
	go wtfApp.watchForConfigChanges()

	// FIXME: This should be moved to the AppManager
	go func() { _ = wtfApp.ghUser.Load() }()
}

// Stop kills all the currently-running widgets in this app
func (wtfApp *WtfApp) Stop() {
	wtfApp.stopAllWidgets()
	close(wtfApp.redrawChan)
}

/* -------------------- Unexported Functions -------------------- */

func (wtfApp *WtfApp) stopAllWidgets() {
	for _, widget := range wtfApp.widgets {
		widget.Stop()
	}
}

func (wtfApp *WtfApp) keyboardIntercept(event *tcell.EventKey) *tcell.EventKey {
	// These keys are global keys used by the app. Widgets should not implement these keys
	switch event.Key() {
	case tcell.KeyCtrlC:
		wtfApp.Stop()
		wtfApp.TViewApp.Stop()
		wtfApp.DisplayExitMessage()
	case tcell.KeyCtrlR:
		wtfApp.refreshAllWidgets()
		return nil
	case tcell.KeyCtrlSpace:
		// FIXME: This can't reside in the app, the app doesn't know about
		// the AppManager. The AppManager needs to catch this one
		fmt.Println("Next app")
		return nil
	case tcell.KeyTab:
		wtfApp.focusTracker.Next()
	case tcell.KeyBacktab:
		wtfApp.focusTracker.Prev()
		return nil
	case tcell.KeyEsc:
		wtfApp.focusTracker.None()
	}

	// Checks to see if any widget has been assigned the pressed key as its focus key
	if wtfApp.focusTracker.FocusOn(string(event.Rune())) {
		return nil
	}

	// If no specific widget has focus, then allow the key presses to fall through to the app
	if !wtfApp.focusTracker.IsFocused {
		switch string(event.Rune()) {
		case "q":
			wtfApp.Exit()
		case "/":
			return nil
		default:
		}
	}

	return event
}

func (wtfApp *WtfApp) refreshAllWidgets() {
	for _, widget := range wtfApp.widgets {
		go widget.Refresh()
	}
}

func (wtfApp *WtfApp) scheduleWidgets() {
	for _, widget := range wtfApp.widgets {
		go Schedule(widget)
	}
}

func (wtfApp *WtfApp) watchForConfigChanges() {
	// watch := watcher.New()
	var err error

	var watch *fsnotify.Watcher

	watch, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}

	// defer watch.Close()

	go watchLoop(watch, wtfApp)

	absPath, _ := utils.ExpandHomeDir(wtfApp.configFilePath)
	if err = watch.Add(absPath); err != nil {
		log.Fatalln(err)
	}

	// Watch config file for changes.
	absPath, _ = utils.ExpandHomeDir(wtfApp.configFilePath)
	if err = watch.Add(absPath); err != nil {
		log.Fatalln(err)
	}

	<-make(chan struct{})
}

func watchLoop(w *fsnotify.Watcher, wtfApp *WtfApp) {
	defer w.Close()

	for {
		select {
		case err, ok := <-w.Errors:
			if !ok || err != nil {

				return
			}
		case e, ok := <-w.Events:
			if !ok {

				return
			}

			if !e.Has(fsnotify.Write) && !e.Has(fsnotify.Create) && !e.Has(fsnotify.Rename) && !e.Has(fsnotify.Chmod) {
				continue
			}

			wtfApp.Stop()

			for {
				// wait for write to finish and file be available
				time.Sleep(100 * time.Millisecond)

				if _, err := os.Stat(e.Name); err == nil {
					break
				}
			}

			config := cfg.LoadWtfConfigFile(wtfApp.configFilePath)
			newApp := NewWtfApp(wtfApp.TViewApp, config, wtfApp.configFilePath)
			openURLUtil := utils.ToStrs(config.UList("wtf.openUrlUtil", []interface{}{}))
			utils.Init(config.UString("wtf.openFileUtil", "open"), openURLUtil)

			newApp.Start()
		}
	}
}
