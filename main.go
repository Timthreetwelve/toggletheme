package main

import (
	_ "embed"
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

//go:embed dark-theme.ico
var iconData []byte

var mToggleTheme *systray.MenuItem
var user32 = syscall.NewLazyDLL("user32.dll")
var procMessageBoxW = user32.NewProc("MessageBoxW")

const personalizePath = `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`
const newline = "\r\n"

func MessageBox(title, text string, style uint) int {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	textPtr, _ := syscall.UTF16PtrFromString(text)
	ret, _, _ := procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(style),
	)
	return int(ret)
}

func onReady() {
	systray.SetIcon(iconData)
	//systray.SetTitle("Go Tray App")
	systray.SetTooltip("ToggleTheme")

	// Menu items
	mToggleTheme = systray.AddMenuItem(getToggleLabel(), "Switch between light and dark mode")
	systray.AddSeparator()
	mAbout := systray.AddMenuItem("About", "About ToggleTheme")
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
				curMode, _ := getCurrentTheme()
				if curMode {
					log.Println("Switched to light mode.")
					fmt.Println("Switched to light mode.")
				} else {
					log.Println("Switched to dark mode.")
					fmt.Println("Switched to dark mode.")
				}
			case <-mAbout.ClickedCh:
				var msg = fmt.Sprintf("Toggles Windows theme between light and dark mode%s%s"+
					"https://github.com/Timthreetwelve/toggletheme%s%sCreated by Tim Kennedy",
					newline, newline, newline, newline)
				MessageBox("About ToggleTheme", msg, 0)
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
	log.Printf("ToggleTheme is shutting down.")
	log.Printf("")
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
		return "Switch to Dark Mode"
	}
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
	temp := os.TempDir()
	logFile := temp + "\\toggletheme.log"
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("error opening log file: %v", err)
	}
	log.SetOutput(file)

	execPath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}
	log.Printf("ToggleTheme is starting up from %s.", execPath)

	// Handle Ctrl+C / SIGTERM gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		systray.Quit()
	}()

	// Run systray
	fmt.Println("ToggleTheme is running in the system tray. ")
	systray.Run(onReady, onExit)
}
