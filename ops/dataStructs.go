package ops

type K8sValues struct {
	Namespace string
	Image     string
	Command   string
	Data      map[string]string
	Env       map[string]string
	CPU       string
	RAM       string
}
