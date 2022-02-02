package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
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

// DEBUG turn on or off debugging on console.
const DEBUG = false

// theme to use
var th = material.NewTheme(gofont.Collection())

// startButton is a clickable widget
var startButton widget.Clickable
var dirEditor widget.Editor
var results []string = make([]string, 0)
var resultsMutex sync.Mutex
var resList widget.List
var wdDir string
var processRunning bool
var ticker = time.NewTicker(500 * time.Millisecond) // force regular refresh of the window

func init() {

	dirEditor.SingleLine = true
	dirEditor.Submit = true
	dirEditor.Alignment = text.Start

	resList.Axis = layout.Vertical
	resList.ScrollToEnd = true

	var err error
	wdDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}
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
		select {

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
				if DEBUG {
					fmt.Println("Exiting now !")
				}
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
							layout.Rigid(drawDirEditor),
							layout.Flexed(1., drawResultsWithMargin), // occupy 100% of the remaining space
							layout.Rigid(drawStartButton),
						)
					},
				)

				e.Frame(gtx.Ops)
			default:
				// ignore other events
			}
		}
	}
}

func drawStartButton(gtx layout.Context) layout.Dimensions {

	if processRunning {

		btn := material.Button(th, &startButton, "Please wait ...")
		return btn.Layout(gtx.Disabled())

	} else {

		btn := material.Button(th, &startButton, "Start")

		if startButton.Clicked() {
			if DEBUG {
				fmt.Println("*** button was clicked !")
			}
			processRunning = true
			go func() {
				err := Process(wdDir, dirEditor.Text())
				if err != nil {
					resultsMutex.Lock()
					results = append(results, "An error occured :", err.Error())
					resultsMutex.Unlock()
				}
				processRunning = false
			}()
		}
		return btn.Layout(gtx)
	}
}

func drawTitle(gtx layout.Context) layout.Dimensions {
	title := material.H3(th, "Duplicates finder")
	maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
	title.Color = maroon
	title.Alignment = text.Middle
	return title.Layout(gtx)
}

func drawDirEditor(gtx layout.Context) layout.Dimensions {
	edt := material.Editor(th, &dirEditor, "directory to search")

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceBetween,
	}.Layout(gtx,

		layout.Rigid(material.Label(th, unit.Dp(16), wdDir).Layout),
		layout.Rigid(edt.Layout),
	)

}

func drawResults(gtx layout.Context) layout.Dimensions {
	if DEBUG {
		fmt.Println("displaying results : ", results)
	}
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
