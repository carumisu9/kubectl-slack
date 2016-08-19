package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/util/wait"
)

func notifySlack(obj interface{}, action string) {

	envLoad()

	pod := obj.(*api.Pod)

	//Incoming Webhook URL
	url := os.Getenv("WEBHOOK_URL")

	//Form JSON payload to send to Slack
	json := `{"text": "Pod ` + action + ` in cluster: ` + pod.ObjectMeta.Name + ` namespace: ` + pod.ObjectMeta.Namespace + `"}`

	//Post JSON payload to the Webhook URL
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

	//Define what we want to look for (Pods)
	watchlist := cache.NewListWatchFromClient(client, "pods", api.NamespaceAll, fields.Everything())
	resyncPeriod := 30 * time.Minute

	//Setup an informer to call functions when the watchlist changes
	eStore, eController := framework.NewInformer(
		watchlist,
		&api.Pod{},
		resyncPeriod,
		framework.ResourceEventHandlerFuncs{
			AddFunc:    podCreated,
			DeleteFunc: podDeleted,
		},
	)

	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)

	return eStore
}

func main() {

	// Get Kube Config
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = clientcmd.RecommendedHomeFile
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	clientConfig, err := loader.ClientConfig()

	//Create a new client to interact with cluster and freak if it doesn't work
	kubeClient, err := client.New(clientConfig)
	if err != nil {
		log.Fatalln("Client not created sucessfully:", err)
	}

	//Create a cache to store Pods
	var podsStore cache.Store

	//Watch for Pods
	podsStore = watchPods(kubeClient, podsStore)

	//Keep alive
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Get env value from .env
func envLoad() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
