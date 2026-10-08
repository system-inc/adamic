package ir

// UsesNullSentinel reports reference representations with distinct null and undefined pointers.
// Other reference kinds retain their existing representation until nullable references generalizes it.
func (t Type) UsesNullSentinel() bool { return t == String }
