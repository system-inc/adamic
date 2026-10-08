// Package flattree defines a scout reference format for parser snapshots.
// This is Go tooling, not an Adamic implementation of typed arrays or mapping.
package flattree

import (
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf16"
)

const columns = 12
const headerSize = 40

var magic = [8]byte{'A', 'D', 'F', 'L', 'A', 'T', 0, 1}

// Node is the transport snapshot of every current ParseNode field. Text and Raw
// use the parser's written() escapes so lone UTF-16 surrogates survive transport.
type Node struct {
	Kind                                string
	Pos, End, Flags, LiteralFlags, List int
	Trailing, MultiLine                 bool
	Operator, Text, Raw, Semantic       string
	Children                            []int
}
type Tree struct {
	Nodes []Node
	Root  int
	Roots []int
}

func units(s string, escaped bool) ([]uint16, error) {
	if !escaped {
		return utf16.Encode([]rune(s)), nil
	}
	var out []uint16
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			if s[i] > 126 || s[i] < 32 {
				return nil, errors.New("invalid written text")
			}
			out = append(out, uint16(s[i]))
			continue
		}
		if i+5 >= len(s) || s[i+1] != 'u' {
			return nil, errors.New("invalid written escape")
		}
		var v uint16
		for j := i + 2; j <= i+5; j++ {
			c := s[j]
			var d byte
			switch {
			case c >= '0' && c <= '9':
				d = c - '0'
			case c >= 'a' && c <= 'f':
				d = c - 'a' + 10
			default:
				return nil, errors.New("invalid hex")
			}
			v = v*16 + uint16(d)
		}
		out = append(out, v)
		i += 5
	}
	return out, nil
}
func key(u []uint16) string {
	b := make([]byte, len(u)*2)
	for i, v := range u {
		binary.LittleEndian.PutUint16(b[i*2:], v)
	}
	return string(b)
}

// Encode writes deterministic little-endian structure-of-arrays columns,
// followed by child indices, expression roots, string ranges and UTF-16 units.
// It keeps table order, unreachable speculative nodes and child order intact.
func Encode(t Tree) ([]byte, error) {
	if len(t.Nodes) == 0 || t.Root < 0 || t.Root >= len(t.Nodes) {
		return nil, errors.New("invalid root")
	}
	strings := [][]uint16{nil}
	ids := map[string]uint32{"": 0}
	intern := func(s string, escaped bool) (uint32, error) {
		u, e := units(s, escaped)
		if e != nil {
			return 0, e
		}
		k := key(u)
		if id, ok := ids[k]; ok {
			return id, nil
		}
		id := uint32(len(strings))
		ids[k] = id
		strings = append(strings, u)
		return id, nil
	}
	values := make([][columns]uint32, len(t.Nodes))
	var children []uint32
	for i, n := range t.Nodes {
		if n.Pos < 0 || n.End < n.Pos || n.Flags < 0 || n.LiteralFlags < 0 || n.List < -1 || uint64(n.Pos) > uint64(^uint32(0)) || uint64(n.End) > uint64(^uint32(0)) || uint64(n.LiteralFlags) > uint64(^uint32(0)) || n.List > 2147483647 || n.Flags & ^32 != 0 {
			return nil, fmt.Errorf("invalid node %d", i)
		}
		v := &values[i]
		v[1] = uint32(n.Pos)
		v[2] = uint32(n.End)
		v[3] = uint32(n.Flags)
		if n.Trailing {
			v[3] |= 1 << 30
		}
		if n.MultiLine {
			v[3] |= 1 << 29
		}
		v[4] = uint32(n.LiteralFlags)
		v[5] = uint32(int32(n.List))
		v[6] = uint32(len(children))
		v[7] = uint32(len(n.Children))
		for j, s := range []string{n.Kind, n.Text, n.Raw, n.Operator, n.Semantic} {
			c := []int{0, 8, 9, 10, 11}[j]
			id, e := intern(s, j == 1 || j == 2)
			if e != nil {
				return nil, e
			}
			v[c] = id
		}
		for _, child := range n.Children {
			if child < 0 || child >= len(t.Nodes) {
				return nil, errors.New("invalid child")
			}
			children = append(children, uint32(child))
		}
	}
	unitCount := 0
	for _, s := range strings {
		unitCount += len(s)
	}
	size := headerSize + columns*4*len(t.Nodes) + 4*(len(children)+len(t.Roots)) + 8*len(strings) + 2*unitCount
	if uint64(size) > uint64(^uint32(0)) {
		return nil, errors.New("file exceeds uint32")
	}
	b := make([]byte, size)
	copy(b, magic[:])
	put := func(off int, v uint32) { binary.LittleEndian.PutUint32(b[off:], v) }
	for i, v := range []uint32{uint32(size), uint32(len(t.Nodes)), uint32(len(children)), uint32(len(t.Roots)), uint32(len(strings)), uint32(unitCount), uint32(t.Root), 0} {
		put(8+4*i, v)
	}
	off := headerSize
	for c := 0; c < columns; c++ {
		for _, v := range values {
			put(off, v[c])
			off += 4
		}
	}
	for _, v := range children {
		put(off, v)
		off += 4
	}
	for _, r := range t.Roots {
		if r < 0 || r >= len(t.Nodes) {
			return nil, errors.New("invalid expression root")
		}
		put(off, uint32(r))
		off += 4
	}
	start := 0
	for _, s := range strings {
		put(off, uint32(start))
		put(off+4, uint32(len(s)))
		off += 8
		start += len(s)
	}
	for _, s := range strings {
		for _, u := range s {
			binary.LittleEndian.PutUint16(b[off:], u)
			off += 2
		}
	}
	if _, err := Open(b); err != nil {
		return nil, err
	}
	return b, nil
}

