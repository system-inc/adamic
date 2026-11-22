package ir

import "strconv"

// CheckedFieldKey is where CheckedFields records a checked view member: under its static receiver's checker type, so an
// unrelated receiver with the same member name is not checked. Gob-encoded programs and the products that hash them
// keep this spelling.
func CheckedFieldKey(receiverTypeID int, name string) string {
	return strconv.Itoa(receiverTypeID) + ":" + name
}

// CheckedField says whether a read or write of name through a receiver of type receiverTypeID is checked: a view member
// under its receiver's key, or a record's or merged class's field under its bare name, through any receiver lowering
// hasn't proven plain (PlainReceivers). An unproven receiver stays checked.
func (p *Program) CheckedField(receiverTypeID int, name string) bool {
	return p.CheckedFields[CheckedFieldKey(receiverTypeID, name)] || p.CheckedFields[name] && !p.PlainReceivers[receiverTypeID]
}

// WidenedReadKey is where WidenedReads records a widened read: its CheckedFieldKey and its place in the source. Two
// reads that share a key, two lowerings of one source read or two reads with no place, are both checked, which only adds
// checks.
func WidenedReadKey(read Property) string {
	return CheckedFieldKey(read.ViewReceiverTypeID, read.Name) + "@" + read.ViewWhere
}

// CheckedRead is CheckedField for a field read, which can prove its receiver plain by its value as well
// as its type. A read made where a storage test found neither a record nor a class instance
// (Property.NotRecord, a weak read's fixed branch) is checked only as a view member under its receiver's
// key, as a read through a PlainReceivers type is. A widened read (WidenedReads) is checked wherever it's made.
func (p *Program) CheckedRead(read Property) bool {
	if p.WidenedReads[WidenedReadKey(read)] {
		return true
	}
	if read.NotRecord {
		return p.CheckedFields[CheckedFieldKey(read.ViewReceiverTypeID, read.Name)]
	}
	return p.CheckedField(read.ViewReceiverTypeID, read.Name)
}

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

	// NominalClass is an existing erased class identity, used by Map stores.
	NominalClass  int
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
	TupleUnion    bool // Disjoint tuple arity alternatives, carried only by tuple plans.
	FixedTuple    bool
	TupleVariable bool // Only tuple plans may carry an arity interval.
	TupleMinimum  int
	TupleRest     ViewContractID
	Tuple         []ViewContractID
	Functions     []int
	// ProducerCertified requires immutable code identity in Functions, even when empty.
	ProducerCertified bool
	Parameters        []ViewContractID
	Result            ViewContractID
	// DiscardResult certifies an erased zero-argument marker; it cannot supply a valued result.
	DiscardResult bool
	// StaticsAdmitted lets a class constructor stand for this object (#s03ndgw): no field it names is
	// a property of Object's, and no optional one is a property of Function's, which a class that
	// doesn't declare it would read on Node. Lowering decides it from the checker; the runtime still
	// asks that every required field be a data static a class on the constructor's chain wrote (name,
	// length and prototype only the class's own), that no field be an accessor, and refuses it in a
	// union a callable could also stand in.
	StaticsAdmitted bool
}

// StaticsAdmittedReceiver is whether a write's static receiver, receiverTypeID, is an object a class
// constructor may stand for (ViewContract.StaticsAdmitted). Such a view reads a class's statics and
// never writes them, so each backend stops the write when the receiver is a class.
func (p *Program) StaticsAdmittedReceiver(receiverTypeID int) bool {
	id := p.ViewContractTypes[receiverTypeID]
	return receiverTypeID != 0 && id > 0 && int(id) <= len(p.ViewContracts) && p.ViewContracts[id-1].StaticsAdmitted
}

// ObjectReadContract is the object contract a checked read expects outside a union, an interface's field
// read straight from it, which no member selection reads: each backend stops a class constructor read
// there unless the contract admits one (StaticsAdmitted) and the class holds every required field as a data
// static, its own or a base's, with no field the run time refuses (#s03ndgw), as a union's selection does.
// A read of another kind, or through a union, a nominal class's or an unsupported contract, is false.
func (p *Program) ObjectReadContract(read Property) (ViewContract, bool) {
	id := read.ViewContract
	if read.Of != Object || id <= 0 || int(id) > len(p.ViewContracts) {
		return ViewContract{}, false
	}
	contract := p.ViewContracts[id-1]
	return contract, contract.Kind == ViewObject && contract.Nominal == "" && contract.Unsupported == ""
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
