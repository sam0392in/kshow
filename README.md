# Kshow
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/gomods/athens.svg)](https://github.com/gomods/athens) ![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/sam0392in/kshow?include_prereleases)
[![Go Report Card](https://goreportcard.com/badge/github.com/sam0392in/kshow)](https://goreportcard.com/report/github.com/sam0392in/kshow)


Kshow is a advanced Kubernetes CLI tool for quick access to k8s objects with custom commands.
This project includes advanced features such as ```--show-tolerations``` , ```--detailed``` , ```resource-stats``` which existing kubernetes cli like kubectl lacks to provide.

## Current Version
- Stable Release: 0.0.4

## Project Owners
- Samarth Kanungo
- Kubernetes Community

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
- Place under the Path.
```
mv kshow /usr/local/bin/kshow
```

## Usage

### Deployments

#### **List Deployments**

namespace is optional. Default is all namespace
```
kshow get deployments -n <NAMESPACE>

DEPLOYMENT         NAMESPACE REPLICAS    
app-db-live        app-server   2        
app-ui-live        app-server   1    
app-backend-live   app-server   2 
```

#### **List Deployments with Tolerations**

```
kshow get deployments -n <NAMESPACE> --detailed

DEPLOYMENT         NAMESPACE  READY DISTRIBUTION    TOLERATIONS
app-db-live        app-server  3/3   OD:0 SP:3      nature-Equal-ondemand-NoSchedule
app-ui-live        app-server  2/2   OD:2 SP:0      nature-Equal-spot-NoSchedule
app-backend-live   app-server  0/0   OD:0 SP:0
```

### Pods

#### **List Pods**
```
kshow get pods -n <NAMESPACE>

POD                                    READY  STATUS     RESTART  AGE   NAMESPACE
app-db-live-54c8d4897f-clfln           1/1    Running    0        2d    app-server
app-ui-live-54c8d4897f-glzrz           1/1    Running    0        6d    app-server
app-backend-live-65b4d7fd57-9gcz8      1/1    Running    0        2d    app-server
```

#### **List Pods with Details**
*NOTE: feature only available for AWS EKS*
This feature is to determine the type of Node (On-Demand / SPOT) on which pod is scheduled.

```
kshow get pods  --namespace <NAMESPACE> --detailed

POD                                   STATUS     NAMESPACE    NODE                                         TENANCY
app-db-live-54c8d4897f-clfln          Running    app-server   ip-172-28-87-236.eu-west-1.compute.internal  SPOT
app-ui-live-54c8d4897f-glzrz          Running    app-server   ip-172-28-83-25.eu-west-1.compute.internal   SPOT
app-backend-live-65b4d7fd57-9gcz8     Running    app-server   ip-172-28-6-173.eu-west-1.compute.internal   ON_DEMAND
```

### Nodes

#### **List Nodes**
```
kshow get nodes

NODE                                         STATUS  AGE   VERSION
ip-172-27-0-105.eu-west-1.compute.internal   Ready   7d    v1.21.5-eks-9017834
ip-172-23-0-179.eu-west-1.compute.internal   Ready   24d   v1.21.5-eks-9017834
ip-172-21-0-216.eu-west-1.compute.internal   Ready   3d    v1.21.5-eks-9017834
```

#### **List Nodes with Details**

*NOTE: feature only available for AWS EKS*

```
kshow get nodes --detailed
-----------------------------------------------------------------------------------------------------
K8S-VERSION			NODE-GROUP: NODECOUNT
v1.21.5-eks-9017834		eks-on-demand:  1
				eks-spot:  9
------------------------------------------------------------------------------------------------------

NODE                                         STATUS  AGE   NODEGROUP      TENANCY    INSTANCE-TYPE  ARCH   AWS-ZONE    
ip-172-24-0-205.eu-west-1.compute.internal   Ready   7d    eks-spot       SPOT       m4.xlarge      amd64  eu-west-1a 
ip-172-21-0-379.eu-west-1.compute.internal   Ready   24d   eks-on-demand  ON_DEMAND  m5a.xlarge     amd64  eu-west-1c
ip-172-23-0-243.eu-west-1.compute.internal   Ready   3d    eks-spot       SPOT       m4.xlarge      amd64  eu-west-1b
```


### Metrics

#### **Get Metrics**

This feature shows current CPU and Memory consumption of pods.

```
kshow resource-stats -n <NAMESPACE>

NAMESPACE    POD                                  CPU   MEMORY
app-server   app-db-live-54c8d4897f-clfln         3m    822Mi
app-server   app-ui-live-54c8d4897f-glzrz         2m    918Mi
app-server   app-backend-live-65b4d7fd57-9gcz8    23m   1457Mi
```

#### **Get Detailed Metrics**
```
kshow resource-stats -n <NAMESPACE> --detailed

-------------------------------------------------------------------------------------------------------
Cluster Stats: 		Total CPU: 736 Cores		Total Memory: 1389 GB
Namespace Stats: 	Consumed CPU: 0.9 Cores		Consumed Memory: 2.1 GB
% Stats: 	        CPU: 0.12 %		        Memory: 0.15 %
-------------------------------------------------------------------------------------------------------

NAMESPACE  	  POD                          CONTAINER    CURRENT-CPU REQ-CPU   LIMIT-CPU  CURRENT-MEM  REQ-MEM  LIMIT-MEM
app-server    app-db-live-54c8d4897f-clfln     app-db       2m          300m      500m       313Mi        512Mi    768Mi
app-server    app-ui-live-54c8d4897f-glzrz     app-ui       2m          300m      500m       507Mi        12Mi     768Mi
app-server    app-backend-live-65bfd57-9gcz8   app-backend  6m          300m      500m       1005Mi       1600Mi   1600Mi
```

#### **Get Deployment Metrics** [Alpha Feature]

Below command shows cummilative CPU and Memory of all the replicas in a deployment.

```
kshow resource-stats deployments -n <NAMESPACE>

DEPLOYMENT         NAMESPACE   REQ-CPU CURRENT-CPU  REQ-MEM CURRENT-MEM
app-db-live        app-server   1800m   9m           2304Mi  2467Mi        
app-ui-live        app-server   2000m   32m          3200Mi  2885Mi   
app-backend-live   app-server   1000m   38m          1536Mi  1594Mi
```
