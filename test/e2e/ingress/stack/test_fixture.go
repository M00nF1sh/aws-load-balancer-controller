package stack

import (
	"context"
	"sigs.k8s.io/aws-load-balancer-controller/test/framework"
)

// TestFixture represents the preparation needed to perform one or more tests, and any associated cleanup actions
type TestFixture interface {
	SetUp(ctx context.Context, tf *framework.Framework) error
	TearDown(ctx context.Context, tf *framework.Framework) error
}


