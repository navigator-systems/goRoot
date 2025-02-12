package k8s

import (
	"fmt"
	"goRoot/ops"
	"log"
	"net/http"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func K8sConfig() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		rules := clientcmd.NewDefaultClientConfigLoadingRules()
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
		config, err = kubeconfig.ClientConfig()
		if err != nil {
			log.Println("Error getting config", err)
			return nil, err
		}
	}

	clientset := kubernetes.NewForConfigOrDie(config)
	return clientset, nil
}

func K8smanagement(w http.ResponseWriter, namespace, image string, dataMap map[string]string) {
	log.Println("Creating resources...")
	fmt.Fprintf(w, "Creating resources...\n")

	var executeFile string   // variable for dynamic command
	var configNames []string // variable for configMap names

	//Check if namespace exists
	statusNamespce, err := createNamespace(namespace)
	if err != nil {
		http.Error(w, "Failed to create namespace", http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "%s", statusNamespce)
	// Create configMap
	for name, data := range dataMap {
		fmt.Fprintln(w, "Creating ConfigMap: ", name)
		configMap, err := createUpdateConfigMap(name, data, namespace)
		fmt.Fprintln(w, "Created ConfigMap: ", configMap)
		if err != nil {
			http.Error(w, "Failed to create/update ConfigMap", http.StatusInternalServerError)
			log.Println("Failed to create/update ConfigMap", err)
		}
		configNames = append(configNames, configMap)
	}

	jobName := ops.CreateName()
	fmt.Fprintln(w, "Creating Job: ", jobName)

	if len(configNames) == 1 {
		executeFile = fmt.Sprintf("/app/%s", configNames[0])
	} else {
		for _, str := range configNames {
			if strings.Contains(str, "execute.c") {
				executeFile = fmt.Sprintf("/app/%s", str)
			}
		}

	}

	err = createOrUpdateJob(jobName, executeFile, namespace, image, configNames)
	if err != nil {
		http.Error(w, "Failed to create/update Job", http.StatusInternalServerError)
		log.Println("Failed to create/update Job", err)
	}

	fmt.Fprintln(w, "Job created...")
	log.Println("Job created...")
	waitForJobCompletion(jobName, namespace)
	logs, err := getJobLogs(jobName, namespace)
	if err != nil {
		http.Error(w, "Failed to get logs", http.StatusInternalServerError)
		log.Println("Failed to get logs", err)
	}
	fmt.Fprintf(w, "%s", logs)
	log.Println("Job completed successfully.")

}
