package ir

// ReflectionMember describes a checker member, including literal promises.
type ReflectionMember struct {
	Kind    Type
	Literal bool
	Text    string
	Number  float64
	Boolean bool
}

type ReflectionField struct {
	Name    string
	Index   bool
	Members []ReflectionMember
}

// ObjectReflection is a visible, unconditional check over actual enumerable slots.
// OwnPropertyOrder is required by OrdinaryOwnPropertyKeys, independently of layout.
type ObjectReflection struct {
	Members          []ReflectionMember
	Sources          [][]ReflectionField
	Targets          []ReflectionField
	Message          string
	OwnPropertyOrder bool
}
