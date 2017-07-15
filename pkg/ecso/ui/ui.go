package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/color"
)

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

func AskString(prompt, def string, validate func(string) error) (string, error) {
	str := ""
	err := AskStringVar(&str, prompt, def, validate)

	return str, err
}

func AskStringIfEmptyVar(dst *string, prompt, def string, validate func(string) error) error {
	if len(*dst) == 0 {
		return AskStringVar(dst, prompt, def, validate)
	}
	return nil
}

func AskStringVar(dst *string, prompt, def string, validate func(string) error) error {
	reader := bufio.NewReader(os.Stdin)

	if len(def) > 0 {
		prompt = fmt.Sprintf("%s (%s)", prompt, def)
	}

	fmt.Printf("%s\n", bold("%s: ", prompt))

	for {
		fmt.Print(" > ")

		str, err := reader.ReadString('\n')

		if err != nil {
			return err
		}

		str = str[:len(str)-1]

		if len(str) == 0 {
			str = def
		}

		if err := validate(str); err != nil {
			fmt.Printf("   %s\n", warn(err.Error()))
		} else {
			*dst = str

			return nil
		}
	}
}

func AskInt(prompt string, def int, validate func(int) error) (int, error) {
	i := 0
	err := AskIntVar(&i, prompt, def, validate)

	return i, err
}

func AskIntIfEmptyVar(dst *int, prompt string, def int, validate func(int) error) error {
	if *dst == 0 {
		return AskIntVar(dst, prompt, def, validate)
	}
	return nil
}

func AskIntVar(dst *int, prompt string, def int, validate func(int) error) error {
	reader := bufio.NewReader(os.Stdin)

	if def != 0 {
		prompt = fmt.Sprintf("%s (%d)", prompt, def)
	}

	fmt.Printf("%s\n", bold("%s: ", prompt))

	for {
		fmt.Print(" > ")

		str, err := reader.ReadString('\n')

		if err != nil {
			return err
		}

		str = str[:len(str)-1]

		if len(str) == 0 {
			str = strconv.Itoa(def)
		}

		i, err := strconv.Atoi(str)

		if err != nil {
			fmt.Printf("   %s\n", warn("Please enter a number"))
		} else {
			if err := validate(i); err != nil {
				fmt.Printf(" %s\n", warn(err.Error()))
			} else {
				*dst = i

				return nil
			}
		}
	}
}

func Choice(prompt string, choices []string) (int, error) {
	i := 0
	err := ChoiceVar(&i, prompt, choices)

	return i, err
}

func ChoiceVar(dst *int, prompt string, choices []string) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s\n", bold("%s: ", prompt))

	for i, choice := range choices {
		fmt.Printf(" %d) %s\n", i+1, choice)
	}

	for {
		fmt.Print("  > ")

		str, err := reader.ReadString('\n')

		if err != nil {
			return err
		}

		str = str[:len(str)-1]

		i, err := strconv.Atoi(str)

		if err != nil || i < 1 || i > len(choices) {
			fmt.Printf("   %s\n", warn("Please enter a number between %d and %d", 1, len(choices)))
		} else {
			*dst = i - 1

			return nil
		}
	}
}

// type TableDataProvider interface {
// 	TableHeader() []string
// 	TableRows() []map[string]string
// }

// func PrintTable(w io.Writer, data TableDataProvider) {
// 	headers := data.TableHeader()
// 	rows := data.TableRows()
// 	format := ""

// 	for _, h := range headers {
// 		l := len(h)

// 		for _, r := range rows {
// 			if v, ok := r[h]; ok && len(v) > l {
// 				l = len(v)
// 			}
// 		}

// 		format = format + fmt.Sprintf("%%-%ds  ", l)
// 	}

// 	format = format + "\n"

// 	headerRow := make([]interface{}, len(headers))

// 	for i, h := range headers {
// 		headerRow[i] = h
// 	}

// 	fmt.Fprintf(w, format, headerRow...)

// 	for _, row := range rows {
// 		r := make([]interface{}, len(headers))

// 		for i, h := range headers {
// 			r[i] = row[h]
// 		}

// 		fmt.Fprintf(w, format, r...)
// 	}
// }
