package usecase

import (
	"context"
	"fmt"

	"github.com/naka-gawa/kubectl-sgmap/pkg/aws"
	"github.com/naka-gawa/kubectl-sgmap/pkg/kubernetes"
	"github.com/naka-gawa/kubectl-sgmap/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// GetPodSecurityGroups は Pod の ENI およびセキュリティグループを取得し、指定の形式で出力する
func GetPodSecurityGroups(streams *genericclioptions.IOStreams, outputMode string) error {
	clientset, err := kubernetes.GetClientset()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes clientset: %v", err)
	}

	// Pod の一覧を取得
	pods, err := clientset.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	// Pod の情報を取得
	var podInfos []utils.PodInfo
	for _, pod := range pods.Items {
		eniID, sgIDs, err := aws.GetENIForIPAddress(pod.Status.PodIP)
		if err != nil {
			eniID = "ENI not found"
		}
		podInfos = append(podInfos, utils.PodInfo{
			PODNAME:          pod.Name,
			IPADDRESS:        pod.Status.PodIP,
			ENIID:            eniID,
			SECURITYGROUPIDS: sgIDs,
		})
	}

	// 指定の出力形式でデータを表示
	switch outputMode {
	case "json":
		return utils.OutputJSON(podInfos, streams.Out)
	case "yaml":
		return utils.OutputYAML(podInfos, streams.Out)
	case "table":
		return utils.OutputTable(podInfos, streams.Out)
	}

	return fmt.Errorf("unsupported output format '%s' found", outputMode)
}
