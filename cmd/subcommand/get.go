package cmd

import (
	"context"
	"fmt"

	"github.com/naka-gawa/kubectl-sg4pod/internal/k8s"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func getCmd() *cobra.Command {
	configFlags := genericclioptions.NewConfigFlags(true)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve resources such as Pods from the Kubernetes cluster",
		Long: `The 'get' command allows you to retrieve various resources from a Kubernetes cluster.
By default, this command will fetch a list of Pods from the specified namespace, or from the default namespace if no namespace is specified.

You can use this command to inspect the status, IP addresses, and other details of Pods running in the cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			getPodAddresses(cmd, configFlags)
		},
	}

	configFlags.AddFlags(cmd.PersistentFlags())
	return cmd
}

func getPodAddresses(cmd *cobra.Command, configFlags *genericclioptions.ConfigFlags) error {
	// GetOptions from k8s package
	streams := genericclioptions.IOStreams{In: cmd.InOrStdin(), Out: cmd.OutOrStdout(), ErrOut: cmd.ErrOrStderr()}
	options := k8s.GetOptions(streams)
	options.ConfigFlags = configFlags
	options.SetNamespace()

	// setup Kubernetes client
	client, err := k8s.GetClientset()
	if err != nil {
		fmt.Errorf("Failed to get Kubernetes clientset: %v", err)
		return err
	}

	// get list of pods in the namespace
	pods, err := client.Clientset.CoreV1().Pods(options.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Errorf("Failed to list pods: %v", err)
		return err
	}

	fmt.Println("Pods in the namespace", options.Namespace, ":")
	for _, pod := range pods.Items {
		fmt.Printf(" %s - %s\n", pod.Name, pod.Status.PodIP)
	}
	return nil
}
