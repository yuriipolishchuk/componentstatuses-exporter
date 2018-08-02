# kube-componentstatuses-prometheus-exporter
Simple app which polls kubernetes component statuses over API and exports them them via HTTP for Prometheus consumption

Can be usefull for monitoring k8s components statuses on managed solutitions like AWS EKS.

# Installation
```
git clone git@github.com:yuriipolishchuk/kube-componentstatuses-prometheus-exporter.git
cd kube-componentstatuses-prometheus-exporter/helm

helm install .
```

# Results
```
export POD_NAME=$(kubectl get pods --namespace core -l "app=kube-componentstatuses-prometheus-exporter,release=k8s-statuses" -o jsonpath="{.items[0].metadata.name}")

kubectl port-forward $POD_NAME 8080:8080 &
Forwarding from 127.0.0.1:8080 -> 8080
Handling connection for 8080

http localhost:8080/metrics | grep kube
Handling connection for 8080
# HELP kube_componentstatus_healthy Kubernetes componentstatus healthy
# TYPE kube_componentstatus_healthy gauge
kube_componentstatus_healthy{component="controller-manager"} 1
kube_componentstatus_healthy{component="etcd-0"} 1
kube_componentstatus_healthy{component="scheduler"} 1
```

# RBAC
Helm chart installs ServiceAccount, ClusterRole and ClusterRoleBinding to allow pod quering k8s `api/componentstatuses`

# Prometheus
[coreos prometheus operator](https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus) ServiceMonitor can be deployed or
metrics scraping can be configured via pod annotations.

To configure follow instructions in [values.yaml](./helm/values.yaml)

