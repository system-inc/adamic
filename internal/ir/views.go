package ir

// ViewContractID is one-based; zero means no available semantic certificate.
type ViewContractID int

type ViewKind uint8

const (
	ViewUnknown ViewKind = iota
	ViewScalar
	ViewObject
	ViewArray
	ViewUnion
	ViewCallable
	ViewNull
	ViewUndefined
	ViewNullable
	ViewMap
	ViewDictionary
)

// Contracts describe declared logical types, independently of physical layout ids.
// Readiness remains the shared non-null state and is never implied by a contract.
type ViewContract struct {
	// Intersection marks conjunctive object members, distinct from union selection.
	Intersection bool
	// IntersectionTag selects disjoint object arms before validating their fields.
	IntersectionTag       string
	IntersectionRecursive bool
	// IntersectionBounded validates a recursive or compound read to contract-level
	// recursion; deeper and deferred fields keep their own checked reads.
	IntersectionBounded bool
	// ObjectPresent retains the canonical descriptor behind optional copies.
	ObjectPresent ViewContractID
	// RepresentationMask preserves members of a boxed callable signature union.
	RepresentationMask uint16
	Payload            ViewContractID
	PayloadTypeID      int

	// Unsupported records the member family that must fail at a demanded read.
	Unsupported string

	NominalBases  []string
	Nominal       string
	Undefined     bool
	Null          bool
	Kind          ViewKind
	Name          string
	Of            Type
	Allowed       []ViewLiteral
	Fields        []ViewFieldContract
	Members       []ViewContractID
	Key           ViewContractID
	MapReadonly   bool
	ArrayReadonly bool
	Element       ViewContractID
	Tuple         []ViewContractID
	Functions     []int
	// ProducerCertified requires immutable code identity in Functions, even when empty.
	ProducerCertified bool
	Parameters        []ViewContractID
	Result            ViewContractID
	// DiscardResult certifies an erased zero-argument marker; it cannot supply a valued result.
	DiscardResult bool
}

type ViewFieldContract struct {
	Name     string
	Contract ViewContractID
	Optional bool
	Readonly bool
}

// Literal constraints are metadata, not emitted constant-pool expressions. An
// unused field contract must not introduce an unused C string declaration.
type ViewLiteral struct {
	Of      Type
	String  string
	Number  float64
	Boolean bool
}