// Reader holds just the input bytes and scalar offsets. Open validates them in
// place without rebuilding nodes, arrays, strings, or a pointer AST. The owner
// must keep the bytes alive and unchanged for the lifetime of Reader.
type Reader struct {
	data                                         []byte
	nodes, children, roots, strings, units, root uint32
	childOff, rootOff, stringOff, unitOff        int
}

func Open(b []byte) (Reader, error) {
	fail := func() (Reader, error) { return Reader{}, errors.New("invalid flat tree") }
	if len(b) < headerSize || string(b[:8]) != string(magic[:]) {
		return fail()
	}
	word := func(off int) uint32 { return binary.LittleEndian.Uint32(b[off:]) }
	r := Reader{data: b, nodes: word(12), children: word(16), roots: word(20), strings: word(24), units: word(28), root: word(32)}
	size := uint64(headerSize) + uint64(r.nodes)*columns*4 + (uint64(r.children)+uint64(r.roots))*4 + uint64(r.strings)*8 + uint64(r.units)*2
	if word(8) != uint32(len(b)) || size != uint64(len(b)) || word(36) != 0 || r.nodes == 0 || r.strings == 0 || r.root >= r.nodes {
		return fail()
	}
	r.childOff = headerSize + int(r.nodes)*columns*4
	r.rootOff = r.childOff + int(r.children)*4
	r.stringOff = r.rootOff + int(r.roots)*4
	r.unitOff = r.stringOff + int(r.strings)*8
	for id := uint32(0); id < r.strings; id++ {
		s, n := r.stringRange(id)
		if uint64(s)+uint64(n) > uint64(r.units) {
			return fail()
		}
	}
	for i := uint32(0); i < r.nodes; i++ {
		if r.Value(1, i) > r.Value(2, i) || int32(r.Value(5, i)) < -1 || r.Value(3, i)&^uint32(32|1<<30|1<<29) != 0 {
			return fail()
		}
		for _, c := range []int{0, 8, 9, 10, 11} {
			if r.Value(c, i) >= r.strings {
				return fail()
			}
		}
		if uint64(r.Value(6, i))+uint64(r.Value(7, i)) > uint64(r.children) {
			return fail()
		}
	}
	for off := r.childOff; off < r.stringOff; off += 4 {
		if word(off) >= r.nodes {
			return fail()
		}
	}
	// Parser tables are DAGs, but speculative entries need not be reachable.
	// Reject cycles using an encoder-independent depth-first proof.
	state := make([]byte, r.nodes)
	var visit func(uint32) bool
	visit = func(i uint32) bool {
		if state[i] == 1 {
			return false
		}
		if state[i] == 2 {
			return true
		}
		state[i] = 1
		for j := uint32(0); j < r.Value(7, i); j++ {
			if !visit(r.Child(i, j)) {
				return false
			}
		}
		state[i] = 2
		return true
	}
	for i := uint32(0); i < r.nodes; i++ {
		if !visit(i) {
			return fail()
		}
	}
	return r, nil
}
func (r Reader) Value(column int, node uint32) uint32 {
	if column < 0 || column >= columns || node >= r.nodes {
		panic("flat column bounds")
	}
	return binary.LittleEndian.Uint32(r.data[headerSize+(column*int(r.nodes)+int(node))*4:])
}
func (r Reader) Child(node, index uint32) uint32 {
	if index >= r.Value(7, node) {
		panic("flat child bounds")
	}
	return binary.LittleEndian.Uint32(r.data[r.childOff+int(r.Value(6, node)+index)*4:])
}
func (r Reader) stringRange(id uint32) (uint32, uint32) {
	off := r.stringOff + int(id)*8
	return binary.LittleEndian.Uint32(r.data[off:]), binary.LittleEndian.Uint32(r.data[off+4:])
}
func (r Reader) StringUnits(id uint32) []byte {
	if id >= r.strings {
		panic("flat string bounds")
	}
	s, n := r.stringRange(id)
	return r.data[r.unitOff+int(s)*2 : r.unitOff+int(s+n)*2]
}
func (r Reader) Root() uint32      { return r.root }
func (r Reader) NodeCount() uint32 { return r.nodes }
func (r Reader) Walk(node uint32) uint64 {
	sum := uint64(r.Value(1, node)) + uint64(r.Value(2, node)) + uint64(r.Value(3, node)) + 1
	for j := uint32(0); j < r.Value(7, node); j++ {
		sum += r.Walk(r.Child(node, j))
	}
	return sum
}
