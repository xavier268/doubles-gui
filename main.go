package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
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

// theme to use
var th = material.NewTheme(gofont.Collection())

// startButton is a clickable widget
var startButton widget.Clickable
var dirEditor widget.Editor
var results []string = make([]string, 0)
var resList widget.List

func init() {

	dirEditor.SingleLine = true
	dirEditor.Submit = true
	dirEditor.Alignment = text.Start

	resList.Axis = layout.Vertical
	resList.ScrollToEnd = true
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

		e := <-w.Events()

		switch e := e.(type) {
		case system.DestroyEvent:
			fmt.Println("Exiting now !")
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			// uniform margin
			layout.UniformInset(unit.Dp(20)).Layout(
				gtx,
				func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						// Vertical alignment, from top to bottom
						Axis: layout.Vertical,
						// Empty space is left at the start, i.e. at the top
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
			fmt.Printf("%T\t%v\n", e, e)
		}
	}
}

func drawStartButton(gtx layout.Context) layout.Dimensions {
	btn := material.Button(th, &startButton, "Start")
	if startButton.Clicked() {
		fmt.Println("*** button was clicked !")
		results = append(results, time.Now().String()+"blabla bla")
	}
	return btn.Layout(gtx)
}

func drawTitle(gtx layout.Context) layout.Dimensions {
	title := material.H3(th, "Double ckecker")
	maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
	title.Color = maroon
	title.Alignment = text.Middle
	return title.Layout(gtx)
}

func drawDirEditor(gtx layout.Context) layout.Dimensions {
	edt := material.Editor(th, &dirEditor, "directory to search")
	return edt.Layout(gtx)
}

func drawResults(gtx layout.Context) layout.Dimensions {
	fmt.Println("displaying results : ", results)
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
	return layout.UniformInset(unit.Dp(10)).Layout(gtx, drawResults)
}
