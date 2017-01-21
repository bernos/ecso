package environmentup

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type Options struct {
	EnvironmentName string
	DryRun          bool
}

func New(environmentName string, options ...func(*Options)) ecso.Command {
	o := &Options{
		EnvironmentName: environmentName,
	}

	for _, option := range options {
		option(o)
	}

	return &envUpCommand{
		options: o,
	}
}

type logfn func(string, ...interface{})

type envUpCommand struct {
	options *Options
}

func (cmd *envUpCommand) Execute(project *ecso.Project, cfg *ecso.Config, prefs ecso.UserPreferences) error {
	if err := validateOptions(cmd.options); err != nil {
		return err
	}

	environment, ok := project.Environments[cmd.options.EnvironmentName]

	if !ok {
		return fmt.Errorf("No environment named '%s' was found", cmd.options.EnvironmentName)
	}

	cfg.Logger.BannerBlue("Bringing up environment '%s'", environment.Name)

	if cmd.options.DryRun {
		cfg.Logger.Infof("THIS IS A DRY RUN - no changes to the environment will be made.")
	}

	templateDir, err := getTemplateDir()

	if err != nil {
		return err
	}

	if err := ensureTemplates(templateDir, cfg.Logger.Infof); err != nil {
		return err
	}

	cfg.Logger.Infof("Packaging infrastructure stack templates")

	if err := deployStack(cfg, project, environment, filepath.Join(templateDir, "stack.yaml"), cmd.options.DryRun); err != nil {
		return err
	}

	if cmd.options.DryRun {
		cfg.Logger.BannerGreen("Review the above changes and re-run the command without the --dry-run option to apply them")
	} else {
		cfg.Logger.BannerGreen("Environment '%s' is up and running", environment.Name)
	}

	return nil
}

func deployStack(cfg *ecso.Config, project *ecso.Project, env ecso.Environment, template string, dryRun bool) error {
	var (
		stackName = fmt.Sprintf("%s-%s", project.Name, env.Name)
		bucket    = env.CloudFormationBucket
		prefix    = path.Join(fmt.Sprintf("%s-%s", project.Name, env.Name), "infrastructure")
		params    = env.CloudFormationParameters
		tags      = env.CloudFormationTags
	)

	packagedTemplate, err := cfg.CloudFormationService.Package(template, bucket, prefix)

	if err != nil {
		return err
	}

	cfg.Logger.Printf("\n")
	cfg.Logger.Infof("Deploying infrastructure stack '%s'", stackName)

	id, err := cfg.CloudFormationService.Deploy(packagedTemplate, stackName, params, tags, dryRun)

	if err != nil {
		return err
	}

	changeSet, err := cfg.CloudFormationService.GetChangeSet(id)

	if err != nil {
		return err
	}

	if dryRun {
		cfg.Logger.BannerGreen("The following changes would be made to the environment:")
	} else {
		cfg.Logger.BannerGreen("The following changes were made to the environment:")
	}

	fmt.Printf("%#v\n", changeSet)

	return nil
}

func ensureTemplates(dst string, log logfn) error {
	exists, err := util.DirExists(dst)

	if err != nil || exists {
		return err
	}

	return copyTemplates(dst, log)
}

func getTemplateDir() (string, error) {
	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return "", err
	}

	return filepath.Join(wd, ".ecso", "infrastructure", "templates"), nil
}

func copyTemplates(dst string, log logfn) error {
	log("Copying infrastructure stack templates to %s", dst)

	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return err
	}

	for file, content := range templates {
		if err := ioutil.WriteFile(filepath.Join(dst, file), []byte(content), os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func validateOptions(opt *Options) error {
	if opt.EnvironmentName == "" {
		return fmt.Errorf("Environment name is required")
	}
	return nil
}
