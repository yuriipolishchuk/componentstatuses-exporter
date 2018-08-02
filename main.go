package main

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var (
	componentStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_componentstatus_healthy",
			Help: "Kubernetes componentstatus healthy",
		},
		[]string{"component"},
	)
)

func init() {
	prometheus.MustRegister(componentStatus)
}

func getComponentStatuses() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	for {
		componetstatuses, err := clientset.CoreV1().ComponentStatuses().List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		health_conditions := map[string]bool{
			"ok": true,
			"{\"health\": \"true\"}": true,
		}

		for _, componentstatus := range componetstatuses.Items {
			if health_conditions[componentstatus.Conditions[0].Message] {
				fmt.Printf("%s: %s\n", componentstatus.Name, "OK")
				componentStatus.With(prometheus.Labels{"component": componentstatus.Name}).Set(1.0)
			} else {
				fmt.Printf("%s: %s, message: %s\n", componentstatus.Name, "FAILURE", componentstatus.Conditions[0].Message)
				componentStatus.With(prometheus.Labels{"component": componentstatus.Name}).Set(0.0)
			}
		}

		refreshRate, err := strconv.Atoi(getEnv("COMPONENTSTATUSES_CHECK_RATE", "10"))
		if err != nil {
			panic(err.Error())
		}

		time.Sleep(time.Duration(refreshRate) * time.Second)
	}
}

func handleGracefulShutdown() {
	var gracefulStop = make(chan os.Signal)

	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v", sig)
		os.Exit(0)
	}()
}

func main() {
	handleGracefulShutdown()

	go getComponentStatuses()

	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(":8080", nil)
}
