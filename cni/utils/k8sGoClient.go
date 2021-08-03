package utils

import (
	"context"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"strings"
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
	config, err := clientcmd.BuildConfigFromFlags("https://172.17.36.211:6443", "")
	//config, err := clientcmd.BuildConfigFromFlags("", "/etc/kubernetes/admin.conf")
	if err != nil {
		log.Println("出错了, err=", err.Error())
		panic(err.Error())
	}
	config.TLSClientConfig.Insecure = true
	config.BearerToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6InVsbGxVNmVuZzBYZFVHZUpMeENGVEo1ei1oNUluWHhvYl92dDFhRmRNZ2cifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJrdWJlLXN5c3RlbSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJhZG1pbi11c2VyLXRva2VuLWZoZzg5Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6ImFkbWluLXVzZXIiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiI0OTM4NjE3MC0zMGFiLTQzOWYtYjJhMi1lZDU1ZDg2NDE5ZjIiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6a3ViZS1zeXN0ZW06YWRtaW4tdXNlciJ9.h_W5z29HH2J9G0Zl3gjRKu1xa-I-vURLptjSSOLmHWdXzFb4zGLzVl1aKGWrlqNuOg1QGToFBRUosxZYRs7bcCstEmnGrIYwsyFCsVUEY7LCwV1Q4YiaKYs2NeSHULtQQI4jIFn4IFCCYsxBuoUjBWYHLhwLSMKmX9SEMRslC6R2PqSr7ZTzwfC0m0d-akJ_cVQpRlvo2Wj7RvNQojIbHzL4EZ9ofyv5C5JUFX0Sx9PT0hPxZqyVTddEsol9lujES4Ay0p50WzZRXBi75HvL3TJRyUmN4xumN7MjTphlgtFojclunZEUE9Jjem9LUnQXQxL-sxDRL0OClpqQ7eVKCQ"
	log.Println("k8s client config的值=", *config)
	// create the clientSet
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	log.Println("k8s client clientSet的值=", clientSet)
	return clientSet
}

// GetPodNet 如果是deployment，返回一个Annotations中key为app_net的切片
func (k *K8s) GetPodNet(NameSpace string, PodName string) []string {
	pods, err := k.client.CoreV1().Pods(NameSpace).Get(context.TODO(), PodName, metaV1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	log.Println("pods=", *pods)

	// 根据pod查找上层控制器
	if pods.OwnerReferences[0].Kind == "ReplicaSet" {
		repSet, err := k.client.AppsV1().ReplicaSets(NameSpace).Get(context.TODO(), pods.OwnerReferences[0].Name, metaV1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}
		//log.Println("repSet.OwnerReferences[0].Kind=", repSet.OwnerReferences[0].Kind)
		//log.Println("repSet.OwnerReferences[0].Name=", repSet.OwnerReferences[0].Name)
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
			log.Println("repSet.OwnerReferences[0].Name=", repSet.OwnerReferences[0].Name)
			return nil
		} else {
			return nil
		}

		// 没有上层控制器就返回nil
	} else {
		return nil
	}
}

func (k *K8s) GetPodIp(NameSpace string, PodName string) string {
	pods, err := k.client.CoreV1().Pods(NameSpace).Get(context.TODO(), PodName, metaV1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	return pods.Status.PodIP
}
