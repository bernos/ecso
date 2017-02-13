package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/fatih/color"
)

var (
	bold     = color.New(color.Bold).SprintfFunc()
	warn     = color.New(color.FgRed).SprintFunc()
	blue     = color.New(color.FgBlue).SprintfFunc()
	blueBold = color.New(color.FgBlue, color.Bold).SprintfFunc()

	green     = color.New(color.FgGreen).SprintfFunc()
	greenBold = color.New(color.FgGreen, color.Bold).SprintfFunc()

	red     = color.New(color.FgRed).SprintfFunc()
	redBold = color.New(color.FgRed, color.Bold).SprintfFunc()
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

func PrintTable(w io.Writer, headers []string, rows ...map[string]string) {
	format := ""

	for _, h := range headers {
		l := len(h)

		for _, r := range rows {
			if v, ok := r[h]; ok && len(v) > l {
				l = len(v)
			}
		}

		format = format + fmt.Sprintf("%%-%ds  ", l)
	}

	format = format + "\n"

	headerRow := make([]interface{}, len(headers))

	for i, h := range headers {
		headerRow[i] = h
	}

	fmt.Fprintf(w, format, headerRow...)

	for _, row := range rows {
		r := make([]interface{}, len(headers))

		for i, h := range headers {
			r[i] = row[h]
		}

		fmt.Fprintf(w, format, r...)
	}
}

func PrintMap(w io.Writer, maps ...map[string]string) {
	l := 0
	items := make(map[string]string)

	for _, m := range maps {
		for k, v := range m {
			if len(k) > l {
				l = len(k)
			}
			items[k] = v
		}
	}

	labelFormat := fmt.Sprintf("  %%%ds:", l)

	for k, v := range items {
		fmt.Fprintf(w, "%s %s\n", bold(labelFormat, k), v)
	}
}

func PrintEnvironmentDescription(env *api.EnvironmentDescription, logger ecso.Logger) {
	BannerBlue(logger, "Details of the '%s' environment:", env.Name)

	Dl(logger, map[string]string{
		"CloudFormation console": env.CloudFormationConsoleURL,
		"CloudWatch logs":        env.CloudWatchLogsConsoleURL,
		"ECS console":            env.ECSConsoleURL,
		"ECS base URL":           env.ECSClusterBaseURL,
	})

	BannerBlue(logger, "CloudFormation Outputs:")
	Dl(logger, env.CloudFormationOutputs)
	logger.Printf("\n")
}

func PrintServiceDescription(service *api.ServiceDescription, logger ecso.Logger) {
	BannerBlue(logger, "Details of the '%s' service:", service.Name)

	Dl(logger, map[string]string{
		"CloudFormation console": service.CloudFormationConsoleURL,
		"CloudWatch logs":        service.CloudWatchLogsConsoleURL,
		"ECS console":            service.ECSConsoleURL,
	})

	if service.URL != "" {
		Dl(logger, map[string]string{
			"Service URL": service.URL,
		})
	}

	BannerBlue(logger, "CloudFormation Outputs:")
	Dl(logger, service.CloudFormationOutputs)
	logger.Printf("\n")
}

func BannerBlue(l ecso.Logger, format string, a ...interface{}) {
	l.Printf("\n%s\n\n", blueBold(format, a...))
}

func BannerGreen(l ecso.Logger, format string, a ...interface{}) {
	l.Printf("\n%s\n\n", greenBold(format, a...))
}

func Dt(l ecso.Logger, label, content string) {
	l.Printf("  %s\n    %s\n", bold("%s:", label), content)
}

func Dl(l ecso.Logger, items ...map[string]string) {
	for _, i := range items {
		for k, v := range i {
			Dt(l, k, v)
		}
	}
}
