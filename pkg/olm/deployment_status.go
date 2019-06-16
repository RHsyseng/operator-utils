package olm

import (
	oappsv1 "github.com/openshift/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("olm")

func GetDaemonSetStatus(dcs []appsv1.DaemonSet) DeploymentStatus {
	return getDeploymentStatus(deployments{
		countFunc: func() int {
			return len(dcs)
		},
		nameFunc: func(i int) string {
			return dcs[i].Name
		},
		specReplicasFunc: func(i int) int32 {
			//DaemonSet means one replica per node, so more than zero requested and can return 1:
			return 1
		},
		statusReplicasFunc: func(i int) int32 {
			return dcs[i].Status.DesiredNumberScheduled
		},
		readyReplicasFunc: func(i int) int32 {
			return dcs[i].Status.NumberReady
		},
	})
}

func GetDeploymentStatus(dcs []appsv1.Deployment) DeploymentStatus {
	return getDeploymentStatus(deployments{
		countFunc: func() int {
			return len(dcs)
		},
		nameFunc: func(i int) string {
			return dcs[i].Name
		},
		specReplicasFunc: func(i int) int32 {
			intPtr := dcs[i].Spec.Replicas
			if intPtr == nil {
				return 0
			} else {
				return *intPtr
			}
		},
		statusReplicasFunc: func(i int) int32 {
			return dcs[i].Status.Replicas
		},
		readyReplicasFunc: func(i int) int32 {
			return dcs[i].Status.ReadyReplicas
		},
	})
}

func GetDeploymentConfigStatus(dcs []oappsv1.DeploymentConfig) DeploymentStatus {
	return getDeploymentStatus(deployments{
		countFunc: func() int {
			return len(dcs)
		},
		nameFunc: func(i int) string {
			return dcs[i].Name
		},
		specReplicasFunc: func(i int) int32 {
			return dcs[i].Spec.Replicas
		},
		statusReplicasFunc: func(i int) int32 {
			return dcs[i].Status.Replicas
		},
		readyReplicasFunc: func(i int) int32 {
			return dcs[i].Status.ReadyReplicas
		},
	})
}

func getDeploymentStatus(obj deployments) DeploymentStatus {
	var ready, starting, stopped []string
	for i := 0; i < obj.count(); i++ {
		if obj.specReplicas(i) == 0 {
			stopped = append(stopped, obj.name(i))
		} else if obj.statusReplicas(i) == 0 {
			stopped = append(stopped, obj.name(i))
		} else if obj.readyReplicas(i) < obj.statusReplicas(i) {
			starting = append(starting, obj.name(i))
		} else {
			ready = append(ready, obj.name(i))
		}
	}
	log.Info("Found deployments with status ", "stopped", stopped, "starting", starting, "ready", ready)
	return DeploymentStatus{
		Stopped:  stopped,
		Starting: starting,
		Ready:    ready,
	}

}
