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
	"os"
	"path/filepath"
)

var globalConfig config.Config

type ImageUploadRequest struct {
	Filename string `json:"filename"`
}

type ExecuteRequest struct {
	Scripts []string `json:"scripts"`
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

func uploadImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// Limit file size to 10MB. This line saves you from those accidental 100MB uploads!
	r.ParseMultipartForm(10 << 20)

	// Retrieve the file from form data
	file, handler, err := r.FormFile("artifact")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fmt.Fprintf(w, "Uploaded File: %s\n", handler.Filename)
	fmt.Fprintf(w, "File Size: %d\n", handler.Size)
	fmt.Fprintf(w, "MIME Header: %v\n", handler.Header)

	// Now let’s save it locally
	dst, err := ops.CreateFile(handler.Filename)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err := dst.ReadFrom(file); err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
	}
}

// Handler to serve images
func serveImageHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len("./uploads/"):] // Extract filename
	imageDir := "./uploads"
	imagePath := filepath.Join(imageDir, filename)
	fmt.Println(imagePath)
	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		fmt.Println(err)
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Serve the image

	http.ServeFile(w, r, imagePath)
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

	k8s.K8smanagement(w,
		globalConfig.Namespace,
		globalConfig.Image,
		data,
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
	http.HandleFunc("/upload", uploadImageHandler)
	http.HandleFunc("/artifacts/", serveImageHandler)

	http.ListenAndServe(port, nil)
}

//curl http://service-name.namespace.svc.cluster.local:port
