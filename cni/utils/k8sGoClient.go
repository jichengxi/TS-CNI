package utils

import (
	"context"
	"flag"
	"fmt"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"strings"

	//"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

type K8s struct {
	client *kubernetes.Clientset
}

// NewK8s 新建一个k8s客户端连接
func NewK8s() *K8s {
	return &K8s{
		client: newK8sClient(),
	}
}

func newK8sClient() *kubernetes.Clientset {
	var kubeConfig *string
	currentDir, _ := os.Getwd()
	// 取k8s config配置文件的路径
	//kubeConfig = flag.String("kubeConfig", filepath.Join(currentDir, "..\\", "config-home"), "k8s配置文件所在位置")
	kubeConfig = flag.String("kubeConfig", filepath.Join(currentDir, "..\\", "config-36-211"), "k8s配置文件所在位置")
	//fmt.Println(kubeConfig)
	flag.Parse()
	// use the current context in kubeConfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		panic(err.Error())
	}
	// create the clientSet
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientSet
}

// GetPodNet 如果是deployment，返回一个Annotations中key为app_net的切片
func (k *K8s) GetPodNet(NameSpace string, PodName string) []string {
	pods, err := k.client.CoreV1().Pods(NameSpace).Get(context.TODO(), PodName, metaV1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	// 根据pod查找上层控制器
	if pods.OwnerReferences[0].Kind == "ReplicaSet" {
		repSet, err := k.client.AppsV1().ReplicaSets(NameSpace).Get(context.TODO(), pods.OwnerReferences[0].Name, metaV1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}
		//fmt.Println(repSet.OwnerReferences[0].Kind)
		//fmt.Println(repSet.OwnerReferences[0].Name)
		if repSet.OwnerReferences[0].Kind == "Deployment" {
			repDeploy, err := k.client.AppsV1().Deployments(NameSpace).Get(context.TODO(), repSet.OwnerReferences[0].Name, metaV1.GetOptions{})
			if err != nil {
				panic(err.Error())
			}
			NetSlice := strings.Split(repDeploy.Annotations["app_net"], ",")
			return NetSlice
		} else if pods.OwnerReferences[0].Kind == "StatefulSet" {
			repSet, err := k.client.AppsV1().StatefulSets(NameSpace).Get(context.TODO(), pods.OwnerReferences[0].Name, metaV1.GetOptions{})
			if err != nil {
				panic(err.Error())
			}
			fmt.Println(repSet.OwnerReferences[0].Name)
			return nil
		} else {
			return nil
		}

		// 没有上层控制器就返回nil
	} else {
		return nil
	}
}
