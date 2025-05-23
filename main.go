package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
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

const (
	VK_LMENU       = 0xA4 // Left Alt key
	VK_RMENU       = 0xA5 // Right Alt key
	VK_Q           = 0x51 // Q key
	VK_LSHIFT      = 0xA0 // Left Shift key
	VK_RSHIFT      = 0xA1 // Right Shift key
	maxHistorySize = 100
	pollInterval   = 500 * time.Millisecond // Reduced from 2000ms to 500ms
)

var (
	clipboardHistory []string
	lastClipboard    string
	currentString    string
	historyMutex     sync.RWMutex
)

func isKeyPressed(key int) bool {
	state, _, _ := windows.NewLazySystemDLL("user32.dll").NewProc("GetAsyncKeyState").Call(uintptr(key))
	return (state & 0x8000) != 0
}

func main() {
	go monitorClipboard()
	go registerHotkeys()

	fyneApp := app.New()
	window := fyneApp.NewWindow("ClipCode")

	dataLabel := widget.NewLabel("Press alt to hide this window\nPress and hold Alt\nPress Q multiple times to navigate through clipboard history\nRelease Alt to paste the selected item")

	content := container.NewCenter(dataLabel)
	window.SetContent(content)

	window.Resize(fyne.NewSize(400, 200))
	window.Hide()
	isWindowVisible := false

	altPressed := false
	qPressed := false
	shiftPressed := false
	qCount := 0
	currentIndex := 0

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			altDown := isKeyPressed(VK_LMENU) || isKeyPressed(VK_RMENU)
			qDown := isKeyPressed(VK_Q)
			shiftDown := isKeyPressed(VK_LSHIFT) || isKeyPressed(VK_RSHIFT)

			if altDown && !altPressed {
				altPressed = true
				qCount = 0
				fmt.Println("Alt pressed - tracking 'q' presses")
			} else if !altDown && altPressed {
				fmt.Printf("Alt released - 'q' was pressed %d times while Alt was held\n", qCount)
				if currentString != lastClipboard {
					clipboard.WriteAll(currentString)
				}

				window.Hide()
				isWindowVisible = false
				altPressed = false
			}

			shiftPressed = shiftDown

			if altPressed {
				if qDown && !qPressed {
					qCount++
					fmt.Printf("'q' pressed while Alt is held (count: %d)\n", qCount)
					if len(clipboardHistory) == 0 {
						dataLabel.SetText("Clipboard is empty")
						currentString = ""
						time.Sleep(1000 * time.Millisecond)
					} else if shiftPressed {
						currentIndex = (currentIndex - 1 + len(clipboardHistory)) % len(clipboardHistory)
						dataLabel.SetText(clipboardHistory[currentIndex])
						currentString = clipboardHistory[currentIndex]
					} else {
						currentIndex = (qCount - 1) % len(clipboardHistory)
						dataLabel.SetText(clipboardHistory[currentIndex])
						currentString = clipboardHistory[currentIndex]
					}

					if !isWindowVisible {
						window.Show()
						isWindowVisible = true
					}
				}
			}

			qPressed = qDown

			time.Sleep(20 * time.Millisecond)
		}
	}()

	go func() {
		<-sig
		fyneApp.Quit()
	}()

	window.SetOnClosed(func() {
		isWindowVisible = false
	})

	window.ShowAndRun()

}

func monitorClipboard() {
	for {
		current, err := clipboard.ReadAll()
		if err != nil {
			fmt.Printf("Error reading clipboard: %v\n", err)
			time.Sleep(pollInterval)
			continue
		}

		historyMutex.Lock()
		if current != lastClipboard && current != "" {
			clipboardHistory = append([]string{current}, clipboardHistory...)
			if len(clipboardHistory) > maxHistorySize {
				clipboardHistory = clipboardHistory[:maxHistorySize]
			}
			lastClipboard = current
		}
		historyMutex.Unlock()

		time.Sleep(pollInterval)
	}
}

func registerHotkeys() {
	fmt.Println("--- Press Shift+alt+1-9 to paste from history ---")

	// Pre-create the hotkey handlers to avoid creating closures in the loop
	hotkeyHandlers := make(map[string]func(hook.Event))
	for i := 1; i <= 9; i++ {
		key := fmt.Sprintf("%d", i)
		hotkeyHandlers[key] = func(e hook.Event) {
			keyStr := string(e.Rawcode)
			pasteFromHistory(keyStr)
		}
	}

	for i := 1; i <= 9; i++ {
		key := fmt.Sprintf("%d", i)
		hook.Register(hook.KeyDown, []string{key, "shift", "alt"}, hotkeyHandlers[key])
	}

	s := hook.Start()
	<-hook.Process(s)
}

func pasteFromHistory(keyStr string) {
	historyMutex.RLock()
	defer historyMutex.RUnlock()

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
