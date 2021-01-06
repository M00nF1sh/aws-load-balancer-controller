package stack

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	"sigs.k8s.io/aws-load-balancer-controller/pkg/k8s"
	"sigs.k8s.io/aws-load-balancer-controller/test/framework"
	"sigs.k8s.io/aws-load-balancer-controller/test/framework/utils"
	"sync"
)

func NewResourceStackBuilder() *ResourceStackBuilder {
	return &ResourceStackBuilder{}
}

type ResourceStackBuilder struct {
	dps  []*appsv1.Deployment
	svcs []*corev1.Service
	ings []*networking.Ingress
}

func (b *ResourceStackBuilder) AddDeployment(dp *appsv1.Deployment) *ResourceStackBuilder {
	b.dps = append(b.dps, dp)
	return b
}

func (b *ResourceStackBuilder) AddService(svc *corev1.Service) *ResourceStackBuilder {
	b.svcs = append(b.svcs, svc)
	return b
}

func (b *ResourceStackBuilder) AddIngress(ing *networking.Ingress) *ResourceStackBuilder {
	b.ings = append(b.ings, ing)
	return b
}

func (b *ResourceStackBuilder) Build() *ResourceStack {
	return &ResourceStack{
		dps:  b.dps,
		svcs: b.svcs,
		ings: b.ings,
	}
}

// ResourceStack is a stack of Kubernetes resources.
// during Deploy, resources will be created in order by resource type: Deployment -> Service -> Ingress.
// during Cleanup, resources will be destroyed in order by resource type: Ingress -> Service -> Deployment.

var _ TestFixture = &ResourceStack{}

type ResourceStack struct {
	// configurations
	dps  []*appsv1.Deployment
	svcs []*corev1.Service
	ings []*networking.Ingress

	// runtime variables
	createdDPs      []*appsv1.Deployment
	createdDPsMutex sync.Mutex

	createdSVCs      []*corev1.Service
	createdSVCsMutex sync.Mutex

	createdINGs      []*networking.Ingress
	createdINGsMutex sync.Mutex
}

func (s *ResourceStack) SetUp(ctx context.Context, tf *framework.Framework) error {
	if err := s.createDeployments(ctx, tf); err != nil {
		return err
	}
	if err := s.createServices(ctx, tf); err != nil {
		return err
	}
	if err := s.createIngresses(ctx, tf); err != nil {
		return err
	}
	return nil
}

func (s *ResourceStack) TearDown(ctx context.Context, tf *framework.Framework) error {
	var tearDownErrs []error
	if err := s.deleteIngresses(ctx, tf); err != nil {
		tearDownErrs = append(tearDownErrs, err)
	}
	if err := s.deleteServices(ctx, tf); err != nil {
		tearDownErrs = append(tearDownErrs, err)
	}
	if err := s.deleteDeployments(ctx, tf); err != nil {
		tearDownErrs = append(tearDownErrs, err)
	}
	if len(tearDownErrs) != 0 {
		return utils.NewMultiError(tearDownErrs...)
	}
	return nil
}

func (s *ResourceStack) createDeployments(ctx context.Context, tf *framework.Framework) error {
	tf.Logger.Info("create all deployments")
	var createErrs []error
	var createErrsMutex sync.Mutex
	var wg sync.WaitGroup

	for _, dp := range s.dps {
		wg.Add(1)
		go func(dp *appsv1.Deployment) {
			defer wg.Done()
			tf.Logger.Info("creating deployment", "dp", k8s.NamespacedName(dp))
			dp = dp.DeepCopy()
			if err := tf.K8sClient.Create(ctx, dp); err != nil {
				createErrsMutex.Lock()
				createErrs = append(createErrs, err)
				createErrsMutex.Unlock()
				return
			}
			s.createdDPsMutex.Lock()
			s.createdDPs = append(s.createdDPs, dp)
			s.createdDPsMutex.Unlock()
			tf.Logger.Info("created deployment", "dp", k8s.NamespacedName(dp))
		}(dp)
	}

	wg.Wait()
	if len(createErrs) != 0 {
		return utils.NewMultiError(createErrs...)
	}
	return nil
}

func (s *ResourceStack) deleteDeployments(ctx context.Context, tf *framework.Framework) error {
	tf.Logger.Info("delete all deployments")
	var cleanupErrs []error
	var cleanupErrsMutex sync.Mutex
	var wg sync.WaitGroup

	s.createdDPsMutex.Lock()
	defer s.createdDPsMutex.Unlock()
	for _, dp := range s.createdDPs {
		wg.Add(1)
		go func(dp *appsv1.Deployment) {
			defer wg.Done()
			tf.Logger.Info("deleting deployment", "dp", k8s.NamespacedName(dp))
			if err := tf.K8sClient.Delete(ctx, dp); err != nil {
				cleanupErrsMutex.Lock()
				cleanupErrs = append(cleanupErrs, err)
				cleanupErrsMutex.Unlock()
				return
			}
			if err := tf.DPManager.WaitUntilDeploymentDeleted(ctx, dp); err != nil {
				cleanupErrsMutex.Lock()
				cleanupErrs = append(cleanupErrs, err)
				cleanupErrsMutex.Unlock()
				return
			}
			tf.Logger.Info("deleted deployment", "dp", k8s.NamespacedName(dp))
		}(dp)
	}

	wg.Wait()
	if len(cleanupErrs) != 0 {
		return utils.NewMultiError(cleanupErrs...)
	}
	return nil
}

