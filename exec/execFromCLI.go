package exec

import (
	"encoding/json"
	"fmt"
	"goRoot/config"
	"goRoot/k8s"
	"goRoot/ops"
	"net/http/httptest"
)

type ArgsRequest struct {
	Memory  string       `json:"memory"`
	CPU     string       `json:"cpu"`
	EnvVars []structEnvs `json:"envVars"`
}

type structEnvs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func CLIExec(cfg config.Config, script, envs string) {
	fmt.Println("CLI Executing")
	var req ArgsRequest

	err := json.Unmarshal([]byte(envs), &req)
	if err != nil {
		fmt.Println("Failed to unmarshal envs")
		return
	}

	// read data from script
	scriptName, scriptData := ops.ReadScript(script, cfg.Directory)

	data := make(map[string]string)
	data[scriptName] = scriptData

	// read env to proper map
	envMap := make(map[string]string)

	for _, env := range req.EnvVars {
		envMap[env.Key] = env.Value
	}

	Values := ops.K8sValues{
		Namespace:     cfg.Namespace,
		Image:         cfg.Image,
		Command:       cfg.Command,
		Data:          data,
		Env:           envMap,
		CPU:           req.CPU,
		RAM:           req.Memory,
		SharedStorage: cfg.SharedStorage,
	}

	fmt.Println("Running from terminal: ")
	rec := httptest.NewRecorder()
	k8s.K8smanagement(rec, Values)

}
