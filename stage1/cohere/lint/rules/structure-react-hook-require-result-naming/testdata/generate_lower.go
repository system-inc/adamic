package main

import (
	"fmt"
	"unicode"
)

type point struct{ r, delta int }

func main() {
	var points []point
	for r := rune(0); r <= 0x10ffff; r++ {
		if lower := unicode.ToLower(r); lower != r {
			points = append(points, point{int(r), int(lower - r)})
		}
	}
	fmt.Println("// Generated from Go unicode.ToLower, Unicode " + unicode.Version + "; regenerate with testdata/generate_lower.go.\nconst ranges: readonly {readonly lo:number;readonly hi:number;readonly stride:number;readonly delta:number}[] = [")
	for i := 0; i < len(points); {
		start := points[i]
		stride := 1
		if i+1 < len(points) && points[i+1].delta == start.delta && points[i+1].r-start.r <= 2 {
			stride = points[i+1].r - start.r
		}
		last := start.r
		j := i + 1
		for j < len(points) && points[j].delta == start.delta && points[j].r == last+stride {
			last = points[j].r
			j++
		}
		fmt.Printf(" {lo:%d,hi:%d,stride:%d,delta:%d},\n", start.r, last, stride, start.delta)
		i = j
	}
	fmt.Println("];\nexport function goLower(text:string):string{\n let result='';\n for(let index=0;index<text.length;index++){\n  let point=text.codePointAt(index) ?? 0; if(point>65535){index++;}\n  for(const range of ranges){if(point<range.lo){break;} if(point<=range.hi && (point-range.lo)%range.stride===0){point+=range.delta;break;}}\n  result+=String.fromCodePoint(point);\n }\n return result;\n}")
}
