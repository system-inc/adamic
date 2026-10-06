package regexp

import (
	"fmt"
	"strings"
)

func (p *Program) nativeRegular(name string) (string, string) {
	r := p.regular
	if r == nil {
		return "", "NULL"
	}
	var out strings.Builder
	fmt.Fprintf(&out, "static const adamic_regex_regular_instruction %s_regular_code[] = {", name)
	for _, i := range r.code {
		fmt.Fprintf(&out, "{%d,%d,%d,%d},", i.op, i.x, i.y, i.source)
	}
	out.WriteString("};\n")
	fmt.Fprintf(&out, "static const unsigned char %s_regular_classes[128] = {", name)
	for _, c := range r.classes {
		fmt.Fprintf(&out, "%d,", c)
	}
	out.WriteString("};\n")
	fmt.Fprintf(&out, "static const unsigned char %s_regular_characters[] = {", name)
	for _, c := range r.representatives {
		fmt.Fprintf(&out, "%d,", c)
	}
	out.WriteString("};\n")
	words := (len(r.code) + 63) / 64
	fmt.Fprintf(&out, "static const uint64_t %s_regular_masks[] = {", name)
	for _, c := range r.representatives {
		masks := make([]uint64, words)
		for pc, n := range r.code {
			if n.op == opSet {
				i := p.code[n.source]
				if i.set.contains(canonicalize(rune(c), i.flags)) {
					masks[pc/64] |= uint64(1) << uint(pc%64)
				}
			}
		}
		for _, mask := range masks {
			fmt.Fprintf(&out, "UINT64_C(%d),", mask)
		}
	}
	out.WriteString("};\n")
	fmt.Fprintf(&out, "static const adamic_regex_regular %s_regular = {%s_regular_code,%d,%d,%s_regular_classes,%s_regular_characters,%d,%s_regular_masks,%d};\n", name, name, len(r.code), r.start, name, name, len(r.representatives), name, words)
	return out.String(), "&" + name + "_regular"
}

// Tiny necessary first sets can reject ASCII input before UTF-16 conversion,
// and jump straight to a candidate in an existing one-pass runner.
func (p *Program) nativeSparseFirst(name string) (string, string) {
	bits, filter := p.nativeFirstASCII()
	if !filter || p.nativeAnchored() {
		return "", "NULL"
	}
	var characters []int
	for c := 0; c < 128; c++ {
		if bits[c/64]&(uint64(1)<<uint(c%64)) != 0 {
			characters = append(characters, c)
		}
	}
	if len(characters) > 4 {
		return "", "NULL"
	}
	var out strings.Builder
	out.WriteString("#include <string.h>\n")
	fmt.Fprintf(&out, "static ptrdiff_t %s_first(const unsigned char *input,size_t length,size_t start) {\n", name)
	out.WriteString("if(start>=length)return -1;(void)input;ptrdiff_t first=-1;\n")
	for _, c := range characters {
		fmt.Fprintf(&out, "{const unsigned char *hit=memchr(input+start,%d,length-start);if(hit!=NULL&&(first<0||hit-input<first))first=hit-input;}\n", c)
	}
	out.WriteString("return first;}\n")
	return out.String(), name + "_first"
}
