package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/bernos/ecso/pkg/ecso/log"
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

type TableDataProvider interface {
	TableHeader() []string
	TableRows() []map[string]string
}

func PrintTable(logger log.Logger, data TableDataProvider) {
	headers := data.TableHeader()
	rows := data.TableRows()
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

	logger.Printf(format, headerRow...)

	for _, row := range rows {
		r := make([]interface{}, len(headers))

		for i, h := range headers {
			r[i] = row[h]
		}

		logger.Printf(format, r...)
	}
}

func PrintMap(logger log.Logger, maps ...map[string]string) {
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
		logger.Printf("%s %s\n", bold(labelFormat, k), v)
	}
}

// func PrintEnvironmentDescription(logger log.Logger, env *api.EnvironmentDescription) {
// 	childLogger := logger.Child()

// 	BannerBlue(logger, "Details of the '%s' environment:", env.Name)

// 	Dl(childLogger, map[string]string{
// 		"CloudFormation console": env.CloudFormationConsoleURL,
// 		"CloudWatch logs":        env.CloudWatchLogsConsoleURL,
// 		"ECS console":            env.ECSConsoleURL,
// 		"ECS base URL":           env.ECSClusterBaseURL,
// 	})

// 	BannerBlue(logger, "CloudFormation Outputs:")
// 	Dl(childLogger, env.CloudFormationOutputs)
// 	logger.Printf("\n")
// }

// func PrintServiceDescription(logger log.Logger, service *api.ServiceDescription) {
// 	childLogger := logger.Child()

// 	BannerBlue(logger, "Details of the '%s' service:", service.Name)

// 	Dl(childLogger, map[string]string{
// 		"CloudFormation console": service.CloudFormationConsoleURL,
// 		"CloudWatch logs":        service.CloudWatchLogsConsoleURL,
// 		"ECS console":            service.ECSConsoleURL,
// 	})

// 	if service.URL != "" {
// 		Dl(childLogger, map[string]string{
// 			"Service URL": service.URL,
// 		})
// 	}

// 	BannerBlue(logger, "CloudFormation Outputs:")
// 	Dl(childLogger, service.CloudFormationOutputs)
// 	logger.Printf("\n")
// }

type BlueBanner string

func (b BlueBanner) Format(f fmt.State, c rune) {
	f.Write([]byte(blueBold("\n%s\n\n", string(b))))
}

type Info string

func (i Info) Format(f fmt.State, c rune) {
	f.Write([]byte(fmt.Sprintf("%s %s\n", bold("Info:"), string(i))))
}

func Infof(format string, a ...interface{}) Info {
	return Info(fmt.Sprintf(format, a...))
}

type Error string

func (e Error) Format(f fmt.State, c rune) {
	f.Write([]byte(fmt.Sprintf("%s %s\n", redBold("Error:"), red("%s", string(e)))))
}

func Errorf(format string, a ...interface{}) Error {
	return Error(fmt.Sprintf(format, a...))
}

func BannerBlue(logger log.Logger, format string, a ...interface{}) {
	logger.Printf("\n%s\n\n", blueBold(format, a...))
}

func BannerGreen(logger log.Logger, format string, a ...interface{}) {
	logger.Printf("\n%s\n\n", greenBold(format, a...))
}

func Dt(logger log.Logger, label, content string) {
	logger.Printf("%s\n", bold("%s:", label))
	logger.Printf("  %s\n", content)
}

func Dl(logger log.Logger, items ...map[string]string) {
	for _, i := range items {
		for k, v := range i {
			Dt(logger, k, v)
		}
	}
}

type writerFunc func([]byte) (int, error)

func (fn writerFunc) Write(p []byte) (int, error) {
	return fn(p)
}

func ErrWriter(w io.Writer) io.Writer {
	return writerFunc(func(p []byte) (int, error) {
		return fmt.Fprint(w, Error(string(p)))
	})
}
