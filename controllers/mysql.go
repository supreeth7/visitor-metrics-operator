package controllers

import (
	"context"
	"fmt"

	"github.com/apex/log"
	visitorsv1 "github.com/supreeth7/visitor-metrics-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func mysqlDeploymentName() string {
	return "mysql"
}

func mysqlServiceName() string {
	return "mysql-service"
}

func mysqlAuthName() string {
	return "mysql-auth"
}

func (r *VistorsAppReconciler) mysqlAuthSecret(v *visitorsv1.VistorsApp) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysqlAuthName(),
			Namespace: v.Namespace,
		},
		Type: "Opaque",
		StringData: map[string]string{
			"username": "visitors-user",
			"password": "visitors-pass",
		},
	}
	controllerutil.SetControllerReference(v, secret, r.Scheme)
	return secret
}

func (r *VistorsAppReconciler) mysqlDeployment(v *visitorsv1.VistorsApp) *appsv1.Deployment {
	labels := labels(v, "mysql")
	size := int32(1)

	userSecret := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: mysqlAuthName()},
			Key:                  "username",
		},
	}

	passwordSecret := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: mysqlAuthName()},
			Key:                  "password",
		},
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysqlDeploymentName(),
			Namespace: v.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "mysql:5.7",
						Name:  "visitors-mysql",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 3306,
							Name:          "mysql",
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "MYSQL_ROOT_PASSWORD",
								Value: "password",
							},
							{
								Name:  "MYSQL_DATABASE",
								Value: "visitors",
							},
							{
								Name:      "MYSQL_USER",
								ValueFrom: userSecret,
							},
							{
								Name:      "MYSQL_PASSWORD",
								ValueFrom: passwordSecret,
							},
						},
					}},
				},
			},
		},
	}

	controllerutil.SetControllerReference(v, dep, r.Scheme)
	return dep
}

func (r *VistorsAppReconciler) mysqlService(v *visitorsv1.VistorsApp) *corev1.Service {
	labels := labels(v, "mysql")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysqlServiceName(),
			Namespace: v.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Port: 3306,
			}},
			ClusterIP: "None",
		},
	}

	controllerutil.SetControllerReference(v, s, r.Scheme)
	return s
}

// Returns whether or not the MySQL deployment is running
func (r *VistorsAppReconciler) isMysqlUp(v *visitorsv1.VistorsApp) bool {
	deployment := &appsv1.Deployment{}

	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      mysqlDeploymentName(),
		Namespace: v.Namespace,
	}, deployment)

	if err != nil {
		log.Error(fmt.Sprintf("Deployment mysql not found: %s", err))
		return false
	}

	if deployment.Status.ReadyReplicas == 1 {
		return true
	}

	return false
}
