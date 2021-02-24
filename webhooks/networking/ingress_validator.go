package networking

import (
	"context"
	"github.com/go-logr/logr"
	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/aws-load-balancer-controller/pkg/webhook"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const apiPathValidateNetworkingIngress = "/validate-networking-v1beta1-ingress"

// NewIngressValidator returns a validator for Ingress API.
func NewIngressValidator(logger logr.Logger) *ingressValidator {
	return &ingressValidator{
		logger: logger,
	}
}

var _ webhook.Validator = &ingressValidator{}

type ingressValidator struct {
	logger logr.Logger

	ingressClass                  string
	disableIngressClassAnnotation bool
	disableIngressGroupAnnotation bool
}

func (v *ingressValidator) Prototype(req admission.Request) (runtime.Object, error) {
	return &networking.Ingress{}, nil
}

func (v *ingressValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	ing := obj.(*networking.Ingress)
	_ = ing
	return nil
}

func (v *ingressValidator) ValidateUpdate(ctx context.Context, obj runtime.Object, oldObj runtime.Object) error {
	ing := obj.(*networking.Ingress)
	oldIng := oldObj.(*networking.Ingress)
	_ = ing
	_ = oldIng
	return nil
}

func (v *ingressValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}

// checkIngressClassAnnotationUsage checks the usage of kubernetes.io/ingress.class annotation.
// kubernetes.io/ingress.class annotation cannot be set to the ingress class for this controller once disabled,
// so that we enforce users to use spec.ingressClassName in Ingress and IngressClass resource instead.
func (v *ingressValidator) checkIngressClassAnnotationUsage() error {

}

// checkGroupNameAnnotationUsage checks the usage of "group.name" annotation.
// "group.name" annotation cannot be set once disabled,
// so that we enforce users to use spec.group in IngressClass resource instead.
func (v *ingressValidator) checkGroupNameAnnotationUsage() error {

}

// +kubebuilder:webhook:path=/validate-networking-v1beta1-ingress,mutating=false,failurePolicy=fail,groups=networking.k8s.io,resources=ingresses,verbs=create;update,versions=v1beta1,name=vingress.elbv2.k8s.aws,sideEffects=None,webhookVersions=v1beta1

func (v *ingressValidator) SetupWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(apiPathValidateNetworkingIngress, webhook.ValidatingWebhookForValidator(v))
}
