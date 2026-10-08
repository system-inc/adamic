package ir

// ReflectionMember describes a checker member, including literal promises.
type ReflectionMember struct {
	Kind    Type
	Literal bool
	Text    string
	Number  float64
	Boolean bool
}

// ObjectReflection preserves actual own-key order. Entries carries a visible value check;
// storage-proven assign does not read values as members of the apparent type.
// OwnPropertyOrder is required by OrdinaryOwnPropertyKeys, independently of layout.
type ObjectReflection struct {
	Members          []ReflectionMember
	Message          string
	OwnPropertyOrder bool
}
