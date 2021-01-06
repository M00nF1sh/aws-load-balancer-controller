package framework

import "context"

// TestFixture represents the preparation needed to perform one or more tests, and any associated cleanup actions
type TestFixture interface {
	SetUp(ctx context.Context, tf *Framework) error
	TearDown(ctx context.Context, tf *Framework) error
}
