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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	devhwv1 "github.com/juozasg/cluster-config-map/api/v1"
)

// ClusterConfigMapReconciler reconciles a ClusterConfigMap object
type ClusterConfigMapReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *ClusterConfigMapReconciler) ListMatchingNamespaces(ctx context.Context,
	generateTo devhwv1.GenerateToSpec,
	namespaces *[]corev1.Namespace) error {
	log := log.FromContext(ctx)

	for _, selector := range generateTo.NamespaceSelectors {
		nsFilter := client.MatchingLabels(selector.MatchLabels)

		var nsList corev1.NamespaceList

		if err := r.List(ctx, &nsList, nsFilter); err != nil {
			log.Error(err, "unable to list namespaces")
			return err
		}

		*namespaces = append(*namespaces, nsList.Items...)
	}

	return nil
}

func ContainsNamespace(nss []corev1.Namespace, name string) bool {
	for _, ns := range nss {
		if ns.Name == name {
			return true
		}
	}

	return false
}

func IsOwner(owner devhwv1.ClusterConfigMap, cm corev1.ConfigMap) bool {
	for _, ownerRef := range cm.OwnerReferences {
		if ownerRef.APIVersion == owner.APIVersion &&
			ownerRef.Kind == owner.Kind &&
			ownerRef.Name == owner.Name {
			return true
		}
	}

	return false
}

//+kubebuilder:rbac:groups=devhw.github.com,resources=clusterconfigmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devhw.github.com,resources=clusterconfigmaps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devhw.github.com,resources=clusterconfigmaps/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;

func (r *ClusterConfigMapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var ccm devhwv1.ClusterConfigMap
	if err := r.Get(ctx, req.NamespacedName, &ccm); err != nil {
		if errors.IsNotFound(err) {
			log.Info(err.Error())
			return ctrl.Result{}, client.IgnoreNotFound(err)
		} else {
			log.Error(err, "unable to fetch ClusterConfigMap")
			return ctrl.Result{}, err
		}
	}

	var matchedNamespaces []corev1.Namespace
	if err := r.ListMatchingNamespaces(ctx, ccm.Spec.GenerateTo, &matchedNamespaces); err != nil {
		return ctrl.Result{}, err
	}

	// reconcile ConfigMap for each namepace
	for _, ns := range matchedNamespaces {
		var configMap corev1.ConfigMap
		err := r.Get(ctx, types.NamespacedName{Namespace: ns.Name, Name: ccm.Name}, &configMap)

		if err == nil {
			// ConfigMap exists. Update data if needed
			if !reflect.DeepEqual(configMap.Data, ccm.Spec.Data) {
				configMap.Data = ccm.Spec.Data
				log.Info("Updating ConfigMap " + configMap.Namespace + "/" + ccm.Name)
				err = r.Update(ctx, &configMap)
				if err != nil {
					return ctrl.Result{}, err
				}
			}
		} else if err != nil && errors.IsNotFound(err) {
			// ConfigMap does not exist. Create
			configMap.Namespace = ns.Name
			configMap.Name = ccm.Name
			configMap.Data = ccm.Spec.Data

			log.Info("Creating ConfigMap " + configMap.Namespace + "/" + ccm.Name)
			// set owner ref
			err = controllerutil.SetOwnerReference(&ccm, &configMap, r.Scheme)
			if err != nil {
				log.Error(err, "Failed to set ConfigMap owner reference")
				return ctrl.Result{}, err
			}

			err = r.Create(ctx, &configMap)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else {
			// give up on other errors
			return ctrl.Result{}, err
		}
	}

	// TODO: use indexing or selectable labels to optimize this
	var allConfigMaps corev1.ConfigMapList
	err := r.List(ctx, &allConfigMaps)
	if err != nil {
		log.Error(err, "unable to list ConfigMaps ")
		return ctrl.Result{}, nil
	}

	for _, cm := range allConfigMaps.Items {
		if IsOwner(ccm, cm) && !ContainsNamespace(matchedNamespaces, cm.Namespace) {
			// cm does not belong to a namespace matching ccm labels
			// must be deleted
			log.Info("Deleting ConfigMap " + cm.Namespace + "/" + cm.Name + " (no matching namespaceSelector)")
			r.Delete(ctx, &cm)
		}

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterConfigMapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devhwv1.ClusterConfigMap{}).
		Owns(&corev1.ConfigMap{}).
		Watches(&source.Kind{Type: &corev1.Namespace{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForNamespace),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}

func (r *ClusterConfigMapReconciler) findObjectsForNamespace(namespace client.Object) []reconcile.Request {
	var ccmList devhwv1.ClusterConfigMapList

	if err := r.List(context.TODO(), &ccmList); err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(ccmList.Items))
	for i, item := range ccmList.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}

	return requests
}
