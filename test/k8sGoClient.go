package main

import (
	"context"
	"flag"
	"fmt"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func homeDir() string {
	if h := os.Getenv("GOPATH"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}

func main() {
	var kubeConfig *string
	currentDir, _ := os.Getwd()
	if home := homeDir(); home != "" {
		kubeConfig = flag.String("kubeConfig", filepath.Join(currentDir, "..\\", "config"), "(optional) absolute path to the kubeConfig file")
	} else {
		kubeConfig = flag.String("kubeConfig", "", "absolute path to the kubeConfig file")
	}
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
	// List(context.TODO(), metaV1.ListOptions{})
	pods, err := clientSet.CoreV1().Pods("default").Get(context.TODO(), "nginx-test-847b659596-c7dgj", metaV1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	if pods.OwnerReferences[0].Kind == "ReplicaSet" {
		repSet, err := clientSet.AppsV1().ReplicaSets("default").Get(context.TODO(), pods.OwnerReferences[0].Name, metaV1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(repSet.OwnerReferences[0].Kind)
		fmt.Println(repSet.OwnerReferences[0].Name)
	} else if pods.OwnerReferences[0].Kind == "StatefulSet" {
		repSet, err := clientSet.AppsV1().StatefulSets("default").Get(context.TODO(), pods.OwnerReferences[0].Name, metaV1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(repSet.OwnerReferences[0].Name)
	}
	//fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	// Examples for error handling:
	// - Use helper functions like e.g. errors.IsNotFound()
	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	//namespace := "default"
	//pod := "example-xxxxx"
	//_, err = clientset.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
	//if errors.IsNotFound(err) {
	//	fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
	//} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
	//	fmt.Printf("Error getting pod %s in namespace %s: %v\n",
	//		pod, namespace, statusError.ErrStatus.Message)
	//} else if err != nil {
	//	panic(err.Error())
	//} else {
	//	fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
	//}
	//
	//time.Sleep(10 * time.Second)
}
