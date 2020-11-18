package controllers

import (
	"context"
	"fmt"
	"time"

	abv1 "github.com/didil/autobucket-operator/api/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Deployment controller", func() {
	const (
		NamespaceName  = "default"
		DeploymentName = "test-deployment"
		BucketName     = "test-deployment"
		BucketFullName = "abtest-default-test-deployment"

		timeout  = time.Second * 5
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a deployment", func() {
		var deployment *appsv1.Deployment
		var bucket *abv1.Bucket

		It("Should create the bucket crd", func() {
			ctx := context.Background()

			gcpSvc.On("CreateBucket", mock.AnythingOfType("*context.emptyCtx"), BucketFullName).Return(nil)

			deployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DeploymentName,
					Namespace: NamespaceName,
					Annotations: map[string]string{
						"ab.leclouddev.com/cloud":       "gcp",
						"ab.leclouddev.com/name-prefix": "abtest",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "test",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								corev1.Container{
									Name:  "test",
									Image: "busybox",
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, deployment)).Should(Succeed())

			// wait for bucket creation
			Eventually(func() error {
				bucket = &abv1.Bucket{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: DeploymentName, Namespace: deployment.Namespace}, bucket)
				if err != nil {
					return err
				}

				if cloud := bucket.Spec.Cloud; cloud != "gcp" {
					return fmt.Errorf("wrong cloud %v", cloud)
				}

				if fullName := bucket.Spec.FullName; fullName != BucketFullName {
					return fmt.Errorf("wrong full name %v", fullName)
				}

				return nil
			}, timeout, interval).Should(BeNil())

		})

		AfterEach(func() {
			ctx := context.Background()
			Expect(k8sClient.Delete(ctx, deployment)).Should(Succeed())
		})
	})

})
