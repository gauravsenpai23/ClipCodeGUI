package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"
	hook "github.com/robotn/gohook"
	"golang.org/x/sys/windows"
)

// Windows API constants
const (
	VK_LMENU       = 0xA4 // Left Alt key
	VK_RMENU       = 0xA5 // Right Alt key
	VK_Q           = 0x51 // Q key
	VK_LSHIFT      = 0xA0 // Left Shift key
	VK_RSHIFT      = 0xA1 // Right Shift key
	maxHistorySize = 100
	pollInterval   = 2000 * time.Millisecond
)

var (
	clipboardHistory []string
	lastClipboard    string
	currentString    string
)

// Check if a key is pressed using Windows API
func isKeyPressed(key int) bool {
	state, _, _ := windows.NewLazySystemDLL("user32.dll").NewProc("GetAsyncKeyState").Call(uintptr(key))
	return (state & 0x8000) != 0
}

func main() {
	go monitorClipboard()
	go registerHotkeys()

	// Create Fyne application
	fyneApp := app.New()
	window := fyneApp.NewWindow("ClipCode")

	// Create the label to display array data
	dataLabel := widget.NewLabel("Press alt to hide this window")

	// Set up the content
	content := container.NewCenter(dataLabel)
	window.SetContent(content)

	// Hide window initially and track visibility
	window.Resize(fyne.NewSize(400, 200))
	window.Hide()
	isWindowVisible := false

	// Variables to track state
	altPressed := false
	qPressed := false
	shiftPressed := false
	qCount := 0
	currentIndex := 0

	// Create channel for graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Main loop for key monitoring
	go func() {
		for {
			// Check key states
			altDown := isKeyPressed(VK_LMENU) || isKeyPressed(VK_RMENU)
			qDown := isKeyPressed(VK_Q)
			shiftDown := isKeyPressed(VK_LSHIFT) || isKeyPressed(VK_RSHIFT)

			// Handle Alt key press/release
			if altDown && !altPressed {
				// Alt key just pressed
				altPressed = true
				qCount = 0
				fmt.Println("Alt pressed - tracking 'q' presses")
			} else if !altDown && altPressed {
				// Alt key just released
				fmt.Printf("Alt released - 'q' was pressed %d times while Alt was held\n", qCount)
				// Hide the window when Alt is released
				clipboard.WriteAll(currentString)
				window.Hide()
				isWindowVisible = false
				altPressed = false
			}

			// Update shift state
			shiftPressed = shiftDown

			// Handle 'q' key press/release while Alt is pressed
			if altPressed {
				if qDown && !qPressed {
					// 'q' key just pressed
					qCount++
					fmt.Printf("'q' pressed while Alt is held (count: %d)\n", qCount)

					// Update current index based on q count (circular)
					if shiftPressed {
						// Go backward when shift is pressed
						currentIndex = (currentIndex - 1 + len(clipboardHistory)) % len(clipboardHistory)
					} else {
						// Go forward normally
						currentIndex = (qCount - 1) % len(clipboardHistory)
					}

					// Update the display
					dataLabel.SetText(clipboardHistory[currentIndex])
					currentString = clipboardHistory[currentIndex]

					// Show the window if it's not visible
					if !isWindowVisible {
						window.Show()
						isWindowVisible = true
					}
				}
			}

			// Update previous state
			qPressed = qDown

			// Sleep to avoid excessive CPU usage
			time.Sleep(20 * time.Millisecond)
		}
	}()

	// Wait for interrupt signal or window close
	go func() {
		<-sig
		fyneApp.Quit()
	}()

	// When window is closed, update visibility state
	window.SetOnClosed(func() {
		isWindowVisible = false
	})

	// Start the application
	window.ShowAndRun()

}

func copyFromControlC() {
	curr, err := clipboard.ReadAll()
	if err != nil {
		fmt.Printf("Error reading clipboard: %v\n", err)

	}
	if lastClipboard != curr && curr != "" {
		clipboardHistory = append([]string{curr}, clipboardHistory...)
		if len(clipboardHistory) > maxHistorySize {
			clipboardHistory = clipboardHistory[:maxHistorySize]
		}
		lastClipboard = curr
	}
}

func monitorClipboard() {
	for {
		current, err := clipboard.ReadAll()
		if err != nil {
			fmt.Printf("Error reading clipboard: %v\n", err)
			continue
		}

		if current != lastClipboard && current != "" {
			// Add new entry to beginning of history
			clipboardHistory = append([]string{current}, clipboardHistory...)

			// Trim history if it exceeds maximum size
			if len(clipboardHistory) > maxHistorySize {
				clipboardHistory = clipboardHistory[:maxHistorySize]
			}

			lastClipboard = current
		}

		time.Sleep(pollInterval)

	}
}

func registerHotkeys() {
	fmt.Println("--- Press Shift+alt+1-9 to paste from history ---")

	for i := 1; i <= 9; i++ {
		key := fmt.Sprintf("%d", i)
		hook.Register(hook.KeyDown, []string{key, "shift", "alt"}, func(e hook.Event) {
			keyStr := string(e.Rawcode)
			pasteFromHistory(keyStr)
		})
	}
	hook.Register(hook.KeyDown, []string{"ctrl", "c"}, func(e hook.Event) {
		copyFromControlC()
	})

	s := hook.Start()
	<-hook.Process(s)
}

func pasteFromHistory(keyStr string) {
	if len(clipboardHistory) == 0 {
		fmt.Println("Clipboard history is empty")
		return
	}

	if keyStr >= "1" && keyStr <= "9" {
		index := int(keyStr[0] - '1')
		if index < len(clipboardHistory) {
			err := clipboard.WriteAll(clipboardHistory[index])
			if err != nil {
				fmt.Printf("Error writing to clipboard: %v\n", err)
				return
			}
			lastClipboard = clipboardHistory[index]
			fmt.Printf("Pasted item %d from history\n", index+1)
		}
	}
}
