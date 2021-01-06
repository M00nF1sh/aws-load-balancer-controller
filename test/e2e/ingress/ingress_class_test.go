package ingress

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("test IngressClass", func() {
	Context("with kubernetes.io/ingress.class annotation", func() {
		//var (
		//	ctx        context.Context
		//	createdNS  *corev1.Namespace
		//	createdDP  *appsv1.Deployment
		//	createdSVC *corev1.Service
		//	createdING *networking.Ingress
		//)
		//
		//BeforeEach(func() {
		//	ctx = context.Background()
		//
		//	tf.Logger.Info("allocating namespace", "basename", namespacePrefix)
		//	ns, err := tf.NSManager.AllocateNamespace(ctx, namespacePrefix)
		//	Expect(err).NotTo(HaveOccurred())
		//	tf.Logger.Info("allocated namespace", "name", ns.Name)
		//	createdNS = ns
		//
		//	dp, svc := manifests.NewHTTPFixedMessageBuilder(ns.Name, "test-app").
		//		WithMessage("HelloWorld").
		//		Build()
		//	ing := &networking.Ingress{
		//		ObjectMeta: metav1.ObjectMeta{
		//			Namespace: ns.Name,
		//			Name:      "test-app",
		//		},
		//		Spec: networking.IngressSpec{
		//			Backend: &networking.IngressBackend{
		//				ServiceName: svc.Name,
		//				ServicePort: intstr.FromInt(80),
		//			},
		//		},
		//	}
		//
		//})
		//
		//AfterEach(func() {
		//	var cleanupErrs []error
		//	if createdNS != nil {
		//		tf.Logger.Info("deleting namespace", "name", createdNS.Name)
		//		if err := tf.K8sClient.Delete(ctx, createdNS); err != nil {
		//			cleanupErrs = append(cleanupErrs, err)
		//		}
		//	}
		//})

		It("should provision ALB when annotation exists and equal to alb", func() {

		})

		It("should not provision ALB when annotation exists but not equal to alb", func() {

		})

		It("should not provision ALB when annotation not exists", func() {

		})
	})

	Context("with spec.ingressClassName", func() {
		It("should provision ALB when ingressClass exists with ingress.k8s.aws/alb controller", func() {

		})

		It("should not provision ALB when ingressClass exists but without to ingress.k8s.aws/alb controller", func() {

		})

		It("should not provision ALB when ingressClass not exists", func() {

		})
	})
})
