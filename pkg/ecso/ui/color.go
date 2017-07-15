package ui

import "github.com/fatih/color"

type Color int

const (
	Blue = iota
	BlueBold
	Green
	GreenBold
	Red
	RedBold
	Bold
	Warn
)

var (
	bold     = color.New(color.Bold).SprintfFunc()
	warn     = color.New(color.FgRed).SprintfFunc()
	blue     = color.New(color.FgBlue).SprintfFunc()
	blueBold = color.New(color.FgBlue, color.Bold).SprintfFunc()

	green     = color.New(color.FgGreen).SprintfFunc()
	greenBold = color.New(color.FgGreen, color.Bold).SprintfFunc()

	red     = color.New(color.FgRed).SprintfFunc()
	redBold = color.New(color.FgRed, color.Bold).SprintfFunc()

	colors = map[Color]func(string, ...interface{}) string{
		Blue:      blue,
		BlueBold:  blueBold,
		Green:     green,
		GreenBold: greenBold,
		Red:       red,
		RedBold:   redBold,
		Bold:      bold,
		Warn:      warn,
	}
)
