package controller

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	reliabilityv1alpha1 "github.com/Johnfv5nk/Controller-reliability/api/v1alpha1"
)

var _ = Describe("Reliability Controller", func() {
	const (
		ReliabilityName      = "test-reliability"
		ReliabilityNamespace = "default"
		timeout              = time.Second * 10
		duration             = time.Second * 10
		interval             = time.Millisecond * 250
	)

	Context("When reconciling a resource", func() {
		It("Should successfully reconcile the resource", func() {
			ctx := context.Background()
			reliability := &reliabilityv1alpha1.Reliability{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "reliability.example.com/v1alpha1",
					Kind:       "Reliability",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ReliabilityName,
					Namespace: ReliabilityNamespace,
				},
				Spec: reliabilityv1alpha1.ReliabilitySpec{},
			}

			Expect(k8sClient.Create(ctx, reliability)).Should(Succeed())

			reliabilityLookupKey := types.NamespacedName{Name: ReliabilityName, Namespace: ReliabilityNamespace}
			createdReliability := &reliabilityv1alpha1.Reliability{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, reliabilityLookupKey, createdReliability)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdReliability.Spec).To(Equal(reliabilityv1alpha1.ReliabilitySpec{}))

			Eventually(func() string {
				err := k8sClient.Get(ctx, reliabilityLookupKey, createdReliability)
				if err != nil {
					return ""
				}
				return createdReliability.Status.Phase
			}, timeout, interval).Should(Equal("Running"))
		})
	})

	Context("When status update returns a conflict error", func() {
		It("Should requeue the request", func() {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			Expect(reliabilityv1alpha1.AddToScheme(scheme)).To(Succeed())

			reliability := &reliabilityv1alpha1.Reliability{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-conflict",
					Namespace: "default",
				},
			}

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(reliability).Build()

			conflictErr := apierrors.NewConflict(schema.GroupResource{Group: "reliability.example.com", Resource: "Reliability"}, "test-conflict", errors.New("conflict"))

			mockStatus := &mockStatusWriter{
				StatusWriter: fakeClient.Status(),
				updateErr:    conflictErr,
			}

			mockCl := &mockClient{
				Client:       fakeClient,
				statusWriter: mockStatus,
			}

			reconciler := &ReliabilityReconciler{
				Client: mockCl,
				Scheme: scheme,
			}

			result, err := reconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-conflict",
					Namespace: "default",
				},
			})

			Expect(err).To(BeNil())
			Expect(result.Requeue).To(BeTrue())
		})
	})

	Context("When status update returns a non-conflict error", func() {
		It("Should return the error", func() {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			Expect(reliabilityv1alpha1.AddToScheme(scheme)).To(Succeed())

			reliability := &reliabilityv1alpha1.Reliability{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-error",
					Namespace: "default",
				},
			}

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(reliability).Build()

			otherErr := errors.New("some other error")

			mockStatus := &mockStatusWriter{
				StatusWriter: fakeClient.Status(),
				updateErr:    otherErr,
			}

			mockCl := &mockClient{
				Client:       fakeClient,
				statusWriter: mockStatus,
			}

			reconciler := &ReliabilityReconciler{
				Client: mockCl,
				Scheme: scheme,
			}

			_, err := reconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-error",
					Namespace: "default",
				},
			})

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(otherErr))
		})
	})
})

type mockStatusWriter struct {
	client.StatusWriter
	updateErr error
}

func (m *mockStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	return m.updateErr
}

type mockClient struct {
	client.Client
	statusWriter client.StatusWriter
}

func (m *mockClient) Status() client.StatusWriter {
	return m.statusWriter
}
