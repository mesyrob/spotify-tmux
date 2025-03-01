// ui/ui.go
package ui

import (
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/yourusername/spotify-tmux/player"
)

// PlayerController defines the interface for player control
type PlayerController interface {
	Play() error
	Pause() error
	Next() error
	Previous() error
	PlayPause() error
	GetCurrentlyPlaying() (*player.CurrentlyPlaying, error)
	FormatTrackInfo() (string, error)
}

// UI handles the terminal user interface
type UI struct {
	app       *tview.Application
	player    PlayerController
	infoText  *tview.TextView
	stopChan  chan struct{}
	updateInt time.Duration
}

// NewUI creates a new terminal UI
func NewUI(player PlayerController) *UI {
	app := tview.NewApplication()
	infoText := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	
	return &UI{
		app:       app,
		player:    player,
		infoText:  infoText,
		stopChan:  make(chan struct{}),
		updateInt: 1 * time.Second,
	}
}

// Start starts the UI
func (u *UI) Start() {
	// Create main layout
	grid := tview.NewGrid().
		SetRows(1, 1, 1).
		SetColumns(0)
	
	// Create buttons
	prevButton := tview.NewButton("◀ Previous").
		SetSelectedFunc(func() {
			if err := u.player.Previous(); err != nil {
				u.showError(err)
			}
		})
	
	playButton := tview.NewButton("▶ Play/Pause").
		SetSelectedFunc(func() {
			if err := u.player.PlayPause(); err != nil {
				u.showError(err)
			}
		})
	
	nextButton := tview.NewButton("Next ▶").
		SetSelectedFunc(func() {
			if err := u.player.Next(); err != nil {
				u.showError(err)
			}
		})
	
	// Create button bar
	buttonBar := tview.NewFlex().
		AddItem(prevButton, 0, 1, false).
		AddItem(playButton, 0, 1, false).
		AddItem(nextButton, 0, 1, false)
	
	// Add elements to grid
	grid.AddItem(u.infoText, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(buttonBar, 1, 0, 1, 1, 0, 0, true)
	grid.AddItem(tview.NewTextView().
		SetText("Shortcuts: p = play/pause, n = next, b = previous, q = quit").
		SetTextAlign(tview.AlignCenter), 2, 0, 1, 1, 0, 0, false)
	
	// Set up keyboard shortcuts
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			u.app.Stop()
			return nil
		case 'p':
			if err := u.player.PlayPause(); err != nil {
				u.showError(err)
			}
			return nil
		case 'n':
			if err := u.player.Next(); err != nil {
				u.showError(err)
			}
			return nil
		case 'b':
			if err := u.player.Previous(); err != nil {
				u.showError(err)
			}
			return nil
		}
		return event
	})
	
	// Start auto-update
	go u.updateLoop()
	
	// Set root and start
	if err := u.app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}

// Stop stops the UI
func (u *UI) Stop() {
	close(u.stopChan)
	u.app.Stop()
}

// updateLoop periodically updates the track info
func (u *UI) updateLoop() {
	ticker := time.NewTicker(u.updateInt)
	defer ticker.Stop()
	
	// Update immediately on start
	u.updateTrackInfo()
	
	for {
		select {
		case <-ticker.C:
			u.updateTrackInfo()
		case <-u.stopChan:
			return
		}
	}
}

// updateTrackInfo updates the track information display
func (u *UI) updateTrackInfo() {
	info, err := u.player.FormatTrackInfo()
	if err != nil {
		u.showError(err)
		return
	}
	
	u.app.QueueUpdateDraw(func() {
		u.infoText.SetText(fmt.Sprintf("[green]%s[white]", info))
	})
}

// showError displays an error message
func (u *UI) showError(err error) {
	u.app.QueueUpdateDraw(func() {
		u.infoText.SetText(fmt.Sprintf("[red]Error: %v[white]", err))
	})
}