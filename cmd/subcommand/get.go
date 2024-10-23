package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/naka-gawa/kubectl-sgmap/internal/aws"
	"github.com/naka-gawa/kubectl-sgmap/internal/k8s"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// GetPodOptions provides the options for the get pods command
type GetPodOptions struct {
	configFlags *genericclioptions.ConfigFlags
	IOStreams   genericclioptions.IOStreams
	Namespace   string
}

// NewGetPodOptions provides an instance of GetPodOptions with default values
func NewGetPodOptions(streams genericclioptions.IOStreams) *GetPodOptions {
	return &GetPodOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

// NewCmdGetPod provides a cobra command to retrieve pod information
func NewCmdGetPod(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewGetPodOptions(streams)

	cmd := &cobra.Command{
		Use:   "get-pods",
		Short: "Retrieve pods information from a Kubernetes cluster",
		Long: `The 'get-pods' command allows you to retrieve a list of Pods in the specified namespace from the Kubernetes cluster.
You can also specify the namespace using the '-n' or '--namespace' flag. If no namespace is provided, the default namespace is used.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.Namespace, _, _ = o.configFlags.ToRawKubeConfigLoader().Namespace()
			if len(o.Namespace) == 0 {
				o.Namespace = "default" // デフォルトネームスペースを設定
			}

			return o.RunGetPods()
		},
	}

	o.configFlags.AddFlags(cmd.Flags()) // フラグに汎用オプションを追加
	return cmd
}

// RunGetPods retrieves the pod information and prints it in table format using tabwriter
func (o *GetPodOptions) RunGetPods() error {
	// Kubernetes クライアントセットを取得
	clientset, err := k8s.GetClientset()
	if err != nil {
		return fmt.Errorf("Failed to create Kubernetes clientset: %v", err)
	}

	// 指定されたネームスペースの Pod リストを取得
	pods, err := clientset.CoreV1().Pods(o.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Failed to list pods: %v", err)
	}

	// tabwriter を使ってテーブル形式で出力
	// IOStreams.Out が正しく初期化されていることを確認する
	if o.IOStreams.Out == nil {
		o.IOStreams.Out = os.Stdout
	}

	w := tabwriter.NewWriter(o.IOStreams.Out, 0, 0, 3, ' ', 0)          // tabwriter を初期化
	fmt.Fprintln(w, "POD NAME\tIP ADDRESS\tENI ID\tSECURITY GROUP IDS") // ヘッダー
	for _, pod := range pods.Items {
		// Pod の IP アドレスから ENI を取得
		eniID, sgIDs, err := aws.GetENIForIPAddress(pod.Status.PodIP)
		if err != nil {
			eniID = "ENI not found"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", pod.Name, pod.Status.PodIP, eniID, sgIDs)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("Failed to flush tabwriter: %v", err)
	}
	return nil
}
