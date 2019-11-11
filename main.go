package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	componentStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_componentstatuses",
			Help: "Kubernetes component status health",
		},
		[]string{"component"},
	)

	refreshRate int
)

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getComponentStatuses() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	checkErr(err)

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	checkErr(err)

	refreshRate, err := strconv.Atoi(getEnv("COMPONENTSTATUSES_CHECK_RATE", "10"))
	checkErr(err)

	for {
		// get component statuses
		componetstatuses, err := clientset.CoreV1().ComponentStatuses().List(metav1.ListOptions{})
		checkErr(err)

		for _, componentstatus := range componetstatuses.Items {
			var metricValue float64

			status := componentstatus.Conditions[0].Status
			msg := fmt.Sprintf("%s: %s", componentstatus.Name, componentstatus.Conditions[0].Message)

			if status == "True" {
				metricValue = 1.0
				log.Info(msg)
			} else {
				metricValue = 0.0
				log.Error(msg)
			}

			// export metrics
			componentStatus.With(prometheus.Labels{"component": componentstatus.Name}).Set(metricValue)

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
		log.Info(fmt.Sprintf("Caught signal: %v", sig))
		os.Exit(0)
	}()
}

func init() {
	handleGracefulShutdown()

	//  configure logger
	log.RegisterExitHandler(handleGracefulShutdown)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	level, err := log.ParseLevel(getEnv("LOG_LEVEL", "info"))
	checkErr(err)
	log.SetLevel(level)

	prometheus.MustRegister(componentStatus)
}

func main() {
	go getComponentStatuses()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
