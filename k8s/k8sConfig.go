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

type jobValues struct {
	NameJob     string
	ExecFiles   string
	Namespace   string
	Image       string
	Environment map[string]string
	CPU         string
	RAM         string
	Command     string
	ConfigNames []string
}

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

func K8smanagement(w http.ResponseWriter, kubernetesData ops.K8sValues) {
	//func K8smanagement(w http.ResponseWriter, namespace, image string, dataMap map[string]string, env map[string]string, cpu, ram string) {

	log.Println("Creating resources...")
	fmt.Fprintf(w, "Creating resources...\n")

	var executeFile string   // variable for dynamic command
	var configNames []string // variable for configMap names

	//Check if namespace exists
	statusNamespce, err := createNamespace(kubernetesData.Namespace)
	if err != nil {
		http.Error(w, "Failed to create namespace", http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "%s", statusNamespce)
	// Create configMap
	for name, data := range kubernetesData.Data {
		fmt.Fprintln(w, "Creating ConfigMap: ", name)
		configMap, err := createUpdateConfigMap(name, data, kubernetesData.Namespace)
		fmt.Fprintln(w, "Created ConfigMap: ", configMap)
		if err != nil {
			http.Error(w, "Failed to create/update ConfigMap", http.StatusInternalServerError)
			log.Println("Failed to create/update ConfigMap", err)
		}
		configNames = append(configNames, configMap)
	}

	// Create Job
	// Create job values
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

	jobVal := jobValues{
		NameJob:     jobName,
		ExecFiles:   executeFile,
		Namespace:   kubernetesData.Namespace,
		Image:       kubernetesData.Image,
		Environment: kubernetesData.Env,
		CPU:         kubernetesData.CPU,
		RAM:         kubernetesData.RAM,
		Command:     kubernetesData.Command,
		ConfigNames: configNames,
	}

	err = createOrUpdateJob(jobVal)
	if err != nil {
		http.Error(w, "Failed to create/update Job", http.StatusInternalServerError)
		log.Println("Failed to create/update Job", err)
	}

	fmt.Fprintln(w, "Job created...")
	log.Println("Job created...")
	waitForJobCompletion(jobName, kubernetesData.Namespace)
	logs, err := getJobLogs(jobName, kubernetesData.Namespace)
	if err != nil {
		http.Error(w, "Failed to get logs", http.StatusInternalServerError)
		log.Println("Failed to get logs", err)
	}
	fmt.Fprintf(w, "%s", logs)
	log.Println("Job completed successfully.")

}
