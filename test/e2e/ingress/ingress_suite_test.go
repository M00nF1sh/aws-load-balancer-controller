package ingress

import (
	"fmt"
	"sigs.k8s.io/aws-load-balancer-controller/test/framework"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var tf *framework.Framework

func TestIngress(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ingress Suite")
}

var _ = BeforeSuite(func() {
	var err error
	tf, err = framework.InitFramework()
	Expect(err).NotTo(HaveOccurred())

	if tf.Options.ControllerImage != "" {
		By(fmt.Sprintf("ensure cluster installed with controller: %s", tf.Options.ControllerImage), func() {
			err := tf.CTRLInstallationManager.UpgradeController(tf.Options.ControllerImage)
			Expect(err).NotTo(HaveOccurred())
		})
	}
})
