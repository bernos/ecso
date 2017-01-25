package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/color"
)

func ValidateIntBetween(min, max int) func(int) error {
	return func(v int) error {
		if v < min || v > max {
			return fmt.Errorf("A value between %d and %d is required", min, max)
		}
		return nil
	}
}

func ValidateAny() func(string) error {
	return func(v string) error {
		return nil
	}
}

func ValidateNotEmpty(msg string) func(string) error {
	return func(v string) error {
		if v == "" {
			return fmt.Errorf(msg)
		}
		return nil
	}
}

func ValidateRequired(name string) func(string) error {
	return ValidateNotEmpty(fmt.Sprintf("%s is required.", name))
}

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

	bold := color.New(color.Bold).SprintfFunc()
	reader := bufio.NewReader(os.Stdin)
	warn := color.New(color.FgRed).SprintFunc()

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
			fmt.Printf(" %s\n", warn(err.Error()))
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

	bold := color.New(color.Bold).SprintfFunc()
	reader := bufio.NewReader(os.Stdin)
	warn := color.New(color.FgRed).SprintFunc()

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
			fmt.Printf(" %s\n", warn("Please enter a number"))
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
