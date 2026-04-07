package commands

import (
	"context"
	"fmt"
	"os"

	"k8sctl/pkg/k8s"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
)

var (
	deleteDryRun bool
)

func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <manifest.yaml>",
		Short: "Delete resources defined in a Kubernetes manifest",
		Args:  cobra.ExactArgs(1),
		RunE:  runDelete,
	}

	cmd.Flags().BoolVar(&deleteDryRun, "dry-run", false, "Perform a dry run without deleting resources")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
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

	if deleteDryRun {
		fmt.Printf("DRY RUN: Would delete %s/%s in namespace %s\n", gvk.Kind, obj.GetName(), namespace)
		return nil
	}

	fmt.Printf("Deleting %s/%s in namespace %s...\n", gvk.Kind, obj.GetName(), namespace)

	var resource dynamic.ResourceInterface
	if namespace != "" {
		resource = dynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		resource = dynamicClient.Resource(gvr)
	}

	deletePolicy := metav1.DeletePropagationForeground
	err = resource.Delete(ctx, obj.GetName(), metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	fmt.Println("Resource deleted succesfully")

	return nil
}
