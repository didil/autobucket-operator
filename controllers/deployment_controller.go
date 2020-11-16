/*


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

	abv1 "github.com/didil/autobucket-operator/api/v1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
// +kubebuilder:rbac:groups=ab.leclouddev.com,resources=buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ab.leclouddev.com,resources=buckets/status,verbs=get;update;patch

func (r *DeploymentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("deployment", req.NamespacedName)

	dep := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Deployment resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get deployment")
		return ctrl.Result{}, err
	}

	bucketCloud := abv1.BucketCloud(dep.Annotations[bucketCloudKey])
	if bucketCloud == "" {
		// no autobucket annotation
		return ctrl.Result{}, nil
	}

	bucketNamePrefix := dep.Annotations[bucketNamePrefixKey]
	if bucketNamePrefix == "" {
		bucketNamePrefix = "ab"
	}

	bucketFullName := bucketFullName(bucketNamePrefix, req.Namespace, dep.Name)

	// Check if the bucket object already exists, if not create a new one
	bucket := &abv1.Bucket{}
	err = r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, bucket)
	if err != nil && errors.IsNotFound(err) {
		// Define new
		bucket, err := r.bucketForDeployment(dep, bucketCloud, bucketFullName)
		if err != nil {
			log.Error(err, "Failed to build new Bucket", "Bucket.Name", dep.Name)
			return ctrl.Result{}, err
		}

		log.Info("Creating a new Bucket", "Bucket.Name", bucket.Name)
		err = r.Create(ctx, bucket)
		if err != nil {
			log.Error(err, "Failed to create new Bucket", "Bucket.Name", bucket.Name)
			return ctrl.Result{}, err
		}

		// created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Bucket")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func bucketFullName(prefix, namespace, depName string) string {
	return prefix + "-" + namespace + "-" + depName
}

const bucketCloudKey = "ab.leclouddev.com/cloud"
const bucketNamePrefixKey = "ab.leclouddev.com/name-prefix"

// bucketForDeployment returns a Bucket object
func (r *DeploymentReconciler) bucketForDeployment(dep *appsv1.Deployment, bucketCloud abv1.BucketCloud, bucketFullName string) (*abv1.Bucket, error) {
	labels := labelsForBucket(dep.Name)

	bucket := &abv1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dep.Name,
			Namespace: dep.Namespace,
			Labels:    labels,
		},
		Spec: abv1.BucketSpec{
			Cloud:    bucketCloud,
			FullName: bucketFullName,
		},
	}
	// Set Project instance as the owner and controller
	err := ctrl.SetControllerReference(dep, bucket, r.Scheme)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}

// labelsForBucket returns the labels for a bucket
func labelsForBucket(deploymentName string) map[string]string {
	return map[string]string{"app": "ab", deploymentCRKey: deploymentName}
}

const deploymentCRKey = "deployment_cr"

func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Owns(&abv1.Bucket{}).
		Complete(r)
}
