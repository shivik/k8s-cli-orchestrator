package commands

import (
	"context"
	"fmt"

	"k8sctl/pkg/k8s"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	statusNamespace string
)

func StatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <deployment>",
		Short: "Show pod statuses for a deployment",
		Args:  cobra.ExactArgs(1),
		RunE:  runStatus,
	}

	cmd.Flags().StringVarP(&statusNamespace, "namespace", "n", "default", "Kubernetes namespace")

	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	deploymentName := args[0]

	client, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	ctx := context.Background()

	deployment, err := client.Clientset.AppsV1().Deployments(statusNamespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment '%s' in namespace '%s': %w", deploymentName, statusNamespace, err)
	}

	labelSelector := metav1.FormatLabelSelector(deployment.Spec.Selector)
	pods, err := client.Clientset.CoreV1().Pods(statusNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("failed to list pods for deployment '%s': %w", deploymentName, err)
	}

	fmt.Printf("Deployment: %s/%s\n", statusNamespace, deploymentName)
	fmt.Printf("Replicas: %d desired | %d updated | %d available | %d unavailable\n",
		*deployment.Spec.Replicas,
		deployment.Status.UpdatedReplicas,
		deployment.Status.AvailableReplicas,
		deployment.Status.UnavailableReplicas,
	)
	fmt.Println()

	if len(pods.Items) == 0 {
		fmt.Println("No pods found")
		return nil
	}

	fmt.Printf("%-40s %-15s %-10s %s\n", "POD NAME", "STATUS", "RESTARTS", "AGE")
	fmt.Println("--------------------------------------------------------------------------------------------")

	for _, pod := range pods.Items {
		status := string(pod.Status.Phase)
		restarts := int32(0)

		for _, containerStatus := range pod.Status.ContainerStatuses {
			restarts += containerStatus.RestartCount
		}

		age := metav1.Now().Sub(pod.CreationTimestamp.Time)
		ageStr := formatDuration(age)

		fmt.Printf("%-40s %-15s %-10d %s\n", pod.Name, status, restarts, ageStr)
	}

	return nil
}

func formatDuration(d interface{}) string {
	switch v := d.(type) {
	case metav1.Duration:
		return v.Duration.String()
	default:
		return fmt.Sprintf("%v", d)
	}
}
