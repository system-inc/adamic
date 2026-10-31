package ir

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math"
	"sync"
)

// A program gob writes reads back as the program it was, so the C emitted from a fresh decode is the C emitted from
// the lowering. Three things gob does on its own would break that, and this file keeps each from reaching the IR:
//
// gob knows a value behind an Expression or a Statement only by a registered name, which RegisterGob gives every
// concrete type the package declares (gob_types_generated.go, kept current from the source by
// TestGobRegistersEveryInterfaceType).
//
// gob omits a struct's zero fields, and -0 == 0, so a -0 it reads back as +0, which the native backend prints apart
// (emit_values.go). Every type holding a float writes the float's IEEE bits instead, so -0 and every NaN's payload come
// back exactly.
//
// gob skips unexported fields. A type whose unexported field matters after a decode writes it itself (GeneratedType's
// identity); the rest are caches their readers rebuild when nil. TestEveryTypeAProgramHoldsGobsExactly holds both.

var registerGob sync.Once

// RegisterGob registers every type that can sit behind an Expression or a Statement, under the name gob.Register gives
// it, so any process that writes or reads a program can do either. It is safe to call any number of times.
func RegisterGob() {
	registerGob.Do(func() {
		for _, value := range gobTypes {
			gob.Register(value)
		}
	})
}

// GobEncode writes a constant as its IEEE bits, eight bytes, big-endian.
func (constant NumberConstant) GobEncode() ([]byte, error) {
	return binary.BigEndian.AppendUint64(nil, math.Float64bits(constant.Value)), nil
}

// GobDecode reads what GobEncode wrote.
func (constant *NumberConstant) GobDecode(data []byte) error {
	if len(data) != 8 {
		return fmt.Errorf("ir: a NumberConstant is 8 bytes, read %d", len(data))
	}
	constant.Value = math.Float64frombits(binary.BigEndian.Uint64(data))
	return nil
}

// GobEncode writes the guard's fields as gob would and MaxStep's bits beside them.
func (guard CounterGuard) GobEncode() ([]byte, error) {
	return encodeFloatBits(counterGuardFields(guard), guard.MaxStep)
}

// GobDecode reads what GobEncode wrote.
func (guard *CounterGuard) GobDecode(data []byte) error {
	return decodeFloatBits(data, (*counterGuardFields)(guard), &guard.MaxStep)
}

// GobEncode writes the literal's fields as gob would and Number's bits beside them.
func (literal ViewLiteral) GobEncode() ([]byte, error) {
	return encodeFloatBits(viewLiteralFields(literal), literal.Number)
}

// GobDecode reads what GobEncode wrote.
func (literal *ViewLiteral) GobDecode(data []byte) error {
	return decodeFloatBits(data, (*viewLiteralFields)(literal), &literal.Number)
}

// GobEncode writes the node's fields as gob would and Number's bits beside them.
func (node JSONDecodeNode) GobEncode() ([]byte, error) {
	return encodeFloatBits(jsonDecodeNodeFields(node), node.Number)
}

// GobDecode reads what GobEncode wrote.
func (node *JSONDecodeNode) GobDecode(data []byte) error {
	return decodeFloatBits(data, (*jsonDecodeNodeFields)(node), &node.Number)
}

// Each float-holding type's fields, as a conversion that leaves its methods behind, so its GobEncode doesn't call
// itself.
type (
	counterGuardFields   CounterGuard
	viewLiteralFields    ViewLiteral
	jsonDecodeNodeFields JSONDecodeNode
)

// floatBits is a struct's fields, as gob writes them, with its floats' IEEE bits beside them. gob writes every element
// of a slice, zero or not, so Bits keeps a -0 the Fields copy drops.
type floatBits[Fields any] struct {
	Fields Fields
	Bits   []uint64
}

func encodeFloatBits[Fields any](fields Fields, floats ...float64) ([]byte, error) {
	bits := make([]uint64, len(floats))
	for index, float := range floats {
		bits[index] = math.Float64bits(float)
	}
	var encoded bytes.Buffer
	err := gob.NewEncoder(&encoded).Encode(floatBits[Fields]{fields, bits})
	return encoded.Bytes(), err
}

// decodeFloatBits reads the fields, then sets each float from its bits, in the order encodeFloatBits was given them.
func decodeFloatBits[Fields any](data []byte, fields *Fields, floats ...*float64) error {
	var decoded floatBits[Fields]
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&decoded); err != nil {
		return err
	}
	if len(decoded.Bits) != len(floats) {
		return fmt.Errorf("ir: %T holds %d floats, read %d", decoded.Fields, len(floats), len(decoded.Bits))
	}
	*fields = decoded.Fields
	for index, float := range floats {
		*float = math.Float64frombits(decoded.Bits[index])
	}
	return nil
}

// generatedTypeWire is a GeneratedType with its identity as the identity's kind, zero for none.
type generatedTypeWire struct {
	Name     string
	Identity int
}

// GobEncode writes the type's name and which canonical identity it has, which the cycle finder reads
// (RuntimeBreaksCycles) and gob would otherwise drop.
func (generated GeneratedType) GobEncode() ([]byte, error) {
	wire := generatedTypeWire{Name: generated.Name}
	if generated.identity != nil {
		wire.Identity = generated.identity.kind
	}
	var encoded bytes.Buffer
	err := gob.NewEncoder(&encoded).Encode(wire)
	return encoded.Bytes(), err
}

// GobDecode reads what GobEncode wrote, back to the one canonical identity of its kind.
func (generated *GeneratedType) GobDecode(data []byte) error {
	var wire generatedTypeWire
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&wire); err != nil {
		return err
	}
	generated.Name, generated.identity = wire.Name, nil
	if wire.Identity == 0 {
		return nil
	}
	for _, identity := range []*runtimeIdentity{asyncFrameIdentity, asyncPromiseIdentity, asyncReactionIdentity} {
		if identity.kind == wire.Identity {
			generated.identity = identity
			return nil
		}
	}
	return fmt.Errorf("ir: GeneratedType %s has no identity of kind %d", wire.Name, wire.Identity)
}
