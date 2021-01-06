~/go/bin/ginkgo -v -r test/e2e/ingress/ -- \
    --kubeconfig=/Users/yyyng/.kube/config \
    --cluster-name=m00nf1sh-dev \
    --aws-region=us-west-2 \
    --aws-vpc-id=vpc-0ca80a176a589a451 \
    --controller-image=docker.io/amazon/aws-alb-ingress-controller:v2.1.0