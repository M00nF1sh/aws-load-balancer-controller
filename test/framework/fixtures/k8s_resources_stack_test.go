package fixtures

import (
	"context"
	"fmt"
	awssdk "github.com/aws/aws-sdk-go/aws"
	zapraw "go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	elbv2api "sigs.k8s.io/aws-load-balancer-controller/apis/elbv2/v1beta1"
	"sigs.k8s.io/aws-load-balancer-controller/test/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
	"time"
)

func TestK8sResourcesStack_SetUp(t *testing.T) {
	os.Setenv("AWS_PROFILE", "m00nf1sh")
	k8sScheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(k8sScheme)
	elbv2api.AddToScheme(k8sScheme)
	restCFG, _ := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: "/Users/yyyng/.kube/config"}, &clientcmd.ConfigOverrides{}).ClientConfig()
	k8sClient, _ := client.New(restCFG, client.Options{Scheme: k8sScheme})
	tf := &framework.Framework{
		K8sClient: k8sClient,
		K8sScheme: k8sScheme,
		Logger: zap.New(zap.UseDevMode(false),
			zap.Level(zapraw.NewAtomicLevelAt(zapraw.InfoLevel)),
			zap.StacktraceLevel(zapraw.NewAtomicLevelAt(zapraw.FatalLevel))),
	}
	dp := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "my-name",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: awssdk.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "my-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "my-app",
					},
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
									Value: fmt.Sprintf("%d", 8080),
								},
							},
						},
					},
				},
			},
		},
	}
	dp2 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "my-name-2",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: awssdk.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "my-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "my-app",
					},
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
									Value: fmt.Sprintf("%d", 8080),
								},
							},
						},
					},
				},
			},
		},
	}

	rs := K8sResourcesStack{
		resources: []K8sResource{dp, dp2},
	}
	ctx := context.Background()
	err := rs.SetUp(ctx, tf)
	fmt.Println(err)
	time.Sleep(10)
	err = rs.TearDown(ctx, tf)
	fmt.Println(err)
}
