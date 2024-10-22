package k8s

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RunGetPods retrieves the pod information and prints it in table format using tabwriter
func RunGetPods(namespace string, outStream io.Writer) error {
	// Kubernetes クライアントセットを取得
	clientset, err := GetClientset()
	if err != nil {
		return fmt.Errorf("Failed to create Kubernetes clientset: %v", err)
	}

	// 指定されたネームスペースの Pod リストを取得
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Failed to list pods: %v", err)
	}

	// tabwriter を使ってテーブル形式で出力
	w := tabwriter.NewWriter(outStream, 0, 0, 3, ' ', 0) // tabwriter を初期化
	fmt.Fprintln(w, "POD NAME\tIP ADDRESS\tSTATUS")      // ヘッダー
	for _, pod := range pods.Items {
		fmt.Fprintf(w, "%s\t%s\t%s\n", pod.Name, pod.Status.PodIP, string(pod.Status.Phase))
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("Failed to flush tabwriter: %v", err)
	}
	return nil
}
