# Kshow

Kshow is a CLI tool for quick access to k8s objects with custom fields.

## Project Owners
- Samarth Kanungo
- Open to Community

## OS Supported
- Mac (Darwin-amd64)
- Linux (amd64)
- Windows (x64)

## Installation
- Download the binary from releases.
- Rename it to kshow.
```
mv kshow-0.0.1-amd64-darwin kshow
```
- Put it under the Path.
```
mv kshow /usr/local/bin/kshow
```

## Usage

### Deployments

#### List Deployments

namespace is optional. Default is all namespace
```
kshow get deployments --namespace <NAMESPACE>

DEPLOYMENT         NAMESPACE REPLICAS    
app-db-live        app-server   2        
app-ui-live        app-server   1    
app-backend-live   app-server   2 
```

#### List Deployments with Tolerations

```
kshow get deployments --namespace <NAMESPACE> --show-tolerations

DEPLOYMENT         NAMESPACE REPLICAS    TOLERATIONS
app-db-live        app-server   2        nature-Equal-stateful-NoSchedule
app-ui-live        app-server   1        nature-Equal-spot-NoSchedule
app-backend-live   app-server   2 
```

### Pods

#### List Pods
```
kshow get pods --namespace <NAMESPACE>

POD                                    READY  STATUS     RESTART  AGE   NAMESPACE
app-db-live-54c8d4897f-clfln           1/1    Running    0        2d    app-server
app-ui-live-54c8d4897f-glzrz           1/1    Running    0        6d    app-server
app-backend-live-65b4d7fd57-9gcz8      1/1    Running    0        2d    app-server
```

#### List Pods with Node Tenancy (only for AWS EKS)
*NOTE: feature only available for AWS EKS*
This feature is to determine the type of Node (On-Demand / SPOT) on which pod is scheduled.

```
kshow get pods  --namespace <NAMESPACE> --show-tenancy

POD                                   STATUS     NAMESPACE    NODE                                         TENANCY
app-db-live-54c8d4897f-clfln          Running    app-server   ip-172-28-87-236.eu-west-1.compute.internal  SPOT
app-ui-live-54c8d4897f-glzrz          Running    app-server   ip-172-28-83-25.eu-west-1.compute.internal   SPOT
app-backend-live-65b4d7fd57-9gcz8     Running    app-server   ip-172-28-6-173.eu-west-1.compute.internal   ON_DEMAND
```

### Metrics

#### Get Metrics

This feature shows current CPU and Memory consumption of pods.

```
kshow resource-stats --namespace kube-system

NAMESPACE    POD                                  CPU   MEMORY
app-server   app-db-live-54c8d4897f-clfln         3m    822Mi
app-server   app-ui-live-54c8d4897f-glzrz         2m    918Mi
app-server   app-backend-live-65b4d7fd57-9gcz8    23m   1457Mi
```

#### Get Container Metrics
```
kshow resource-stats --namespace kube-system --containers

NAMESPACE    POD                              	 CONTAINER    CPU   MEMORY
app-server   app-db-live-54c8d4897f-clfln        app-db-live  3m    822Mi
app-server   app-ui-live-54c8d4897f-glzrz        app-ui-live  2m    918Mi
app-server   app-backend-live-65b4d7fd57-9gcz8   app-backend  23m   1457Mi
```