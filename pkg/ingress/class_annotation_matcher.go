package ingress

import (
	"context"
	networking "k8s.io/api/networking/v1beta1"
)

// ClassAnnotationMatcher tests whether the kubernetes.io/ingress.class annotation on Ingresses matches the IngressClass of this controller.
type ClassAnnotationMatcher interface {
	Matches(ctx context.Context, ing *networking.Ingress) bool
}

var _ ClassAnnotationMatcher = &defaultClassAnnotationMatcher{}

// default implementation for ClassAnnotationMatcher, which supports users to provide a single custom IngressClass.
type defaultClassAnnotationMatcher struct {
	ingressClass string
}

func (m *defaultClassAnnotationMatcher) Matches(ctx context.Context, ing *networking.Ingress) bool {
	if m.ingressClass == "" && ingClassAnnotation == ingressClassALB {
		return true
	}
	return ingClassAnnotation == m.ingressClass
}
