package main

import (
	"fmt"
	"os/exec"
	"runtime"

	"fyne.io/systray"
)

func openInBrowser(url string) error {
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

	fmt.Printf("Opening %s in browser\n", url)
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func systrayOnReady(url string) {
	systray.SetTemplateIcon(faviconpng, faviconpng)

	tooltip := fmt.Sprintf("Bloghead is live at %s", url)
	systray.SetTitle(tooltip)
	systray.SetTooltip(tooltip)

	header := systray.AddMenuItem(tooltip, tooltip)
	header.Disable()
	systray.AddSeparator()

	mOpen := systray.AddMenuItem("Open Web UI", "Open Web UI")
	go func() {
		for range mOpen.ClickedCh {
			if err := openInBrowser(url); err != nil {
				panic(err)
			}
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
