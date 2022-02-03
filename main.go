package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Version
const Version = "Version 1.2 (c) 2022 Xavier Gandillot" // DEBUG turn on or off debugging on console.
const DEBUG = false

// theme to use
var th = material.NewTheme(gofont.Collection())
var separator = string(filepath.Separator) // file separator
var wdDir = mustString(os.Getwd())         // working directory

// various flags
var ignoreEmpty, ignoreGit = true, true
var processRunning bool

// globals
var startButton, quitButton, saveButton widget.Clickable
var dirEditor = widget.Editor{
	Alignment:  text.Start,
	SingleLine: true,
	Submit:     true,
	Mask:       0,
	InputHint:  0,
}
var results []string = []string{Version, Help}
var resultsMutex sync.Mutex
var resList = widget.List{
	List: layout.List{
		Axis:        layout.Vertical,
		ScrollToEnd: true,
	},
}
var ticker = time.NewTicker(500 * time.Millisecond) // force regular refresh of the window

var red = color.NRGBA{
	R: 100,
	G: 0,
	B: 0,
	A: 255,
}

const Help = `
Select the directory (relative to the current directory) in the upper editor.
You may use '..' to move up in the directory tree. but make sure you use the appropriate file separator ('/' for linux and '\' for windows).

Press Start to search for file with identical content.

Pressing Save will Save will save the content of the main window in a text file named resulst-xxxxxx.txt, where xxxx is a time stamp.

Press Quit or close the window to quit the program.`

// utility to check init errors
func mustString(s string, e error) string {
	if e != nil {
		panic(e)
	}
	return s
}

func main() {

	go runmainwindow()
	app.Main()
}

func runmainwindow() {
	w := app.NewWindow()

	err := mainloop(w)
	if err != nil {
		log.Fatal(err) // abort all
	}
	os.Exit(0)
}

// mainloop is window's main event loop
func mainloop(w *app.Window) error {

	var ops op.Ops

	for {

		select { // handle main loop

		case <-ticker.C: // if we have a tick waiting, invalidate and loop again
			w.Invalidate()
			if DEBUG {
				fmt.Println("Invalidated window")
			}

		case e := <-w.Events(): // process available events
			if DEBUG {
				fmt.Printf("Event : %T - %v\n", e, time.Now())
			}
			switch e := e.(type) {
			case system.DestroyEvent:
				fmt.Println("Exiting now ! (from main window closed)")
				return e.Err

			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				// uniform margin
				layout.UniformInset(unit.Dp(20)).Layout(
					gtx,
					func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis:    layout.Vertical,
							Spacing: layout.SpaceBetween,
						}.Layout(
							gtx,
							// We insert 3 rigid elements:

							layout.Rigid(drawTitle),
							layout.Rigid(drawDirEditorInset),
							layout.Rigid(drawButtons),
							layout.Flexed(1., drawResultsWithMargin), // occupy 100% of the remaining space
						)
					},
				)

				e.Frame(gtx.Ops)

			default:
				// ignore other events
			}
		}

		// handle DirEditor loop
		for _, evt := range dirEditor.Events() {
			switch evt.(type) {
			case widget.ChangeEvent:
				// Do something when Change the text
			case widget.SelectEvent:
				// Do something when select the text
			case widget.SubmitEvent:
				go doProcess()
			}
		}
	}
}

func drawButtons(gtx layout.Context) layout.Dimensions {

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceStart,
	}.Layout(
		gtx,

		layout.Rigid(drawStartButton),
		layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
		layout.Rigid(drawSaveButton),
		layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
		layout.Rigid(drawQuitButton),
	)
}

func drawQuitButton(gtx layout.Context) layout.Dimensions {
	btn := material.Button(th, &quitButton, " Quit ")
	btn.Background = red

	if quitButton.Clicked() {
		fmt.Println("Exiting now ! (from quit button)")
		os.Exit(0)
	}
	return btn.Layout(gtx)
}

func drawStartButton(gtx layout.Context) layout.Dimensions {

	if processRunning {

		btn := material.Button(th, &startButton, "Processing, please wait ...")
		return btn.Layout(gtx.Disabled())

	} else {

		btn := material.Button(th, &startButton, " Start ")

		if startButton.Clicked() {
			// drain multiple clicks
			startButton.Clicks()
			if DEBUG {
				fmt.Println("*** start button was clicked !")
			}
			go doProcess()
		}
		return btn.Layout(gtx)
	}
}

func drawSaveButton(gtx layout.Context) layout.Dimensions {
	btn := material.Button(th, &saveButton, " Save ")
	if !processRunning && saveButton.Clicked() {
		// drain multiple clicks
		saveButton.Clicks()
		if DEBUG {
			fmt.Println("*** saving results to results.txt !")
		}
		rr := []byte(strings.Join(results, "\n"))
		fn := fmt.Sprintf("results-%s.txt", time.Now().Format("2006-01-02-150405"))
		err := ioutil.WriteFile(fn, rr, 0644)
		if err != nil {
			results = append(results, err.Error())
		}
	}
	if processRunning {
		return btn.Layout(gtx.Disabled())
	} else {
		return btn.Layout(gtx)
	}

}

func doProcess() {
	if processRunning { // don't run twice at the same time. This is not 100% thread safe, but should be ok.
		return
	} else {
		processRunning = true
	}
	err := Process(wdDir, dirEditor.Text())
	if err != nil {
		resultsMutex.Lock()
		results = append(results, "An error occured :", err.Error())
		resultsMutex.Unlock()
	}
	processRunning = false
}

func drawTitle(gtx layout.Context) layout.Dimensions {
	title := material.H3(th, "Duplicates finder")

	title.Color = red
	title.Alignment = text.Middle
	return title.Layout(gtx)
}

func drawDirEditor(gtx layout.Context) layout.Dimensions {
	edt := material.Editor(th, &dirEditor, "directory to search")
	edt.Color = red
	edt.Font.Weight = 700

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceEnd,
	}.Layout(
		gtx,
		layout.Rigid(material.Label(th, unit.Dp(16), wdDir+separator).Layout),
		layout.Flexed(1., edt.Layout),
	)
}

func drawDirEditorInset(gtx layout.Context) layout.Dimensions {
	return layout.Inset{
		Top:    unit.Dp(30),
		Bottom: unit.Dp(30),
	}.Layout(gtx, drawDirEditor)
}

func drawResults(gtx layout.Context) layout.Dimensions {
	/* if DEBUG {
		 fmt.Println("displaying results : ", results)
	} */
	res := material.List(th, &resList)
	return res.Layout(gtx,
		len(results),
		func(gtx layout.Context, index int) layout.Dimensions {
			r := ""
			if index < len(results) {
				r = results[index]
			}
			parag := material.Label(th, unit.Dp(15), r)
			return parag.Layout(gtx)
		},
	)
}

func drawResultsWithMargin(gtx layout.Context) layout.Dimensions {
	resultsMutex.Lock()
	dr := layout.UniformInset(unit.Dp(10)).Layout(gtx, drawResults)
	resultsMutex.Unlock()
	return dr
}
