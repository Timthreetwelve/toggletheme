package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"fyne.io/systray"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var iconData []byte
var mToggleTheme *systray.MenuItem
var user32 = syscall.NewLazyDLL("user32.dll")
var procMessageBoxW = user32.NewProc("MessageBoxW")

const personalizePath = `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`
const newline = "\r\n"

func MessageBox(title, text string, style uint) int {
	ret, _, _ := procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(style),
	)
	return int(ret)
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("Go Tray App")
	systray.SetTooltip("Toggle Theme")

	// Menu items
	mToggleTheme = systray.AddMenuItem(getToggleLabel(), "Switch between light and dark mode")
	systray.AddSeparator()
	mAbout := systray.AddMenuItem("About", "About Toggle Theme")
	mQuit := systray.AddMenuItem("Quit", "Exit the application")

	// Start registry watcher in background
	go watchThemeChanges()

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mToggleTheme.ClickedCh:
				if err := toggleTheme(); err != nil {
					log.Printf("%s", "Failed to toggle theme: "+err.Error())
				} else {
					mToggleTheme.SetTitle(getToggleLabel())
				}
			case <-mAbout.ClickedCh:
				var msg = fmt.Sprintf("Toggles Windows theme between light and dark mode%s%s© 2026 Tim Kennedy",
					newline, newline)
				MessageBox("About Toggle Theme", msg, 0)
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	// Cleanup if needed
	fmt.Println("Done.")
	log.Println("Exiting.")
}

func loadIcon(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to load icon: %v", err)
	}
	return data
}

func getCurrentTheme() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, personalizePath, registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return false, err
	}
	return val == 1, nil
}

func getToggleLabel() string {
	isLight, err := getCurrentTheme()
	if err != nil {
		return "Toggle Light/Dark Mode"
	}
	if isLight {
		log.Println("Switched to Light Mode")
		return "Switch to Dark Mode"
	}
	log.Println("Switched to Dark Mode")
	return "Switch to Light Mode"
}

func toggleTheme() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, personalizePath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	current, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return err
	}

	var newVal uint64
	if current == 1 {
		newVal = 0 // Dark mode
	} else {
		newVal = 1 // Light mode
	}

	if err := k.SetDWordValue("AppsUseLightTheme", uint32(newVal)); err != nil {
		return err
	}
	if err := k.SetDWordValue("SystemUsesLightTheme", uint32(newVal)); err != nil {
		return err
	}

	return nil
}

// watchThemeChanges listens for registry changes and updates the menu label
func watchThemeChanges() {
	k, err := registry.OpenKey(registry.CURRENT_USER, personalizePath, registry.NOTIFY)
	if err != nil {
		log.Printf("Failed to open registry key for watching: %v", err)
		return
	}
	defer k.Close()

	h := windows.Handle(k)
	for {
		// Wait for registry change notification
		err := windows.RegNotifyChangeKeyValue(
			h,
			false,
			windows.REG_NOTIFY_CHANGE_LAST_SET,
			0,
			false,
		)
		if err != nil {
			log.Printf("Registry watch error: %v", err)
			return
		}

		// Update menu label after change
		mToggleTheme.SetTitle(getToggleLabel())
	}
}

func main() {
	// Load icon (must be .ico for Windows)
	iconData = loadIcon("dark-theme.ico")

	f, _ := os.CreateTemp("", "toggletheme-*.log")
	log.SetOutput(f)
	log.Println("Toggle Theme is starting up.")
	fmt.Println("Toggle Theme is starting. Check the system tray.")

	// Handle Ctrl+C / SIGTERM gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		systray.Quit()
	}()

	// Run systray
	systray.Run(onReady, onExit)
}
