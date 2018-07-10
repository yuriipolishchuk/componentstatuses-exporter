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
	"strconv"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var (
	schedulerHealthy = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_scheduler_healthy",
			Help: "Scheduler healthy",
		},
		[]string{"job"},
	)

	controllerManagerHealthy = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_controller_manager_healthy",
			Help: "Control-manager healthy",
		},
		[]string{"job"},
	)

	etcdHealthy = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_etcd_healthy",
			Help: "etcd healthy",
		},
		[]string{"job"},
	)
)

func init() {
	prometheus.MustRegister(schedulerHealthy)
	prometheus.MustRegister(controllerManagerHealthy)
	prometheus.MustRegister(etcdHealthy)
}

func reportStatus(component string, healthy float64) {
	switch component {
	case "scheduler":
		schedulerHealthy.With(prometheus.Labels{"job": "kube-scheduler"}).Set(healthy)
	case "controller-manager":
		controllerManagerHealthy.With(prometheus.Labels{"job": "kube-controller-manager"}).Set(healthy)
	case "etcd-0":
		etcdHealthy.With(prometheus.Labels{"job": "kube-etcd"}).Set(healthy)
	default:
		fmt.Printf("Unknown component %s.", component)
	}
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
				reportStatus(componentstatus.Name, 1)
			} else {
				fmt.Printf("%s: %s, message: %s\n", componentstatus.Name, "FAILURE", componentstatus.Conditions[0].Message)
				reportStatus(componentstatus.Name, 0)
			}
		}

		refreshRate, err := strconv.Atoi(getEnv("COMPONENTSTATUSES_CHECK_RATE", "10"))
		if err != nil {
			panic(err.Error())
		}

		time.Sleep(time.Duration(refreshRate) * time.Second)
	}
}

func main() {
	go getComponentStatuses()

	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(":8080", nil)
}
