// Package latent holds the census driver contracts shared with its overlays.
package latent

// DeferredStop selects the lowest-priority refusal after the ordinary walk.
// A non-nil ordinary refusal suppresses the deferred one throughout that walk.
func DeferredStop(ordinary, deferred error) error {
	if ordinary != nil {
		return nil
	}
	return deferred
}
