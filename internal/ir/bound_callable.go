package ir

// BoundCallable records actual immutable captures of a generated bound function.
// This reflection metadata supplies no producer signature.
type BoundCallable struct {
	Source    int
	Arguments []int
}

func (p *Program) BoundCaptureSlot(function Function, local int) int {
	for slot, captured := range function.Environment {
		if captured == local {
			return slot
		}
	}
	panic("bound capture missing from environment")
}
