//go:build ignore

package main

import (
	"fmt"
	"unicode"
)

func main() {
	fmt.Println("// Generated from Go unicode", unicode.Version, "by testdata/tables.go.")
	for _, name := range []string{"Mn", "Me", "Mc", "Cc", "Zl", "Zp", "Cf", "So", "Lo"} {
		fmt.Printf("const %s: readonly number[] = [", name)
		for _, r := range unicode.Categories[name].R16 {
			fmt.Printf("%d,%d,%d,", r.Lo, r.Hi, r.Stride)
		}
		for _, r := range unicode.Categories[name].R32 {
			fmt.Printf("%d,%d,%d,", r.Lo, r.Hi, r.Stride)
		}
		fmt.Println("];")
	}
	fmt.Println(`export function category(c: number, name: string): boolean {
 let rows: readonly number[] = [];
 switch(name) {case 'Mn': rows = Mn; break; case 'Me': rows = Me; break; case 'Mc': rows = Mc; break; case 'Cc': rows = Cc; break; case 'Zl': rows = Zl; break; case 'Zp': rows = Zp; break; case 'Cf': rows = Cf; break; case 'So': rows = So; break; case 'Lo': rows = Lo; break;}
 for(let i = 0; i < rows.length; i += 3) {
 const lo = rows[i] ?? 0; const hi = rows[i + 1] ?? 0; const stride = rows[i + 2] ?? 1;
 if(c < lo) { return false; }
 if(c <= hi && (c - lo) % stride === 0) { return true; }
 }
 return false;
}`)
}
