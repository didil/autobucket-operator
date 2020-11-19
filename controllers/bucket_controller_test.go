package controllers

import (
	"context"
	"time"

	abv1 "github.com/didil/autobucket-operator/api/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Bucket controller", func() {
	const (
		NamespaceName  = "default"
		BucketName     = "test-bucket"
		BucketFullName = "ab-default-test-bucket"

		timeout  = time.Second * 5
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a bucket", func() {
		var bucket *abv1.Bucket

		It("Should create the storage bucket", func() {
			ctx := context.Background()

			gcpSvc.On("CreateBucket", mock.AnythingOfType("*context.emptyCtx"), BucketFullName).Return(nil)

			bucket = &abv1.Bucket{
				ObjectMeta: metav1.ObjectMeta{
					Name:      BucketName,
					Namespace: NamespaceName,
				},
				Spec: abv1.BucketSpec{
					Cloud:          abv1.BucketCloudGCP,
					FullName:       BucketFullName,
					OnDeletePolicy: abv1.BucketOnDeletePolicyIgnore,
				},
			}
			Expect(k8sClient.Create(ctx, bucket)).Should(Succeed())

			// check mock call
			Eventually(func() bool {
				calls := gcpSvc.Calls
				if len(calls) == 0 {
					return false
				}
				lastCall := calls[len(calls)-1]
				if lastCall.Method != "CreateBucket" {
					return false
				}
				if lastCall.Arguments[1].(string) != BucketFullName {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())

			// wait for bucket creation
			Eventually(func() bool {
				updatedBucket := &abv1.Bucket{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: bucket.Name, Namespace: bucket.Namespace}, updatedBucket)
				if err != nil {
					return false
				}

				ti, err := time.Parse(time.RFC3339, updatedBucket.Status.CreatedAt)
				if err != nil {
					return false
				}
				// check createdat timestamp is reasobable
				if ti.Before(time.Now().Add(-30*time.Second)) || ti.After(time.Now()) {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())

		})

		AfterEach(func() {
			ctx := context.Background()
			Expect(k8sClient.Delete(ctx, bucket)).Should(Succeed())
		})
	})

})
