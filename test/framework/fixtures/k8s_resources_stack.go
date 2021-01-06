package fixtures

import (
	"context"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/aws-load-balancer-controller/pkg/k8s"
	"sigs.k8s.io/aws-load-balancer-controller/test/framework"
	"sigs.k8s.io/aws-load-balancer-controller/test/framework/utils"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sync"
)

type K8sResource interface {
	metav1.Object
	runtime.Object
}

var _ framework.TestFixture = &K8sResourcesStack{}

// K8sResourcesStack is an TestFixture that creates specified K8s resources during setup and deletes created K8S resources during cleanUp.
type K8sResourcesStack struct {
	resources        []K8sResource
	createdResources []K8sResource
}

func (s *K8sResourcesStack) SetUp(ctx context.Context, tf *framework.Framework) error {
	for _, res := range s.resources {
		gvk, err := apiutil.GVKForObject(res, tf.K8sScheme)
		if err != nil {
			return err
		}

		tf.Logger.Info("creating resource",
			"kind", gvk.Kind,
			"namespacedName", k8s.NamespacedName(res))
		if err := tf.K8sClient.Create(ctx, res.DeepCopyObject()); err != nil {
			return err
		}
		s.createdResources = append(s.createdResources, res)
		tf.Logger.Info("created resource",
			"kind", gvk.Kind,
			"namespacedName", k8s.NamespacedName(res))
	}

	return nil
}

func (s *K8sResourcesStack) TearDown(ctx context.Context, tf *framework.Framework) error {
	var errs []error
	var errsMutex sync.Mutex
	var wg sync.WaitGroup
	for _, res := range s.createdResources {
		wg.Add(1)
		go func(res K8sResource) {
			defer wg.Done()
			if err := s.deleteOneResource(ctx, tf, res); err != nil {
				errsMutex.Lock()
				errs = append(errs, err)
				errsMutex.Unlock()
			}
		}(res)
	}
	wg.Wait()
	return utilerrors.NewAggregate(errs)
}

func (s *K8sResourcesStack) deleteOneResource(ctx context.Context, tf *framework.Framework, res K8sResource) error {
	gvk, err := apiutil.GVKForObject(res, tf.K8sScheme)
	if err != nil {
		return err
	}

	tf.Logger.Info("deleting resource",
		"kind", gvk.Kind,
		"namespacedName", k8s.NamespacedName(res))
	if err := tf.K8sClient.Delete(ctx, res); err != nil {
		return err
	}
	tf.Logger.Info("waiting resource deletion",
		"kind", gvk.Kind,
		"namespacedName", k8s.NamespacedName(res))
	if err := s.waitUntilResourceDeleted(ctx, tf, res); err != nil {
		return err
	}
	tf.Logger.Info("deleted resource",
		"kind", gvk.Kind,
		"namespacedName", k8s.NamespacedName(res))
	return nil
}

func (s *K8sResourcesStack) waitUntilResourceDeleted(ctx context.Context, tf *framework.Framework, res K8sResource) error {
	objMeta, err := meta.Accessor(res)
	if err != nil {
		return err
	}
	observedRes := res.DeepCopyObject()
	return wait.PollImmediateUntil(utils.PollIntervalShort, func() (bool, error) {
		if err := tf.K8sClient.Get(ctx, k8s.NamespacedName(objMeta), observedRes); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	}, ctx.Done())
}

func (s *K8sResourcesStack) safeGetObjectKind(scheme *runtime.Scheme, res K8sResource) string {
	gvk, err := apiutil.GVKForObject(res, scheme)
	if err != nil {
		return ""
	}
	return gvk.Kind
}
