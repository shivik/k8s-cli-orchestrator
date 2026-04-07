package commands

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"k8sctl/pkg/k8s"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	logsNamespace string
	logsFollow    bool
	logsTail      int64
)

func LogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <pod>",
		Short: "Tail logs from a pod",
		Args:  cobra.ExactArgs(1),
		RunE:  runLogs,
	}

	cmd.Flags().StringVarP(&logsNamespace, "namespace", "n", "default", "Kubernetes namespace")
	cmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	cmd.Flags().Int64Var(&logsTail, "tail", 100, "Number of lines to show from the end of the logs")

	return cmd
}

func runLogs(cmd *cobra.Command, args []string) error {
	podName := args[0]

	client, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	ctx := context.Background()

	pod, err := client.Clientset.CoreV1().Pods(logsNamespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod '%s' in namespace '%s': %w", podName, logsNamespace, err)
	}

	containerName := ""
	if len(pod.Spec.Containers) > 0 {
		containerName = pod.Spec.Containers[0].Name
	}

	if len(pod.Spec.Containers) > 1 {
		fmt.Printf("Pod has mutiple containers, using container: %s\n", containerName)
	}

	logOptions := &corev1.PodLogOptions{
		Container: containerName,
		Follow:    logsFollow,
		TailLines: &logsTail,
	}

	req := client.Clientset.CoreV1().Pods(logsNamespace).GetLogs(podName, logOptions)
	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to get log stream for pod '%s': %w", podName, err)
	}
	defer stream.Close()

	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if line != "" {
					fmt.Print(line)
				}
				break
			}
			return fmt.Errorf("error reading logs: %w", err)
		}
		fmt.Print(line)
	}

	return nil
}
