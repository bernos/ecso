package ui

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

func AskString(r io.Reader, w io.Writer, prompt, def string, v StringValidator) (string, error) {
	str := ""
	err := AskStringVar(r, w, &str, prompt, def, v)

	return str, err
}

func AskStringIfEmptyVar(r io.Reader, w io.Writer, dst *string, prompt, def string, v StringValidator) error {
	if len(*dst) == 0 {
		return AskStringVar(r, w, dst, prompt, def, v)
	}
	return nil
}

func AskStringVar(r io.Reader, w io.Writer, dst *string, prompt, def string, v StringValidator) error {
	scanner := bufio.NewScanner(r)

	if len(def) > 0 {
		prompt = fmt.Sprintf("%s (%s)", prompt, def)
	}

	fmt.Fprintf(w, "%s\n", bold("%s: ", prompt))

	for {
		fmt.Fprint(w, " > ")

		if scanner.Scan() {
			str := scanner.Text()

			if len(str) == 0 {
				str = def
			}

			if err := v.Validate(str); err != nil {
				fmt.Fprintf(w, "   %s\n", warn(err.Error()))
			} else {
				*dst = str
				return nil
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}
	}
}

func AskInt(r io.Reader, w io.Writer, prompt string, def int, v IntValidator) (int, error) {
	i := 0
	err := AskIntVar(r, w, &i, prompt, def, v)

	return i, err
}

func AskIntIfEmptyVar(r io.Reader, w io.Writer, dst *int, prompt string, def int, v IntValidator) error {
	if *dst == 0 {
		return AskIntVar(r, w, dst, prompt, def, v)
	}
	return nil
}

func AskIntVar(r io.Reader, w io.Writer, dst *int, prompt string, def int, v IntValidator) error {
	scanner := bufio.NewScanner(r)

	if def != 0 {
		prompt = fmt.Sprintf("%s (%d)", prompt, def)
	}

	fmt.Fprintf(w, "%s\n", bold("%s: ", prompt))

	for {
		fmt.Fprint(w, " > ")

		if scanner.Scan() {
			str := scanner.Text()

			if len(str) == 0 {
				str = strconv.Itoa(def)
			}

			i, err := strconv.Atoi(str)

			if err != nil {
				fmt.Fprintf(w, "   %s\n", warn("Please enter a number"))
			} else {
				if err := v.Validate(i); err != nil {
					fmt.Fprintf(w, "   %s\n", warn(err.Error()))
				} else {
					*dst = i
					return nil
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}
	}
}

func Choice(r io.Reader, w io.Writer, prompt string, choices []string) (int, error) {
	i := 0
	err := ChoiceVar(r, w, &i, prompt, choices)

	return i, err
}

func ChoiceVar(r io.Reader, w io.Writer, dst *int, prompt string, choices []string) error {
	scanner := bufio.NewScanner(r)

	fmt.Fprintf(w, "%s\n", bold("%s: ", prompt))

	for i, choice := range choices {
		fmt.Fprintf(w, " %d) %s\n", i+1, choice)
	}

	for {
		fmt.Fprint(w, "  > ")

		if scanner.Scan() {

			str := scanner.Text()

			i, err := strconv.Atoi(str)
			if err != nil || i < 1 || i > len(choices) {
				fmt.Fprintf(w, "   %s\n", warn("Please enter a number between %d and %d", 1, len(choices)))
			} else {
				*dst = i - 1
				return nil
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}

	}
}
