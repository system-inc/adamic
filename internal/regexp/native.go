package regexp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// NativeDeclarations serializes this already compiled program for the C VM.
// Bounds beyond the C counter width are refused before emitting any C.
func (program *Program) NativeDeclarations(name string) (string, error) {
	if nativeCompatibilityError := program.NativeCompatibility(); nativeCompatibilityError != nil {
		return "", nativeCompatibilityError
	}
	var out strings.Builder
	var emit func(*Program, string) error
	emit = func(program *Program, name string) error {
		for index, instruction := range program.code {
			if instruction.min != nil && !instruction.min.IsUint64() ||
				instruction.max != nil && !instruction.max.IsUint64() {
				return fmt.Errorf("native regexp quantifier bounds above uint64 are not yet supported")
			}
			if instruction.look != nil {
				if emitError := emit(instruction.look, fmt.Sprintf("%s_look_%d", name, index)); emitError != nil {
					return emitError
				}
			}
			if len(instruction.set.ranges) > 0 {
				fmt.Fprintf(&out, "static const adamic_regex_range %s_ranges_%d[] = {", name, index)
				for _, runeRange := range instruction.set.ranges {
					fmt.Fprintf(&out, "{%d,%d},", runeRange.From, runeRange.To)
				}
				out.WriteString("};\n")
			}
			for k, s := range instruction.set.strings {
				fmt.Fprintf(&out, "static const uint32_t %s_text_%d_%d[] = {", name, index, k)
				if len(s) == 0 {
					out.WriteString("0")
				}
				for _, character := range s {
					fmt.Fprintf(&out, "%d,", character)
				}
				out.WriteString("};\n")
			}
			if len(instruction.set.strings) > 0 {
				fmt.Fprintf(&out, "static const adamic_regex_text %s_strings_%d[] = {", name, index)
				for k, s := range instruction.set.strings {
					fmt.Fprintf(&out, "{%s_text_%d_%d,%d},", name, index, k, len(s))
				}
				out.WriteString("};\n")
			}
			ids := instruction.references
			if instruction.op == opRepeat {
				ids = instruction.clear
			}
			if len(ids) > 0 {
				fmt.Fprintf(&out, "static const size_t %s_ids_%d[] = {", name, index)
				for _, id := range ids {
					fmt.Fprintf(&out, "%d,", id)
				}
				out.WriteString("};\n")
			}
		}
		fmt.Fprintf(&out, "static const adamic_regex_instruction %s_code[] = {\n", name)
		for index, instruction := range program.code {
			rangeName, stringName, idName, lookName := "NULL", "NULL", "NULL", "NULL"
			if len(instruction.set.ranges) > 0 {
				rangeName = fmt.Sprintf("%s_ranges_%d", name, index)
			}
			if len(instruction.set.strings) > 0 {
				stringName = fmt.Sprintf("%s_strings_%d", name, index)
			}
			ids := instruction.references
			if instruction.op == opRepeat {
				ids = instruction.clear
			}
			if len(ids) > 0 {
				idName = fmt.Sprintf("%s_ids_%d", name, index)
			}
			if instruction.look != nil {
				lookName = fmt.Sprintf("&%s_look_%d", name, index)
			}
			var lo, hi uint64
			if instruction.min != nil {
				lo = instruction.min.Uint64()
			}
			if instruction.max != nil {
				hi = instruction.max.Uint64()
			}
			fmt.Fprintf(
				&out,
				"{%d,%d,%d,%d,%d,%d,%t,%t,%t,UINT64_C(%d),UINT64_C(%d),%s,%d,%s,%d,%s,%d,%s},\n",
				instruction.op,
				instruction.x,
				instruction.y,
				instruction.direction,
				nativeFlags(instruction.flags),
				instruction.assertion,
				instruction.negative,
				instruction.greedy,
				instruction.max == nil,
				lo,
				hi,
				rangeName,
				len(instruction.set.ranges),
				stringName,
				len(instruction.set.strings),
				idName,
				len(ids),
				lookName,
			)
		}
		out.WriteString("};\n")
		names := make([]string, 0, len(program.names))
		for key := range program.names {
			names = append(names, key)
		}
		sort.Strings(names)
		groupNames := "NULL"
		if len(names) > 0 {
			groupNames = name + "_groups"
			for index, n := range names {
				fmt.Fprintf(&out, "static const size_t %s_group_%d[] = {", name, index)
				for _, id := range program.names[n] {
					fmt.Fprintf(&out, "%d,", id)
				}
				out.WriteString("};\n")
			}
			fmt.Fprintf(&out, "static const adamic_regex_group %s[] = {", groupNames)
			for index, n := range names {
				fmt.Fprintf(&out, "{%s,%s_group_%d,%d},", strconv.QuoteToASCII(n), name, index, len(program.names[n]))
			}
			out.WriteString("};\n")
		}
		shape := "NULL"
		if len(names) > 0 {
			shape = "&" + name + "_group_shape"
			fmt.Fprintf(&out, "static const char *const %s_group_names[] = {", name)
			for _, name := range names {
				fmt.Fprintf(&out, "%s,", strconv.QuoteToASCII(name))
			}
			out.WriteString("};\n")
			fmt.Fprintf(&out, "static const bool %s_group_references[] = {", name)
			for range names {
				out.WriteString("true,")
			}
			out.WriteString("};\n")
			fmt.Fprintf(&out, "static const adamic_field_kind %s_group_kinds[] = {", name)
			for range names {
				out.WriteString("adamic_field_reference,")
			}
			out.WriteString("};\n")
			fmt.Fprintf(
				&out,
				"static const adamic_shape %s_group_shape = {%d,%s_group_names,%s_group_references,&adamic_host_methods,%s_group_kinds};\n",
				name,
				len(names),
				name,
				name,
				name,
			)
		}
		bits, filter := program.nativeFirstAscii()
		prefix := program.nativePrefix()
		prefixName := "NULL"
		if len(prefix) > 0 {
			prefixName = name + "_prefix"
			fmt.Fprintf(&out, "static const uint16_t %s[] = {", prefixName)
			for _, character := range prefix {
				fmt.Fprintf(&out, "%d,", character)
			}
			out.WriteString("};\n")
		}
		fast := "NULL"
		ascii := "NULL"
		if body := program.nativeStraightLine(name); body != "" {
			out.WriteString(body)
			fast = name + "_fast"
		}
		if body := program.nativeStraightLineAscii(name); body != "" {
			out.WriteString(body)
			ascii = name + "_ascii"
		}
		behind := program.nativeLookbehindPrefix()
		behindName := "NULL"
		if len(behind) > 0 {
			behindName = name + "_before"
			fmt.Fprintf(&out, "static const uint16_t %s[] = {", behindName)
			for _, character := range behind {
				fmt.Fprintf(&out, "%d,", character)
			}
			out.WriteString("};\n")
		}
		firstBody, firstName := program.nativeSparseFirst(name)
		out.WriteString(firstBody)
		testName := "NULL"
		if body := program.nativeBooleanAscii(name, firstName); body != "" {
			out.WriteString(body)
			testName = name + "_test"
		}
		regularBody, regularName := program.nativeRegular(name)
		out.WriteString(regularBody)
		_, endAnchored := program.nativeBooleanWidth()
		testPrefix := testName != "NULL" &&
			!program.nativeAnchored() &&
			!program.flags.Sticky &&
			len(prefix) != 0 &&
			!endAnchored
		fmt.Fprintf(
			&out,
			"static const adamic_regex_program %s __attribute__((aligned(64))) = {%s_code,%s,%t,%d,%d,%d,%s,%d,%s,{UINT64_C(%d),UINT64_C(%d)},%t,%s,%d,%s,%t,%s,%s,%s,%d,%s};\n",
			name,
			name,
			testName,
			testPrefix,
			program.captures,
			program.repeats,
			nativeFlags(program.flags),
			groupNames,
			len(names),
			shape,
			bits[0],
			bits[1],
			filter,
			prefixName,
			len(prefix),
			fast,
			program.nativeAnchored(),
			ascii,
			regularName,
			behindName,
			len(behind),
			firstName,
		)
		return nil
	}
	if emitError := emit(program, name); emitError != nil {
		return "", emitError
	}
	return out.String(), nil
}
func nativeFlags(flags Flags) int {
	result := 0
	for bit, on := range []bool{
		flags.IgnoreCase,
		flags.Multiline,
		unicodeMode(flags),
		flags.Global,
		flags.Sticky,
		flags.HasIndices,
		flags.UnicodeSets,
		flags.DotAll,
	} {
		if on {
			result |= 1 << bit
		}
	}
	return result
}
