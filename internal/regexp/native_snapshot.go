package regexp

import (
	"encoding/binary"
	"fmt"
	"sort"
)

// NativeBytecodeSnapshot exposes the VM's pointer-free contents for C compiler
// identity tests. Every numeric instruction field uses little-endian uint64;
// arrays and strings are length-prefixed. It excludes optional search/compiled
// fast paths, which are derived optimizations rather than bytecode. This is a
// test seam, not a persistent bytecode file format.
func (p *Program) NativeBytecodeSnapshot() ([]byte, error) {
	var out []byte
	number := func(value uint64) { out = binary.LittleEndian.AppendUint64(out, value) }
	boolean := func(value bool) {
		if value {
			number(1)
		} else {
			number(0)
		}
	}
	text := func(value string) { number(uint64(len(value))); out = append(out, value...) }
	var emit func(*Program) error
	emit = func(program *Program) error {
		number(uint64(nativeFlags(program.flags)))
		number(uint64(program.captures))
		number(uint64(program.repeats))
		number(uint64(len(program.code)))
		names := make([]string, 0, len(program.names))
		for name := range program.names {
			names = append(names, name)
		}
		sort.Strings(names)
		number(uint64(len(names)))
		for _, name := range names {
			text(name)
			ids := program.names[name]
			number(uint64(len(ids)))
			for _, id := range ids {
				number(uint64(id))
			}
		}
		for _, i := range program.code {
			if i.min != nil && !i.min.IsUint64() || i.max != nil && !i.max.IsUint64() {
				return fmt.Errorf("native regexp quantifier bounds above uint64 are not yet supported")
			}
			number(uint64(i.op))
			number(uint64(i.x))
			number(uint64(i.y))
			number(uint64(i.direction))
			number(uint64(nativeFlags(i.flags)))
			number(uint64(i.assertion))
			boolean(i.negative)
			boolean(i.greedy)
			boolean(i.max == nil)
			var lo, hi uint64
			if i.min != nil {
				lo = i.min.Uint64()
			}
			if i.max != nil {
				hi = i.max.Uint64()
			}
			number(lo)
			number(hi)
			number(uint64(len(i.set.ranges)))
			for _, interval := range i.set.ranges {
				number(uint64(interval.From))
				number(uint64(interval.To))
			}
			number(uint64(len(i.set.strings)))
			for _, sequence := range i.set.strings {
				number(uint64(len(sequence)))
				for _, point := range sequence {
					number(uint64(point))
				}
			}
			ids := i.references
			if i.op == opRepeat {
				ids = i.clear
			}
			number(uint64(len(ids)))
			for _, id := range ids {
				number(uint64(id))
			}
			boolean(i.look != nil)
			if i.look != nil {
				if err := emit(i.look); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := emit(p); err != nil {
		return nil, err
	}
	return out, nil
}