func (s *ResourceStack) createServices(ctx context.Context, f *framework.Framework) error {
	f.Logger.Info("create all services")
	var createErrs []error
	var createErrsMutex sync.Mutex
	var wg sync.WaitGroup

	for _, svc := range s.svcs {
		wg.Add(1)
		go func(svc *corev1.Service) {
			defer wg.Done()
			f.Logger.Info("creating service", "svc", k8s.NamespacedName(svc))
			svc = svc.DeepCopy()
			if err := f.K8sClient.Create(ctx, svc); err != nil {
				createErrsMutex.Lock()
				createErrs = append(createErrs, err)
				createErrsMutex.Unlock()
				return
			}
			s.createdSVCsMutex.Lock()
			s.createdSVCs = append(s.createdSVCs, svc)
			s.createdSVCsMutex.Unlock()
			f.Logger.Info("created service", "svc", k8s.NamespacedName(svc))
		}(svc)
	}

	wg.Wait()
	if len(createErrs) != 0 {
		return utils.NewMultiError(createErrs...)
	}
	return nil
}

func (s *ResourceStack) deleteServices(ctx context.Context, f *framework.Framework) error {
	f.Logger.Info("delete all services")
	var cleanupErrs []error
	var cleanupErrsMutex sync.Mutex
	var wg sync.WaitGroup

	s.createdSVCsMutex.Lock()
	defer s.createdSVCsMutex.Unlock()
	for _, svc := range s.createdSVCs {
		wg.Add(1)
		go func(svc *corev1.Service) {
			defer wg.Done()
			f.Logger.Info("deleting service", "svc", k8s.NamespacedName(svc))
			if err := f.K8sClient.Delete(ctx, svc); err != nil {
				cleanupErrsMutex.Lock()
				cleanupErrs = append(cleanupErrs, err)
				cleanupErrsMutex.Unlock()
				return
			}
			if err := f.SVCManager.WaitUntilServiceDeleted(ctx, svc); err != nil {
				cleanupErrsMutex.Lock()
				cleanupErrs = append(cleanupErrs, err)
				cleanupErrsMutex.Unlock()
				return
			}
			f.Logger.Info("deleted service", "svc", k8s.NamespacedName(svc))
		}(svc)
	}

	wg.Wait()
	if len(cleanupErrs) != 0 {
		return utils.NewMultiError(cleanupErrs...)
	}
	return nil
}

func (s *ResourceStack) createIngresses(ctx context.Context, f *framework.Framework) error {
	f.Logger.Info("create all ingresses")
	var createErrs []error
	var createErrsMutex sync.Mutex
	var wg sync.WaitGroup

	for _, ing := range s.ings {
		wg.Add(1)
		go func(ing *networking.Ingress) {
			defer wg.Done()
			f.Logger.Info("creating ingress", "ing", k8s.NamespacedName(ing))
			ing = ing.DeepCopy()
			if err := f.K8sClient.Create(ctx, ing); err != nil {
				createErrsMutex.Lock()
				createErrs = append(createErrs, err)
				createErrsMutex.Unlock()
				return
			}
			s.createdINGsMutex.Lock()
			s.createdINGs = append(s.createdINGs, ing)
			s.createdINGsMutex.Unlock()
			f.Logger.Info("created ingress", "ing", k8s.NamespacedName(ing))
		}(ing)
	}

	wg.Wait()
	if len(createErrs) != 0 {
		return utils.NewMultiError(createErrs...)
	}
	return nil
}

func (s *ResourceStack) deleteIngresses(ctx context.Context, f *framework.Framework) error {
	f.Logger.Info("delete all ingresses")
	var cleanupErrs []error
	var cleanupErrsMutex sync.Mutex
	var wg sync.WaitGroup

	s.createdINGsMutex.Lock()
	defer s.createdINGsMutex.Unlock()
	for _, ing := range s.createdINGs {
		wg.Add(1)
		go func(ing *networking.Ingress) {
			defer wg.Done()
			f.Logger.Info("deleting ingress", "ing", k8s.NamespacedName(ing))
			if err := f.K8sClient.Delete(ctx, ing); err != nil {
				cleanupErrsMutex.Lock()
				cleanupErrs = append(cleanupErrs, err)
				cleanupErrsMutex.Unlock()
				return
			}
			if err := f.INGManager.WaitUntilIngressDeleted(ctx, ing); err != nil {
				cleanupErrsMutex.Lock()
				cleanupErrs = append(cleanupErrs, err)
				cleanupErrsMutex.Unlock()
				return
			}
			f.Logger.Info("deleted ingress", "ing", k8s.NamespacedName(ing))
		}(ing)
	}

	wg.Wait()
	if len(cleanupErrs) != 0 {
		return utils.NewMultiError(cleanupErrs...)
	}
	return nil
}
