package server

import (
	"encoding/json"
	"fmt"
	"goRoot/config"
	"goRoot/k8s"
	"goRoot/ops"
	"html/template"
	"io"
	"log"
	"net/http"
)

var globalConfig config.Config

type ImageUploadRequest struct {
	Filename string `json:"filename"`
}

type ExecuteRequest struct {
	Scripts []string          `json:"scripts"`
	Command string            `json:"command"`
	Env     map[string]string `json:"env"`
	CPU     string            `json:"cpu"`
	RAM     string            `json:"ram"`
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	scripts, err := ops.ListScripts(globalConfig.Directory)
	if err != nil {
		http.Error(w, "Failed to list scripts", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("frontend/index.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, scripts)
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Executing script")
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req ExecuteRequest
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()
	err = json.Unmarshal(body, &req)

	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	if len(req.Scripts) == 0 {
		http.Error(w, "No scripts selected", http.StatusBadRequest)
		return
	}

	data := make(map[string]string)
	for _, script := range req.Scripts {

		scriptName, scriptData := ops.ReadScript(script, globalConfig.Directory)
		data[scriptName] = scriptData

	}
	var command string
	if req.Command != "" {
		command = req.Command
	} else {
		command = globalConfig.Command
	}
	Values := ops.K8sValues{
		Namespace: globalConfig.Namespace,
		Image:     globalConfig.Image,
		Command:   command,
		Data:      data,
		Env:       req.Env,
		CPU:       req.CPU,
		RAM:       req.RAM,
	}
	k8s.K8smanagement(w,
		Values,
	)

}

func MainServer(cfg config.Config) {
	globalConfig = cfg
	port := fmt.Sprintf(":%d", globalConfig.Port)

	if cfg.Directory == "" {
		log.Println("Directory must be specified in the config file")

	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/execute", executeHandler)

	http.ListenAndServe(port, nil)
}

//curl http://service-name.namespace.svc.cluster.local:port
