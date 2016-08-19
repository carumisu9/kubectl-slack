package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/util/wait"
)

func notifySlack(obj interface{}, action string) {

	pod := obj.(*api.Pod)
	url := "https://hooks.slack.com/services/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"
	json := `{"text": "Pod ` + action + ` in cluster: ` + pod.ObjectMeta.Name + `"}`
	client := http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(json))
	req.Header.Set("Content-Type", "application/json")
	_, err = client.Do(req)
	if err != nil {
		fmt.Println("Unable to reach the server.")
	}
}
func podCreated(obj interface{}) {
	notifySlack(obj, "created")
}
func podDeleted(obj interface{}) {
	notifySlack(obj, "deleted")
}
func watchPods(client *client.Client, store cache.Store) cache.Store {

	watchlist := cache.NewListWatchFromClient(client, "pods", api.NamespaceAll, fields.Everything())
	resyncPeriod := 30 * time.Minute

	eStore, eController := framework.NewInformer(
		watchlist,
		&api.Pod{},
		resyncPeriod,
		framework.ResourceEventHandlerFuncs{
			AddFunc:    podCreated,
			DeleteFunc: podDeleted,
		},
	)

	go eController.Run(wait.NeverStop)

	return eStore
}

func main() {

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = clientcmd.RecommendedHomeFile
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

	clientConfig, err := loader.ClientConfig()

	kubeClient, err := client.New(clientConfig)
	if err != nil {
		log.Fatalln("Client not created sucessfully:", err)
	}

	var podsStore cache.Store

	podsStore = watchPods(kubeClient, podsStore)

}
