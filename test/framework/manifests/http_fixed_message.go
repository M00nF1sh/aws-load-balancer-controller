package manifests

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewHTTPFixedMessageBuilder constructs new HTTPFixedMessage.
func NewHTTPFixedMessageBuilder(namespace string, name string) *HTTPFixedMessageBuilder {
	return &HTTPFixedMessageBuilder{
		namespace: namespace,
		name:      name,

		replicas:       1,
		message:        "DUMMY",
		containerPort:  8080,
		svcPort:        80,
		svcAnnotations: nil,
	}
}

// HTTPFixedMessageBuilder is a builder to build the Deployment & Service for a app that returns fixed message
type HTTPFixedMessageBuilder struct {
	namespace string
	name      string

	replicas       int32
	message        string
	containerPort  int32
	svcPort        int32
	svcAnnotations map[string]string
}

func (b *HTTPFixedMessageBuilder) WithMessage(message string) *HTTPFixedMessageBuilder {
	b.message = message
	return b
}

func (b *HTTPFixedMessageBuilder) WithServicePort(port int32) *HTTPFixedMessageBuilder {
	b.svcPort = port
	return b
}

func (b *HTTPFixedMessageBuilder) WithServiceAnnotations(annotations map[string]string) *HTTPFixedMessageBuilder {
	b.svcAnnotations = annotations
	return b
}

func (b *HTTPFixedMessageBuilder) Build() (*appsv1.Deployment, *corev1.Service) {
	dpName := b.deploymentName()
	dpLabels := b.deploymentLabels()
	dp := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: b.namespace,
			Name:      dpName,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: dpLabels,
			},
			Replicas: aws.Int32(b.replicas),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: dpLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "970805265562.dkr.ecr.us-west-2.amazonaws.com/colorteller:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "SERVER_PORT",
									Value: fmt.Sprintf("%d", b.containerPort),
								},
								{
									Name:  "COLOR",
									Value: b.message,
								},
							},
						},
					},
				},
			},
		},
	}

	svcName := b.serviceName()
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   b.namespace,
			Name:        svcName,
			Annotations: b.svcAnnotations,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: dpLabels,
			Ports: []corev1.ServicePort{
				{
					Port:       b.svcPort,
					TargetPort: intstr.FromInt(int(b.containerPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	return dp, svc
}

func (b *HTTPFixedMessageBuilder) deploymentName() string {
	return b.name
}

func (b *HTTPFixedMessageBuilder) deploymentLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name": b.name,
	}
}

func (b *HTTPFixedMessageBuilder) serviceName() string {
	return b.name
}
