# ToggleTheme

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26.4-00ADD8?logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows-0078D6?logo=windows)](https://www.microsoft.com/windows)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/Timthreetwelve/toggletheme)](https://github.com/Timthreetwelve/toggletheme/releases/latest)
[![GitHub Release Date](https://img.shields.io/github/release-date/timthreetwelve/toggletheme)](https://github.com/Timthreetwelve/toggletheme/releases/latest)
[![GitHub commits since latest release (by date)](https://img.shields.io/github/commits-since/timthreetwelve/toggletheme/latest)](https://github.com/Timthreetwelve/toggletheme/commits/main)
[![GitHub last commit](https://img.shields.io/github/last-commit/timthreetwelve/toggletheme)](https://github.com/Timthreetwelve/toggletheme/commits/main)

ToggleTheme is a lightweight Windows system tray app that switches your system theme between Light and Dark mode. Nothing fancy, just switches themes. 

It runs in the notification area, lets you toggle themes with a click, and keeps its menu label in sync when theme settings change externally.

I wrote it to quickly switch between themes while working on a new theme for my .NET apps. Maybe you could find a use for it too.

## Features

- Runs as a tray app (no main window)
- Left-click tray icon to toggle Light/Dark mode
- Right-click menu with:
  - `Switch to Dark Mode` / `Switch to Light Mode`
  - `About`
  - `Quit`
- Watches Windows theme registry keys and updates menu text automatically
- Writes logs to `%TEMP%\\toggletheme.log`

## How It Works

ToggleTheme updates these Windows registry values under:

`HKEY_CURRENT_USER\\Software\\Microsoft\\Windows\\CurrentVersion\\Themes\\Personalize`

- `AppsUseLightTheme`
- `SystemUsesLightTheme`

When either value changes, the app refreshes the tray menu label to reflect the current mode.

## Requirements

- Windows
- Go `1.26.4` (from `go.mod`)

## Build

From the project root:

```bash
go mod tidy
go build -ldflags "-s -w -H=windowsgui" -trimpath -o toggletheme.exe
```

## Run

```bash
./toggletheme.exe
```

After launch, look for the tray icon in the Windows notification area.

## Regenerate Windows Resources

If you update `winres/winres.json`, regenerate resources with:

```bash
go-winres make
```

## Project Structure

```text
.
|- main.go
|- dark-theme.ico
|- go.mod
|- go.sum
|- go.work
|- winres/
|  |- winres.json
|- rsrc_windows_386.syso
|- rsrc_windows_amd64.syso
```

## Troubleshooting

- **No tray icon appears**: Ensure the app is running and check hidden icons in the taskbar overflow area.
- **Theme does not switch**: Verify your Windows personalization settings are not restricted by policy.
- **Need diagnostics**: Open `%TEMP%\\toggletheme.log` for startup and runtime logs.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
