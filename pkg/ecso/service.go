package ecso

type Service struct {
	Name         string
	ComposeFile  string
	DesiredCount int
	Route        string
	Port         int
	Tags         map[string]string
}
