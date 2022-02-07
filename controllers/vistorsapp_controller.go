/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	visitorsv1 "github.com/supreeth7/visitor-metrics-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// VistorsAppReconciler reconciles a VistorsApp object
type VistorsAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=visitors.example.com,resources=vistorsapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=visitors.example.com,resources=vistorsapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=visitors.example.com,resources=vistorsapps/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VistorsApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *VistorsAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instance := &visitorsv1.VistorsApp{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	// MySQL
	_, err = r.checkDeployment(req, instance, &appsv1.Deployment{})

	return ctrl.Result{}, nil
}

func (r *VistorsAppReconciler) checkDeployment(req ctrl.Request, app *visitorsv1.VistorsApp, deployment *appsv1.Deployment) (*ctrl.Result, error) {
	ctx := context.TODO()
	log := logr.FromContext(ctx)
	found := &appsv1.Deployment{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      deployment.Name,
			Namespace: app.Namespace,
		},
		found,
	)

	if err != nil && errors.IsNotFound(err) {
		// Create a new deployment
		log.Info("Creating new deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.Create(ctx, deployment)

		if err != nil {
			log.Error(err, "Failed creating a new deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		} else {
			log.Info("Success")
			return nil, nil
		}
	} else if err != nil {
		log.Error(err, "Failed creating a new deployment")
		return &ctrl.Result{}, err
	}

	return nil, nil
}

func (r *VistorsAppReconciler) checkService(req ctrl.Request, app *visitorsv1.VistorsApp, svc *corev1.Service) (*ctrl.Result, error) {
	ctx := context.TODO()
	log := logr.FromContext(ctx)
	found := &corev1.Service{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      svc.Name,
			Namespace: app.Namespace,
		},
		found,
	)

	if err != nil && errors.IsNotFound(err) {
		// Create new service
		log.Info("Creating new service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.Create(ctx, svc)

		if err != nil {
			log.Info("Service creation failed")
			return &ctrl.Result{}, nil
		} else {
			return nil, nil
		}
	} else if err != nil {
		log.Info("Service creation failed. Failed to get service.")
		return &ctrl.Result{}, nil
	}
	return nil, nil
}

func (r *VistorsAppReconciler) checkSecret(req ctrl.Request, app *visitorsv1.VistorsApp, secret *corev1.Secret) (*ctrl.Result, error) {
	ctx := context.TODO()
	log := logr.FromContext(ctx)
	found := &corev1.Secret{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      secret.Name,
			Namespace: app.Namespace,
		},
		found,
	)

	if err != nil && errors.IsNotFound(err) {
		// Create new service
		log.Info("Creating new secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.Create(ctx, secret)

		if err != nil {
			log.Info("Secret creation failed")
			return &ctrl.Result{}, nil
		} else {
			return nil, nil
		}
	} else if err != nil {
		log.Info("Secret creation failed. Failed to get secret.")
		return &ctrl.Result{}, nil
	}
	return nil, nil

}

func labels(v *visitorsv1.VistorsApp, tier string) map[string]string {
	return map[string]string{
		"app":             "visitors",
		"visitorssite_cr": v.Name,
		"tier":            tier,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *VistorsAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&visitorsv1.VistorsApp{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
