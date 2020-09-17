package service

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/aws-alb-ingress-controller/controllers/service/eventhandlers"
	"sigs.k8s.io/aws-alb-ingress-controller/pkg/annotations"
	"sigs.k8s.io/aws-alb-ingress-controller/pkg/aws/services"
	"sigs.k8s.io/aws-alb-ingress-controller/pkg/deploy"
	"sigs.k8s.io/aws-alb-ingress-controller/pkg/k8s"
	"sigs.k8s.io/aws-alb-ingress-controller/pkg/model/core"
	elbv2model "sigs.k8s.io/aws-alb-ingress-controller/pkg/model/elbv2"
	"sigs.k8s.io/aws-alb-ingress-controller/pkg/networking"
	"sigs.k8s.io/aws-alb-ingress-controller/pkg/runtime"
	"sigs.k8s.io/aws-alb-ingress-controller/pkg/service/nlb"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	LoadBalancerFinalizer   = "service.k8s.aws/load-balancer-finalizer"
	DefaultTagPrefix        = "service.k8s.aws"
	ServiceAnnotationPrefix = "service.beta.kubernetes.io"
	controllerName          = "service"
)

func NewServiceReconciler(k8sClient client.Client, ec2Client services.EC2, elbv2Client services.ELBV2,
	sgManager networking.SecurityGroupManager, sgReconciler networking.SecurityGroupReconciler,
	vpcID string, clusterName string, resolver networking.SubnetsResolver, logger logr.Logger) *ServiceReconciler {
	return &ServiceReconciler{
		k8sClient:        k8sClient,
		logger:           logger,
		annotationParser: annotations.NewSuffixAnnotationParser(ServiceAnnotationPrefix),
		finalizerManager: k8s.NewDefaultFinalizerManager(k8sClient, logger),
		subnetsResolver:  resolver,
		stackMarshaller:  deploy.NewDefaultStackMarshaller(),
		stackDeployer:    deploy.NewDefaultStackDeployer(k8sClient, ec2Client, elbv2Client, sgManager, sgReconciler, vpcID, clusterName, DefaultTagPrefix, logger),
	}
}

type ServiceReconciler struct {
	k8sClient        client.Client
	logger           logr.Logger
	annotationParser annotations.Parser
	finalizerManager k8s.FinalizerManager
	subnetsResolver  networking.SubnetsResolver

	stackMarshaller deploy.StackMarshaller
	stackDeployer   deploy.StackDeployer
}

// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch
func (r *ServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return runtime.HandleReconcileError(r.reconcile(req), r.logger)
}

func (r *ServiceReconciler) reconcile(req ctrl.Request) error {
	ctx := context.Background()
	svc := &corev1.Service{}
	if err := r.k8sClient.Get(ctx, req.NamespacedName, svc); err != nil {
		return client.IgnoreNotFound(err)
	}
	if !svc.DeletionTimestamp.IsZero() {
		return r.cleanupLoadBalancerResources(ctx, svc)
	}
	return r.reconcileLoadBalancerResources(ctx, svc)
}

func (r *ServiceReconciler) buildAndDeployModel(ctx context.Context, svc *corev1.Service) (core.Stack, *elbv2model.LoadBalancer, error) {
	nlbBuilder := nlb.NewServiceBuilder(svc, r.subnetsResolver, r.annotationParser)
	stack, lb, err := nlbBuilder.Build(ctx)
	if err != nil {
		return nil, nil, err
	}

	jsonString, err := r.stackMarshaller.Marshal(stack)
	r.logger.Info("Built service model", "stack", jsonString)

	err = r.stackDeployer.Deploy(ctx, stack)
	if err != nil {
		return nil, nil, err
	}
	r.logger.Info("Successfully deployed service resources")
	return stack, lb, nil
}

func (r *ServiceReconciler) reconcileLoadBalancerResources(ctx context.Context, svc *corev1.Service) error {
	if err := r.finalizerManager.AddFinalizers(ctx, svc, LoadBalancerFinalizer); err != nil {
		return err
	}
	stack, lb, err := r.buildAndDeployModel(ctx, svc)
	if err != nil {
		return err
	}
	dnsName, _ := lb.DNSName().Resolve(ctx)
	err = r.updateServiceStatus(ctx, svc, dnsName)
	if err != nil {
		return err
	}
	var resTGs []*elbv2model.TargetGroup
	stack.ListResources(&resTGs)
	r.logger.Info("Deployed LoadBalancer", "dnsname", dnsName, "#target groups", len(resTGs))
	return nil
}

func (r *ServiceReconciler) cleanupLoadBalancerResources(ctx context.Context, svc *corev1.Service) error {
	if k8s.HasFinalizer(svc, LoadBalancerFinalizer) {
		_, _, err := r.buildAndDeployModel(ctx, svc)
		if err != nil {
			return err
		}
		if err := r.finalizerManager.RemoveFinalizers(ctx, svc, LoadBalancerFinalizer); err != nil {
			return err
		}
	}
	return nil
}

func (r *ServiceReconciler) updateServiceStatus(ctx context.Context, svc *corev1.Service, lbDNS string) error {
	if len(svc.Status.LoadBalancer.Ingress) != 1 || svc.Status.LoadBalancer.Ingress[0].IP != "" || svc.Status.LoadBalancer.Ingress[0].Hostname != lbDNS {
		svc.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{
			{
				Hostname: lbDNS,
			},
		}
		if err := r.k8sClient.Status().Update(ctx, svc); err != nil {
			return errors.Wrapf(err, "failed to update service:%v", svc)
		}
		return r.k8sClient.Status().Update(ctx, svc)
	}
	return nil
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New(controllerName, mgr, controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler:              r,
	})
	if err != nil {
		return err
	}
	return r.setupWatches(mgr, c)
}

func (r *ServiceReconciler) setupWatches(mgr ctrl.Manager, c controller.Controller) error {
	svcEventHandler := eventhandlers.NewEnqueueRequestForServiceEvent(mgr.GetEventRecorderFor(controllerName), r.annotationParser)
	if err := c.Watch(&source.Kind{Type: &corev1.Service{}}, svcEventHandler); err != nil {
		return err
	}
	return nil
}