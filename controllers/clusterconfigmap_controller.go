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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	devhwv1 "github.com/juozasg/cluster-config-map/api/v1"
)

// ClusterConfigMapReconciler reconciles a ClusterConfigMap object
type ClusterConfigMapReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devhw.github.com,resources=clusterconfigmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devhw.github.com,resources=clusterconfigmaps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devhw.github.com,resources=clusterconfigmaps/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterConfigMap object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *ClusterConfigMapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var ccm devhwv1.ClusterConfigMap
	if err := r.Get(ctx, req.NamespacedName, &ccm); err != nil {
		log.Error(err, "unable to fetch ClusterConfigMap")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	selectors := ccm.Spec.GenerateTo.NamespaceSelectors
	matchLabels0 := selectors[0].MatchLabels
	nsFilter := client.MatchingLabels(matchLabels0)

	var namespaces corev1.NamespaceList

	if err := r.List(ctx, &namespaces, nsFilter); err != nil {
		log.Error(err, "unable to list namespaces")
		return ctrl.Result{}, err
	}

	fmt.Println("namespaces", namespaces)

	// get all namespaces matching labels

	// get all configmaps owned by this resource
	// if not in labeled namespaces, delete the config map

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterConfigMapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devhwv1.ClusterConfigMap{}).
		Complete(r)

	// (watch namespaces for creation)
	//.Watches
	// https://master.book.kubebuilder.io/reference/watching-resources/externally-managed.html
}
