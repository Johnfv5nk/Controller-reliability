package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	reliabilityv1alpha1 "github.com/Johnfv5nk/Controller-reliability/api/v1alpha1"
)

// ReliabilityReconciler reconciles a Reliability object
type ReliabilityReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=reliability.example.com,resources=reliabilities,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=reliability.example.com,resources=reliabilities/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=reliability.example.com,resources=reliabilities/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Reliability object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ReliabilityReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Reliability instance
	reliability := &reliabilityv1alpha1.Reliability{}
	err := r.Get(ctx, req.NamespacedName, reliability)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Reliability resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get Reliability")
		return ctrl.Result{}, err
	}

	// Update the status
	reliability.Status.ObservedGeneration = reliability.Generation
	reliability.Status.Phase = "Running"

	err = r.Status().Update(ctx, reliability)
	if err != nil {
		if errors.IsConflict(err) {
			logger.Info("Conflict detected during status update, requeuing for retry", "error", err.Error())
			return ctrl.Result{Requeue: true}, nil
		}
		logger.Error(err, "Failed to update Reliability status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReliabilityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&reliabilityv1alpha1.Reliability{}).
		Complete(r)
}
