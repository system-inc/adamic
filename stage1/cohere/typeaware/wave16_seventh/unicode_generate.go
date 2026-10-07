//go:build ignore

package main

import (
	"fmt"
	"strconv"
	"unicode"
)

func ranges(name string, accept func(rune) bool) {
	fmt.Printf("const %s: readonly number[] = [", name)
	begin := -1
	for at := 0; at <= 0x110000; at++ {
		held := at < 0x110000 && accept(rune(at))
		if held && begin < 0 {
			begin = at
		}
		if !held && begin >= 0 {
			fmt.Printf("%d,%d,", begin, at-1)
			begin = -1
		}
	}
	fmt.Println("];")
}
func main() {
	fmt.Printf("// Generated from Go unicode %s; oracle for capitalization and fmt %%q.\n", unicode.Version)
	ranges("upper", unicode.IsUpper)
	ranges("printable", strconv.IsPrint)
	fmt.Print(`
function contained(ranges: readonly number[], value: number): boolean {
    let low = 0, high = ranges.length / 2;
    while(low < high) {
        const mid = Math.floor((low + high) / 2);
        if((ranges[mid * 2 + 1] ?? -1) < value) low = mid + 1;
        else high = mid;
    }
    return low < ranges.length / 2 && (ranges[low * 2] ?? -1) <= value;
}
function point(text: string, index: number): number {
    const first = text.charCodeAt(index), second = text.charCodeAt(index + 1);
    return first >= 55296 && first <= 56319 && second >= 56320 && second <= 57343 ?
        65536 + (first - 55296) * 1024 + second - 56320 : first;
}
export function capital(text: string): boolean {return text !== '' && contained(upper, point(text, 0));}
function hexadecimal(value: number, width: number): string {
    let result = '', at = value;
    for(let index = 0; index < width; index++) {
        result = '0123456789abcdef'.slice(at % 16, at % 16 + 1) + result;
        at = Math.floor(at / 16);
    }
    return result;
}
export function quoted(text: string): string {
    let result = '"';
    for(let at = 0; at < text.length; at++) {
        const value = point(text, at), length = value > 65535 ? 2 : 1;
        if(value === 34) result += '\\"';
        else if(value === 92) result += '\\\\';
        else if(value === 7) result += '\\a';
        else if(value === 8) result += '\\b';
        else if(value === 9) result += '\\t';
        else if(value === 10) result += '\\n';
        else if(value === 11) result += '\\v';
        else if(value === 12) result += '\\f';
        else if(value === 13) result += '\\r';
        else if(contained(printable, value)) result += text.slice(at, at + length);
        else if(value < 32 || value === 127) result += '\\x' + hexadecimal(value, 2);
        else if(value <= 65535) result += '\\u' + hexadecimal(value, 4);
        else result += '\\U' + hexadecimal(value, 8);
        at += length - 1;
    }
    return result + '"';
}
`)
}
