package ingress

import "time"

const (
	// prefix for namespaces used for testing
	namespacePrefix = "aws-lb-e2e"

	// timeout for a created LoadBalancer to propagate DNS records.
	dnsPropagationTimeout = 5 * time.Minute
)
