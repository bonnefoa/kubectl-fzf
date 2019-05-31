package k8sresources

// CtorConfig is the configuration passed to all resource constructors
type CtorConfig struct {
	RoleBlacklist map[string]bool
}
