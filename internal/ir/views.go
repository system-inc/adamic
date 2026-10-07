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
)

// Contracts describe declared logical types, independently of physical layout ids.
// Readiness remains the shared non-null state and is never implied by a contract.
type ViewContract struct {
	Kind       ViewKind
	Name       string
	Of         Type
	Allowed    []ViewLiteral
	Fields     []ViewFieldContract
	Members    []ViewContractID
	Element    ViewContractID
	Tuple      []ViewContractID
	Functions  []int
	Parameters []ViewContractID
	Result     ViewContractID
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
