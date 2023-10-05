package main

import (
	"fmt"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/edjubert/hyprland-ipc-go/hyprctl"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

func main() {
	app := gtk.NewApplication("hyprland.share.picker", gio.ApplicationFlagsNone)
	app.ConnectActivate(func() { activate(app) })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func createButton(label string, margin int, callback func()) *gtk.Button {
	newButton := gtk.NewButtonWithLabel(label)
	newButton.SetMarginStart(margin)
	newButton.SetMarginEnd(margin)
	newButton.SetMarginBottom(margin)
	newButton.SetMarginTop(margin)
	newButton.ConnectClicked(callback)
	return newButton
}

type Screen struct {
	Label  string
	Index  int
	X      int
	Y      int
	Width  int
	Height int
}

func getScreenList() []Screen {
	getter := hyprctl.Get{}
	monitors, err := getter.Monitors("-j")
	if err != nil {
		fmt.Println(err)
	}
	var screens []Screen
	for _, monitor := range monitors {
		screens = append(screens, Screen{
			Label:  monitor.Name,
			Index:  monitor.Id,
			X:      monitor.X,
			Y:      monitor.Y,
			Width:  monitor.Width,
			Height: monitor.Height,
		})
	}

	return screens
}

func createScreenPage() *gtk.Box {
	screenPage := gtk.NewBox(gtk.OrientationVertical, 0)

	for _, screen := range getScreenList() {
		callback := func() {
			fmt.Printf("screen:%s\n", screen.Label)
			os.Exit(0)
		}
		label := fmt.Sprintf("Screen %d at %d, %d (%dx%d) (%s)", screen.Index, screen.X, screen.Y, screen.Width, screen.Height, screen.Label)
		screenPage.Append(createButton(label, 6, callback))
	}

	return screenPage
}

func scrollableBox(box *gtk.Box) *gtk.ScrolledWindow {
	viewport := gtk.NewViewport(nil, nil)
	viewport.SetScrollToFocus(true)
	viewport.SetChild(box)

	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolledWindow.SetChild(viewport)
	scrolledWindow.SetPropagateNaturalHeight(true)

	return scrolledWindow
}

type Window struct {
	Address string
	Title   string
	X       int
	Y       int
	Width   int
	Height  int
}

func getWindowList() []Window {
	getter := hyprctl.Get{}
	clients, err := getter.Clients()
	if err != nil {
		fmt.Println(err)
	}
	var windows []Window
	for _, client := range clients {
		windows = append(windows, Window{Title: client.Title, Address: client.Address, X: client.At[0], Y: client.At[1], Width: client.Size[0], Height: client.Size[1]})
	}

	return windows
}

func createWindowPage() *gtk.Box {
	windowPage := gtk.NewBox(gtk.OrientationVertical, 0)

	for _, window := range getWindowList() {
		callback := func() {
			fmt.Printf("window:%s@%d,%d,%d,%d\n", window.Address, window.X, window.Y, window.Width, window.Height)
			os.Exit(0)
		}
		windowPage.Append(createButton(window.Title, 6, callback))
	}

	return windowPage
}

type Region struct {
	Screen string
	X      int
	Y      int
	Width  int
	Height int
}

func createDrawingRegion() Region {
	cmd := exec.Command("slurp")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("could not run slurp ", err)
	}

	findDigit := regexp.MustCompile(`(\d.*),(\d.*) (\d.*)x(\d.*)`)
	match := findDigit.FindSubmatch(out)
	if len(match) != 5 {
		fmt.Println("len: ", len(match))
		os.Exit(1)
	}

	x, _ := strconv.Atoi(string(match[1]))
	y, _ := strconv.Atoi(string(match[2]))
	width, _ := strconv.Atoi(string(match[3]))
	height, _ := strconv.Atoi(string(match[4]))

	getter := hyprctl.Get{}
	monitors, err := getter.Monitors("-j")
	if err != nil {
		fmt.Println(err)
	}

	region := Region{X: x, Y: y, Width: width, Height: height}
	for _, monitor := range monitors {
		if x > monitor.X && x < monitor.X+monitor.Width && y > monitor.Y && y < monitor.Y+monitor.Height {
			region.Screen = monitor.Name
		}
	}
	fmt.Println(region)

	return region
}

func createRegionPage() *gtk.Box {
	callback := func() {
		region := createDrawingRegion()
		fmt.Printf("region:%s@%d,%d,%d,%d", region.Screen, region.X, region.Y, region.Width, region.Height)
		os.Exit(0)
	}
	regionPage := gtk.NewBox(gtk.OrientationVertical, 0)
	regionPage.Append(createButton("Select region...", 6, callback))

	return regionPage
}

func createNotebook() *gtk.Notebook {
	notebook := gtk.NewNotebook()

	screenPage := scrollableBox(createScreenPage())
	windowPage := scrollableBox(createWindowPage())
	regionPage := createRegionPage()

	notebook.AppendPageMenu(screenPage, gtk.NewLabel("screen"), gtk.NewLabel("hey"))
	notebook.AppendPageMenu(windowPage, gtk.NewLabel("window"), gtk.NewLabel("hey2"))
	notebook.AppendPageMenu(regionPage, gtk.NewLabel("region"), gtk.NewLabel("hey2"))

	return notebook
}

func activate(app *gtk.Application) {
	window := gtk.NewApplicationWindow(app)

	notebook := createNotebook()

	window.SetChild(notebook)
	window.SetDefaultSize(400, 300)
	window.Show()
}
