package registry

import "github.com/Protocol-Lattice/graphql/executor"

// Global executor instance for backward compatibility
var globalExecutor = executor.New()

// ResolverFunc defines the function signature for all resolvers.
// This is re-exported for convenience.
type ResolverFunc = executor.ResolverFunc

// RegisterQueryResolver registers a resolver for a query field in the global executor.
func RegisterQueryResolver(field string, resolver ResolverFunc) {
	globalExecutor.RegisterQueryResolver(field, resolver)
}

// RegisterMutationResolver registers a resolver for a mutation field in the global executor.
func RegisterMutationResolver(field string, resolver ResolverFunc) {
	globalExecutor.RegisterMutationResolver(field, resolver)
}

// RegisterSubscriptionResolver registers a resolver for a subscription field in the global executor.
func RegisterSubscriptionResolver(field string, resolver ResolverFunc) {
	globalExecutor.RegisterSubscriptionResolver(field, resolver)
}

// GetGlobalExecutor returns the global executor instance.
// This allows the handler package to access the registered resolvers.
func GetGlobalExecutor() *executor.Executor {
	return globalExecutor
}
