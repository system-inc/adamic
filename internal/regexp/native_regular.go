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
	if !p.nativeAnchored() {
		if body := p.nativeLiteralWindow(name); body != "" {
			return body, name + "_first"
		}
	}
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

// In an ASCII input every forward scalar set consumes one byte. A literal
// window after a fixed number of such sets is necessary even when the first
// set is broad. Stop before any variable-width or branching instruction.
func (p *Program) nativeLiteralWindow(name string) string {
	var best, run []rune
	offset, runOffset, bestOffset := 0, 0, 0
	for _, i := range p.code {
		switch i.op {
		case opSave, opAssert, opLook:
			continue
		case opSet:
			if i.direction != 1 || len(i.set.strings) != 0 {
				goto finished
			}
			if !i.flags.IgnoreCase && len(i.set.ranges) == 1 && i.set.ranges[0].From == i.set.ranges[0].To && i.set.ranges[0].To < 128 {
				if len(run) == 0 {
					runOffset = offset
				}
				run = append(run, i.set.ranges[0].From)
				if runOffset > 0 && len(run) > len(best) {
					best = run
					bestOffset = runOffset
				}
			} else {
				run = nil
			}
			offset++
		default:
			goto finished
		}
	}
finished:
	if len(best) < 2 {
		return ""
	}
	var out strings.Builder
	out.WriteString("#include <string.h>\n")
	fmt.Fprintf(&out, "static const unsigned char %s_window[]={", name)
	for _, c := range best {
		fmt.Fprintf(&out, "%d,", c)
	}
	fmt.Fprintf(&out, "};\nstatic ptrdiff_t %s_first(const unsigned char *input,size_t length,size_t start){\n", name)
	fmt.Fprintf(&out, "if(start>length||length-start<%d)return -1;size_t scan=start+%d;\n", bestOffset+len(best), bestOffset)
	fmt.Fprintf(&out, "while(length-scan>=%d){const unsigned char *hit=memchr(input+scan,%d,length-scan-%d+1);if(hit==NULL)return -1;scan=(size_t)(hit-input);\n", len(best), best[0], len(best))
	fmt.Fprintf(&out, "if(memcmp(hit,%s_window,%d)==0)return (ptrdiff_t)(scan-%d);scan++;}return -1;}\n", name, len(best), bestOffset)
	return out.String()
}
