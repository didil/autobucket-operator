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

	// examine DeletionTimestamp to determine if object is under deletion
	if bucket.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.

		if !containsString(bucket.ObjectMeta.Finalizers, bucketFinalizerName) {
			bucket.ObjectMeta.Finalizers = append(bucket.ObjectMeta.Finalizers, bucketFinalizerName)
			if err := r.Update(ctx, bucket); err != nil {
				log.Error(err, "Failed to update bucket finalizers")
				return ctrl.Result{}, err
			}

			// Object updated - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	} else {
		// The object is being deleted
		if containsString(bucket.ObjectMeta.Finalizers, bucketFinalizerName) {
			// our finalizer is present, delete bucket
			if bucket.Spec.OnDeletePolicy == abv1.BucketOnDeletePolicyDestroy {
				log.Info("Deleting Storage Bucket", "Bucket.Cloud", bucket.Spec.Cloud, "Bucket.Name", bucket.Name)

				switch bucket.Spec.Cloud {
				case abv1.BucketCloudGCP:
					err := r.deleteGCPBucket(ctx, bucket)
					if err != nil {
						log.Error(err, "Failed to delete gcp Bucket", "Bucket.Name", bucket.Name)
						return ctrl.Result{}, err
					}
				default:
					log.Info("Bucket Cloud unknown.", "Bucket.Cloud", bucket.Spec.Cloud)
					return ctrl.Result{}, nil
				}
			}

			// remove our finalizer from the list and update it.
			bucket.ObjectMeta.Finalizers = removeString(bucket.ObjectMeta.Finalizers, bucketFinalizerName)
			if err := r.Update(context.Background(), bucket); err != nil {
				log.Error(err, "Failed to delete bucket finalizer")
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// check if the storage bucket has been created yet
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

const bucketFinalizerName = "ab.leclouddev.com/bucket-finalizer"

func (r *BucketReconciler) createGCPBucket(ctx context.Context, bucket *abv1.Bucket) error {
	// create bucket
	err := r.GCPSvc.CreateBucket(ctx, bucket.Spec.FullName)
	if err != nil {
		return err
	}

	return nil
}

func (r *BucketReconciler) deleteGCPBucket(ctx context.Context, bucket *abv1.Bucket) error {
	// delete bucket
	err := r.GCPSvc.DeleteGCPBucket(ctx, bucket.Spec.FullName)
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

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
