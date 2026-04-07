package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8sctl/pkg/k8s"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
)

var (
	deployDryRun bool
)

func DeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy <manifest.yaml>",
		Short: "Apply a Kubernetes manifest and wait for rollout",
		Args:  cobra.ExactArgs(1),
		RunE:  runDeploy,
	}

	cmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "Perform a dry run without applying changes")

	return cmd
}

func runDeploy(cmd *cobra.Command, args []string) error {
	manifestPath := args[0]

	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest file '%s': %w", manifestPath, err)
	}

	client, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(client.Config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, gvk, err := decoder.Decode(manifestData, nil, obj)
	if err != nil {
		return fmt.Errorf("failed to decode manifest: %w", err)
	}

	gvr, err := k8s.GetGVR(client.Config, gvk)
	if err != nil {
		return fmt.Errorf("failed to get resource mapping: %w", err)
	}

	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	ctx := context.Background()

	if deployDryRun {
		fmt.Printf("DRY RUN: Would apply %s/%s in namespace %s\n", gvk.Kind, obj.GetName(), namespace)
		return nil
	}

	fmt.Printf("Appling %s/%s in namespace %s...\n", gvk.Kind, obj.GetName(), namespace)

	var resource dynamic.ResourceInterface
	if namespace != "" {
		resource = dynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		resource = dynamicClient.Resource(gvr)
	}

	_, err = resource.Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		_, err = resource.Update(ctx, obj, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to apply manifest: %w", err)
		}
		fmt.Println("Resource updated succesfully")
	} else {
		fmt.Println("Resource created succesfully")
	}

	if gvk.Kind == "Deployment" {
		fmt.Println("Waiting for rollout to compelte...")
		err = waitForDeployment(ctx, client, obj.GetName(), namespace)
		if err != nil {
			return fmt.Errorf("rollout failed: %w", err)
		}
		fmt.Println("Rollout completed successfully")
	}

	return nil
}

func waitForDeployment(ctx context.Context, client *k8s.Client, name, namespace string) error {
	return wait.PollImmediate(2*time.Second, 5*time.Minute, func() (bool, error) {
		deployment, err := client.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if deployment.Status.UpdatedReplicas == *deployment.Spec.Replicas &&
			deployment.Status.Replicas == *deployment.Spec.Replicas &&
			deployment.Status.AvailableReplicas == *deployment.Spec.Replicas &&
			deployment.Status.ObservedGeneration >= deployment.Generation {
			return true, nil
		}

		return false, nil
	})
}
