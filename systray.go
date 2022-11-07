package main

import (
	"fmt"
	"os/exec"
	"runtime"

	"fyne.io/systray"
)

func openInBrowser() error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}

	url := fmt.Sprintf("http://localhost:%d", Port)
	fmt.Printf("Opening %s in browser\n", url)

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func systrayOnReady() {
	systray.SetTemplateIcon(faviconpng, faviconpng)

	tooltip := fmt.Sprintf("Bloghead is live at http://localhost:%d", Port)
	systray.SetTitle(tooltip)
	systray.SetTooltip(tooltip)

	header := systray.AddMenuItem(tooltip, tooltip)
	header.Disable()
	systray.AddSeparator()

	mOpen := systray.AddMenuItem("Open Web UI", "Open Web UI")
	go func() {
		<-mOpen.ClickedCh
		if err := openInBrowser(); err != nil {
			panic(err)
		}
	}()

	mExit := systray.AddMenuItem("Exit", "Exit")
	go func() {
		<-mExit.ClickedCh
		fmt.Println("Exiting")
		systray.Quit()
		fmt.Println("Finished exiting")
	}()
}
