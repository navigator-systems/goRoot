package k8s

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createOrUpdateJob(name, execFiles, namespace, image, command string, configNames []string) error {
	ctx := context.TODO()

	clientset, _ := K8sConfig()
	// Check if the job exists
	_, err := clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		log.Println("Job already exists, deleting old job before recreating...")
		deletePolicy := metav1.DeletePropagationForeground
		err = clientset.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		})
		time.Sleep(10 * time.Second)
		if err != nil {
			log.Fatalf("Failed to delete old job: %v", err)
		}
	}

	// Create volume projections for each ConfigMap
	var projections []corev1.VolumeProjection
	for _, configName := range configNames {
		projections = append(projections, corev1.VolumeProjection{
			ConfigMap: &corev1.ConfigMapProjection{
				LocalObjectReference: corev1.LocalObjectReference{Name: configName},
			},
		})
	}

	// Define the projected volume
	permPtr := new(int32)
	*permPtr = 0777

	volumes := []corev1.Volume{
		{
			Name: "script-volume",
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources:     projections,
					DefaultMode: permPtr,
				},
			},
		},
	}

	container := []corev1.Container{
		{
			Name:    "script-runner",
			Image:   image,
			Command: []string{command, fmt.Sprintf("%s", execFiles)},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "script-volume",
					MountPath: "/app",
				},
			},
		},
	}

	// Define Job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers:    container,
					RestartPolicy: corev1.RestartPolicyNever,
					Volumes:       volumes,
				},
			},
		},
	}

	// Create Job
	_, err = clientset.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Failed to create Job: %v", err)
	}
	log.Println("Job created successfully.")
	return nil
}

func waitForJobCompletion(jobName, namespace string) {
	ctx := context.TODO()
	clientset, _ := K8sConfig()
	for {
		job, err := clientset.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			log.Printf("Failed to get job: %v", err)
			break
		}

		if job.Status.Succeeded > 0 {
			log.Println("✅ Job completed successfully.")
			break
		}
		if job.Status.Failed > 0 {
			log.Printf("❌ Job failed!")
			break
		}

		log.Println("⏳ Waiting for Job to complete...")
		time.Sleep(2 * time.Second)
	}
}

func getJobLogs(jobName, namespace string) (string, error) {
	clientset, _ := K8sConfig()
	ctx := context.TODO()

	// Find the Pod associated with the Job
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list pods: %v", err)
	}

	if len(podList.Items) == 0 {
		return "", fmt.Errorf("no pods found for job %s", jobName)
	}

	podName := podList.Items[0].Name
	log.Printf("📄 Fetching logs from Pod: %s\n", podName)

	// Get logs from the first Pod
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{})
	logStream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %v", err)
	}
	defer logStream.Close()

	// Print logs
	logData, err := io.ReadAll(logStream)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %v", err)
	}

	log.Println("📜 Logs from Job:")
	log.Println(string(logData))
	return string(logData), nil
}
