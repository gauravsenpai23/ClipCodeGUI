# ClipCode - Clipboard Manager

ClipCode is a powerful and efficient clipboard manager application that allows you to manage and access your clipboard history with ease. It provides quick access to previously copied items through keyboard shortcuts and a user-friendly interface.

## Features

- **Clipboard History**: Automatically saves your clipboard history with thread-safe operations
- **Quick Access**: Use keyboard shortcuts to access clipboard items
- **Hotkey Support**: 
  - Alt + Q: Show clipboard history window
  - Alt + Shift + Q: Navigate backward through history
  - Alt + Q: Navigate forward through history
  - Shift + Alt + 1-9: Quick paste from history
- **Modern UI**: Clean and intuitive interface
- **Windows Support**: Optimized for Windows operating system
- **Thread-Safe Operations**: Protected clipboard operations to prevent data races
- **Efficient Monitoring**: Fast clipboard monitoring with optimized polling

## Installation

1. Download the latest release from the releases page
2. Extract the executable file
3. Run the application

## Usage

1. **Basic Navigation**:
   - Press and hold Alt
   - Press Q multiple times to navigate through clipboard history
   - Release Alt to paste the selected item

2. **Quick Access**:
   - Press Shift + Alt + (1-9) to quickly paste items from history
   - Number 1 corresponds to the most recent item

3. **Window Management**:
   - The window appears when you start navigating
   - It automatically hides when you release Alt
   - The window shows the current selected item
   - Shows "Clipboard is empty" when no items are available

## Technical Details

- Built with Go and Fyne UI framework
- Uses Windows API for keyboard monitoring
- Maintains a history of up to 100 items
- Thread-safe clipboard operations using mutex
- Efficient clipboard monitoring with 500ms polling interval
- Optimized hotkey registration for better performance

## Requirements

- Windows operating system
- Go 1.16 or higher (for development)

## Development

To build from source:

```bash
go build -o clipcode.exe
```

## Performance Considerations

- Thread-safe clipboard operations
- Optimized hotkey handling
- Efficient clipboard monitoring
- Memory-efficient history management

## License

This project is open source and available under the MIT License. 