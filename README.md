# operator-utils library
[![Go Report Card](https://goreportcard.com/badge/github.com/RHsyseng/operator-utils)](https://goreportcard.com/report/github.com/RHsyseng/operator-utils)
[![Build Status](https://travis-ci.org/RHsyseng/operator-utils.svg?branch=master)](https://travis-ci.org/RHsyseng/operator-utils)


This library layers on top of the Operator SDK, having set of utilities function as a library to easily create Kubernetes operators.

## Kubernetes / OpenShift Version Support

In July of 2020, our team [moved away from using the term `master`](https://www.redhat.com/en/blog/making-open-source-more-inclusive-eradicating-problematic-language) for our default branch. As a result, our branching scheme is as follows:
- The [main](https://github.com/RHsyseng/operator-utils/tree/main) (default) branch currently supports **OCP 4.6** (K8S 1.19), the latest GA release.
- The [next](https://github.com/RHsyseng/operator-utils/tree/next) branch currently supports **OCP 4.7** (K8S 1.20).
- For versions of `operator-utils` targeting earlier releases of OCP (starting with 4.2), please refer to the [tags](https://github.com/RHsyseng/operator-utils/tags) section.
  - tag `v1.X.Y` indicates support for OCP `vX.Y`
- With each General Availability release of OCP, the `main` branch will be given a tag matching the previously supported OCP version in `main`, then the `next` branch will be rebased onto `main`.


## Contributing to the `operator-utils` Project

All bugs, tasks, fixes or enhancements should be tracked as [GitHub Issues](https://github.com/RHsyseng/operator-utils/issues) & [Pull Requests](https://github.com/RHsyseng/operator-utils/pulls).

- To contribute features targeting **OCP 4.6** only, use a local feature branch based off of & targeting `origin/main` with any PR's. Reference any JIRA/GitHub issues in PR's where applicable.
- To contribute features targeting **OCP 4.7** only, use a local feature branch based off of & targeting `origin/next` with any PR's, Reference any JIRA/GitHub issues in PR's where applicable.
- To contribute features targeting **both currently supported versions**, first complete the commit/PR work targeting `next`. Once that PR is merged to `next`, create a new PR with cherry-pick of the commit targeting `main`.
- Contributions targeting OCP versions older than what's currently supported by `main` will typically no longer be accepted. Please contact contributors for further discussion. 

## Declaring operator-utils dependency


Regardless of dependency framework, we suggest following the best practice of declaring any and all dependencies your project utilizes regardless of target branch, tag, or revision. 

With regards to `operator-utils`, please **carefully** consider the given version support information above when declaring your dependency, as depending on or defaulting to `main` branch will likely result in future build complications as our project continues to evolve and cycle minor version support.   

- Go.mod example specifying **REVISION**:
```
github.com/RHsyseng/operator-utils v0.0.0-20200108204558-82090ef57586
```


## Features

1. [managing CR and CRD validation](#managing-cr-and-crd-validation)
2. [pods deployment status ](#pods-deployment-status)
3. [resource comparison, adding, updating and deleting](#resource-comparison-adding-updating-and-deleting)
4. [platform detection Kubernetes VS Openshift](#platform-detection-kubernetes-vs-openshift)


## Managing CR and CRD validation

Operator util library use package ``validation`` for validate the CRD  and CR file, these function use as a unit test within operator

**CRD validation Usage**:
```go

schema := getCompleteSchema(t)
 missingEntries := schema.GetMissingEntries(&sampleApp{})
 for _, missing := range missingEntries {
    if strings.HasPrefix(missing.Path, "/status") {
       //Not using subresources, so status is not expected to appear in CRD
    } else {
       assert.Fail(t, "Discrepancy between CRD and Struct", "Missing or incorrect schema validation at %v, expected type %v", missing.Path, missing.Type)
    }
 }


```


**CR validation Usage**:
```go
schema, err := New([]byte(schemaYaml))
 assert.NoError(t, err)

 type myAppSpec struct {
    Number float64 `json:"number,omitempty"`
 }

 type myApp struct {
    Spec myAppSpec `json:"spec,omitempty"`
 }

 cr := myApp{
    Spec: myAppSpec{
       Number: float64(23),
    },
 }
 missingEntries := schema.GetMissingEntries(&cr)
 assert.Len(t, missingEntries, 0, "Expect no missing entries in CRD for this struct: %v", missingEntries)
```
A full example is provided [here](./pkg/validation/schema_sync_test.go)


## Pods deployment status

 showes the status of the deployment on OLM UI in the form of PI chart, as seen in below screenshot

![alt text](deployment-satus.png "pods PI chart")

**Usage**:

Below seen line required to add into types.go status structure

```go
PodStatus olm.DeploymentStatus `json:"podStatus"`

```
Add these lines into CSV file inside statusDescriptors section:

```yaml
statusDescriptors:
        - description: The current pods
          displayName: Pods Status
          path: podStatus
          x-descriptors:
            - "urn:alm:descriptor:com.tectonic.ui:podStatuses"
```

For DeploymentConfig deployment status:

```go
var dcs []oappsv1.DeploymentConfig

deploymentStatus := olm.GetDeploymentConfigStatus(dcs)
 if !reflect.DeepEqual(instance.Status.Deployments, deploymentStatus) {
    r.reqLogger.Info("Deployment status will be updated")
    instance.Status.Deployments = deploymentStatus
    err = r.client.Status().Update(context.TODO(), instance)
    if err != nil {
       r.reqLogger.Error(err, "Failed to update deployment status")
       return err
    }
 }

```

For StatefulSet Deployment status:

```go
var status olm.DeploymentStatus
 sfsFound := &appsv1.StatefulSet{}

 err := client.Get(context.TODO(), namespacedName, sfsFound)
 if err == nil {
    status = olm.GetSingleStatefulSetStatus(*sfsFound)
 } else {
    dsFound := &appsv1.DaemonSet{}
    err = client.Get(context.TODO(), namespacedName, dsFound)
    if err == nil {
       status = olm.GetSingleDaemonSetStatus(*dsFound)
    }
 }

```

## Resource comparison (adding, updating and deleting)

Common function for listing, adding, updating, deleting kubernetes objects like seen below:


List of objects that are deployed

```go
reader := read.New(client).WithNamespace(instance.Namespace).WithOwnerObject(instance)
  resourceMap, err := reader.ListAll(
     &corev1.PersistentVolumeClaimList{},
     &corev1.ServiceList{},
     &appsv1.StatefulSetList{},
     &routev1.RouteList{},
  )
```



Compare what's deployed with what should be deployed

```go
requested := compare.NewMapBuilder().Add(requestedResources...).ResourceMap()
comparator := compare.NewMapComparator()
deltas := comparator.Compare(deployed, requested)
```

Adding the objects:

```go

added, err := writer.AddResources(delta.Added)

```

Updating the objects:

```go
updated, err := writer.UpdateResources(deployed[resourceType], delta.Updated)
```

Removing the objects:
```go
removed, err := writer.RemoveResources(delta.Removed)
```


A full usage is provided [here]( https://github.com/kiegroup/kie-cloud-operator/blob/6964179113e4f57d47bead03578ae6ed8e9caa8b/pkg/controller/kieapp/kieapp_controller.go#L136-L163)




## Platform detection Kubernetes VS Openshift

To detect platform whether operator is running on kuberenete or openshift  or what version of openshift is using

```go
  info, err := pv.GetPlatformInfo(c.discoverer, c.config)
```
A full example is provided [here](./internal/platform/platform_versioner_test.go)

## Who is using this Library

operator-utils is used by several Red Hat product & community operators, including the following:  

- [3scale APIcast Operator](https://github.com/3scale/apicast-operator)
- [3scale Operator](https://github.com/3scale/3scale-operator)
- [ActiveMQ Artemis Operator](https://github.com/rh-messaging/activemq-artemis-operator)
- [ActiveMQ Artemis Broker Test Suite](https://github.com/artemiscloud/openshift-broker-test-suite)
- [AtlasMap Operator](https://github.com/atlasmap/atlasmap-operator)
- [Barometer Operator](https://github.com/aneeshkp/barometer-operator)
- [Integreatly Operator](https://github.com/integr8ly/integreatly-operator)
- [Kie Cloud Operator](https://github.com/kiegroup/kie-cloud-operator)
- [Kogito Operator](https://github.com/kiegroup/kogito-cloud-operator)
- [KubeDB Operator](https://github.com/mrhillsman/kubedb-operator)
- [KubeVirt Containerized Data Importer](https://github.com/kubevirt/containerized-data-importer)
- [KubeVirt Hyperconverged Operator](https://github.com/kubevirt/hyperconverged-cluster-operator)
- [Nexus Operator](https://github.com/m88i/nexus-operator)
- [OCS Meta Operator](https://github.com/openshift/ocs-operator)
- [Performance Addon Operator](https://github.com/openshift-kni/performance-addon-operators)
- [QDR Interconnect Operator](https://github.com/interconnectedcloud/qdr-operator)
- [RHI Operator](https://github.com/redhat-integration/rhi-operator)
- [Serverless Orchestration](https://github.com/RHsyseng/serverless-orchestration)
- [ShipShape Testing Operator](https://github.com/rh-messaging/shipshape)
- [Teiid Operator](https://github.com/teiid/teiid-operator)
- [Wildfly Operator](https://github.com/wildfly/wildfly-operator)

