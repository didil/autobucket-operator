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
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	abv1 "github.com/didil/autobucket-operator/api/v1"

	"github.com/didil/autobucket-operator/services"
)

// BucketReconciler reconciles a Bucket object
type BucketReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	GCPSvc services.GCPSvc
}

// +kubebuilder:rbac:groups=ab.leclouddev.com,resources=buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ab.leclouddev.com,resources=buckets/status,verbs=get;update;patch

func (r *BucketReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("bucket", req.NamespacedName)

	bucket := &abv1.Bucket{}
	err := r.Get(ctx, req.NamespacedName, bucket)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Bucket resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get bucket")
		return ctrl.Result{}, err
	}

	if bucket.Status.CreatedAt == "" {
		// bucket not yet created
		log.Info("Creating Bucket", "Bucket.Cloud", bucket.Spec.Cloud, "Bucket.Name", bucket.Name)

		switch bucket.Spec.Cloud {
		case abv1.BucketCloudGCP:
			err := r.createGCPBucket(ctx, bucket)
			if err != nil {
				log.Error(err, "Failed to create gcp Bucket", "Bucket.Name", bucket.Name)
				return ctrl.Result{}, err
			}
		default:
			log.Info("Bucket Cloud unknown.", "Bucket.Cloud", bucket.Spec.Cloud)
			return ctrl.Result{}, nil
		}

		bucket.Status.CreatedAt = time.Now().Format(time.RFC3339)
		err = r.Client.Status().Update(ctx, bucket)
		if err != nil {
			log.Error(err, "Failed to update bucket status")
			return ctrl.Result{}, err
		}

		// Status updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *BucketReconciler) createGCPBucket(ctx context.Context, bucket *abv1.Bucket) error {
	// create bucket
	err := r.GCPSvc.CreateBucket(ctx, bucket.Spec.FullName)
	if err != nil {
		return err
	}

	return nil
}

func (r *BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&abv1.Bucket{}).
		Complete(r)
}
