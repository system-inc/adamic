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
func (program *Program) NativeBytecodeSnapshot() ([]byte, error) {
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
		for _, instruction := range program.code {
			if instruction.min != nil && !instruction.min.IsUint64() ||
				instruction.max != nil && !instruction.max.IsUint64() {
				return fmt.Errorf("native regexp quantifier bounds above uint64 are not yet supported")
			}
			number(uint64(instruction.op))
			number(uint64(instruction.x))
			number(uint64(instruction.y))
			number(uint64(instruction.direction))
			number(uint64(nativeFlags(instruction.flags)))
			number(uint64(instruction.assertion))
			boolean(instruction.negative)
			boolean(instruction.greedy)
			boolean(instruction.max == nil)
			var lo, hi uint64
			if instruction.min != nil {
				lo = instruction.min.Uint64()
			}
			if instruction.max != nil {
				hi = instruction.max.Uint64()
			}
			number(lo)
			number(hi)
			number(uint64(len(instruction.set.ranges)))
			for _, interval := range instruction.set.ranges {
				number(uint64(interval.From))
				number(uint64(interval.To))
			}
			number(uint64(len(instruction.set.strings)))
			for _, sequence := range instruction.set.strings {
				number(uint64(len(sequence)))
				for _, point := range sequence {
					number(uint64(point))
				}
			}
			ids := instruction.references
			if instruction.op == opRepeat {
				ids = instruction.clear
			}
			number(uint64(len(ids)))
			for _, id := range ids {
				number(uint64(id))
			}
			boolean(instruction.look != nil)
			if instruction.look != nil {
				if emitError := emit(instruction.look); emitError != nil {
					return emitError
				}
			}
		}
		return nil
	}
	if emitError := emit(program); emitError != nil {
		return nil, emitError
	}
	return out, nil
}
