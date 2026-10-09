package regexp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// NativeDeclarations serializes this already compiled program for the C VM.
// Bounds beyond the C counter width are refused before emitting any C.
func (p *Program) NativeDeclarations(name string) (string, error) {
	if err := p.NativeCompatibility(); err != nil {
		return "", err
	}
	var out strings.Builder
	var emit func(*Program, string) error
	emit = func(p *Program, name string) error {
		for index, i := range p.code {
			if i.min != nil && !i.min.IsUint64() || i.max != nil && !i.max.IsUint64() {
				return fmt.Errorf("native regexp quantifier bounds above uint64 are not yet supported")
			}
			if i.look != nil {
				if err := emit(i.look, fmt.Sprintf("%s_look_%d", name, index)); err != nil {
					return err
				}
			}
			if len(i.set.ranges) > 0 {
				fmt.Fprintf(&out, "static const adamic_regex_range %s_ranges_%d[] = {", name, index)
				for _, r := range i.set.ranges {
					fmt.Fprintf(&out, "{%d,%d},", r.From, r.To)
				}
				out.WriteString("};\n")
			}
			for k, s := range i.set.strings {
				fmt.Fprintf(&out, "static const uint32_t %s_text_%d_%d[] = {", name, index, k)
				if len(s) == 0 {
					out.WriteString("0")
				}
				for _, c := range s {
					fmt.Fprintf(&out, "%d,", c)
				}
				out.WriteString("};\n")
			}
			if len(i.set.strings) > 0 {
				fmt.Fprintf(&out, "static const adamic_regex_text %s_strings_%d[] = {", name, index)
				for k, s := range i.set.strings {
					fmt.Fprintf(&out, "{%s_text_%d_%d,%d},", name, index, k, len(s))
				}
				out.WriteString("};\n")
			}
			ids := i.references
			if i.op == opRepeat {
				ids = i.clear
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
		for index, i := range p.code {
			rangeName, stringName, idName, lookName := "NULL", "NULL", "NULL", "NULL"
			if len(i.set.ranges) > 0 {
				rangeName = fmt.Sprintf("%s_ranges_%d", name, index)
			}
			if len(i.set.strings) > 0 {
				stringName = fmt.Sprintf("%s_strings_%d", name, index)
			}
			ids := i.references
			if i.op == opRepeat {
				ids = i.clear
			}
			if len(ids) > 0 {
				idName = fmt.Sprintf("%s_ids_%d", name, index)
			}
			if i.look != nil {
				lookName = fmt.Sprintf("&%s_look_%d", name, index)
			}
			var lo, hi uint64
			if i.min != nil {
				lo = i.min.Uint64()
			}
			if i.max != nil {
				hi = i.max.Uint64()
			}
			fmt.Fprintf(&out, "{%d,%d,%d,%d,%d,%d,%t,%t,%t,UINT64_C(%d),UINT64_C(%d),%s,%d,%s,%d,%s,%d,%s},\n", i.op, i.x, i.y, i.direction, nativeFlags(i.flags), i.assertion, i.negative, i.greedy, i.max == nil, lo, hi, rangeName, len(i.set.ranges), stringName, len(i.set.strings), idName, len(ids), lookName)
		}
		out.WriteString("};\n")
		names := make([]string, 0, len(p.names))
		for n := range p.names {
			names = append(names, n)
		}
		sort.Strings(names)
		groupNames := "NULL"
		if len(names) > 0 {
			groupNames = name + "_groups"
			for k, n := range names {
				fmt.Fprintf(&out, "static const size_t %s_group_%d[] = {", name, k)
				for _, id := range p.names[n] {
					fmt.Fprintf(&out, "%d,", id)
				}
				out.WriteString("};\n")
			}
			fmt.Fprintf(&out, "static const adamic_regex_group %s[] = {", groupNames)
			for k, n := range names {
				fmt.Fprintf(&out, "{%s,%s_group_%d,%d},", strconv.QuoteToASCII(n), name, k, len(p.names[n]))
			}
			out.WriteString("};\n")
		}
		shape := "NULL"
		if len(names) > 0 {
			shape = "&" + name + "_group_shape"
			fmt.Fprintf(&out, "static const char *const %s_group_names[] = {", name)
			for _, n := range names {
				fmt.Fprintf(&out, "%s,", strconv.QuoteToASCII(n))
			}
			out.WriteString("};\n")
			fmt.Fprintf(&out, "static const bool %s_group_references[] = {", name)
			for range names {
				out.WriteString("true,")
			}
			out.WriteString("};\n")
			fmt.Fprintf(&out, "static const adamic_shape %s_group_shape = {%d,%s_group_names,%s_group_references,NULL};\n", name, len(names), name, name)
		}
		bits, filter := p.nativeFirstASCII()
		prefix := p.nativePrefix()
		prefixName := "NULL"
		if len(prefix) > 0 {
			prefixName = name + "_prefix"
			fmt.Fprintf(&out, "static const uint16_t %s[] = {", prefixName)
			for _, c := range prefix {
				fmt.Fprintf(&out, "%d,", c)
			}
			out.WriteString("};\n")
		}
		fast := "NULL"
		ascii := "NULL"
		if body := p.nativeStraightLine(name); body != "" {
			out.WriteString(body)
			fast = name + "_fast"
		}
		if body := p.nativeStraightLineASCII(name); body != "" {
			out.WriteString(body)
			ascii = name + "_ascii"
		}
		behind := p.nativeLookbehindPrefix()
		behindName := "NULL"
		if len(behind) > 0 {
			behindName = name + "_before"
			fmt.Fprintf(&out, "static const uint16_t %s[] = {", behindName)
			for _, c := range behind {
				fmt.Fprintf(&out, "%d,", c)
			}
			out.WriteString("};\n")
		}
		firstBody, firstName := p.nativeSparseFirst(name)
		out.WriteString(firstBody)
		testName := "NULL"
		if body := p.nativeBooleanASCII(name, firstName); body != "" {
			out.WriteString(body)
			testName = name + "_test"
		}
		regularBody, regularName := p.nativeRegular(name)
		out.WriteString(regularBody)
		_, endAnchored := p.nativeBooleanWidth()
		testPrefix := testName != "NULL" && !p.nativeAnchored() && !p.flags.Sticky && len(prefix) != 0 && !endAnchored
		fmt.Fprintf(&out, "static const adamic_regex_program %s __attribute__((aligned(64))) = {%s_code,%s,%t,%d,%d,%d,%s,%d,%s,{UINT64_C(%d),UINT64_C(%d)},%t,%s,%d,%s,%t,%s,%s,%s,%d,%s};\n", name, name, testName, testPrefix, p.captures, p.repeats, nativeFlags(p.flags), groupNames, len(names), shape, bits[0], bits[1], filter, prefixName, len(prefix), fast, p.nativeAnchored(), ascii, regularName, behindName, len(behind), firstName)
		return nil
	}
	if err := emit(p, name); err != nil {
		return "", err
	}
	return out.String(), nil
}
func nativeFlags(f Flags) int {
	result := 0
	for bit, on := range []bool{f.IgnoreCase, f.Multiline, unicodeMode(f), f.Global, f.Sticky, f.HasIndices, f.UnicodeSets, f.DotAll} {
		if on {
			result |= 1 << bit
		}
	}
	return result
}
