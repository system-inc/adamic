package lower

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// An overload whose promised result is a phantom extension of its implementation's is an overload mint,
// held to a brand mint's rule (brandMintOf): accepted in .ts with its own line in the checked report naming
// the overload, refused in .a at each call that resolves to it (Adamic's ruling on #z1qkw7d). A primitive
// brand, at the top or on an array's elements, stays erased with no line in either.
//
// Mutants (overload_mint.mutants.md): m1, the report line dropped from overloadPhantomResult, and both .ts
// rows of TestOverloadMintIsListedInTs go red on their missing line; m2, the extension served by its one
// base without brandMintOf, and TestOverloadMintKeepsRefusingARealMember goes red.

// overloadMintSorted is an array brand on the array itself, Sorted<number> over number[], whose returns are
// array literals, so the fresh-array condition holds. Overload 1 is at line 2 and the call at 10:14.
const overloadMintSorted = `interface Sorted<T> extends Array<T> { __sorted: undefined }
function sorted(a: number, b: number, sure: true): Sorted<number>;
function sorted(a: number, b: number, sure: boolean): number[];
function sorted(a: number, b: number, sure: boolean): number[] {
    if (a <= b) {
        return [a, b];
    }
    return [b, a];
}
const pair = sorted(2, 1, true);
console.log(pair.join(","));
`

// overloadMintPathComponents is tsc's getPathComponents shape (path.ts:493) over overloadPhantomPrelude: the
// extension's base is Array<Path> over an implementation's string[], so the element's primitive brand comes
// off first and the extension is the mint. Overload 1 is at line 11 and the call at 17:15. tsc may print the
// brand with its defaulted argument, so the rows below read its name as a prefix.
const overloadMintPathComponents = overloadPhantomPrelude + `function pathComponents(path: string, rootLength: number): string[] {
    const root = path.substring(0, rootLength);
    const rest = path.substring(rootLength).split("/");
    if (rest.length > 0 && rest[rest.length - 1] === "") {
        rest.pop();
    }
    return [root, ...rest];
}
function getPathComponents(path: Path): PathPathComponents;
function getPathComponents(path: string, currentDirectory?: string): string[];
function getPathComponents(path: string, currentDirectory = ""): string[] {
    path = currentDirectory + path;
    return pathComponents(path, path.startsWith("/") ? 1 : 0);
}
const parts = getPathComponents("/a/b/" as Path);
console.log(parts.join("|"));
`

// overloadMintLines is the report's overload mint records, each as where, target, members and overload.
func overloadMintLines(program *ir.Program) []ir.BrandMint {
	var mints []ir.BrandMint
	for _, mint := range program.BrandMints {
		if mint.Overload != "" {
			mints = append(mints, mint)
		}
	}
	return mints
}

// In .ts each mint lowers, agrees with Node, and has one line in the checked report naming its overload,
// its brand and its member, and no cast's line besides.
func TestOverloadMintIsListedInTs(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		name, source string
		want         ir.BrandMint
	}{
		{"sorted", overloadMintSorted, ir.BrandMint{Where: "main.ts:2:1", Target: "Sorted<number>", Members: []string{"__sorted"}, Overload: "overload 1 of sorted"}},
		{"getPathComponents", overloadMintPathComponents, ir.BrandMint{Where: "main.ts:11:1", Target: "PathPathComponents", Members: []string{"__pathComponensBrand"}, Overload: "overload 1 of getPathComponents"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, row.source, ".ts")
			mints := overloadMintLines(program)
			if len(program.BrandMints) != 1 || len(mints) != 1 {
				t.Fatalf("want one overload mint line, got %+v", program.BrandMints)
			}
			got := mints[0]
			if !strings.HasSuffix(got.Where, row.want.Where) || !strings.HasPrefix(got.Target, row.want.Target) || !slices.Equal(got.Members, row.want.Members) || got.Overload != row.want.Overload {
				t.Fatalf("got %+v, want %+v", got, row.want)
			}
		})
	}
}

// In .a every mint is refused for now (#yth65gz), so each stops at its call, in the overload mint's words.
func TestOverloadMintIsRefusedInA(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		name, source, where string
		what                []string
	}{
		{"sorted", overloadMintSorted, "main.a:10:14", []string{"overload 1 of sorted result Sorted<number>", ", a brand mint in .a: an overload result that adds only its phantom member __sorted to implementation result number[]"}},
		{"getPathComponents", overloadMintPathComponents, "main.a:17:15", []string{"overload 1 of getPathComponents result PathPathComponents", ", a brand mint in .a: an overload result that adds only its phantom member __pathComponensBrand to implementation result string[]"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, row.source, ".a")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasPrefix(refused.What, row.what[0]) || !strings.HasSuffix(refused.What, row.what[1]) || !strings.HasSuffix(refused.Where, row.where) || !strings.HasPrefix(refused.Fix, "mint the brand in a .ts file") {
				t.Fatalf("want %q refused at %s, got %v", row.what, row.where, err)
			}
		})
	}
}

// An extension with a real member isn't a mint (brandMintOf), so its overload stays refused at the call in
// the census's own words, though its base is the implementation's result. In .a, where a mint's words differ
// from the census's, so m2 (served by its one base, with no members) turns this to the mint's refusal. An
// array extension with a real member stops earlier, at its representation, so the base here is an object.
func TestOverloadMintKeepsRefusingARealMember(t *testing.T) {
	t.Parallel()
	source := `interface Box<T> { readonly value: T }
interface Named<T> extends Box<T> { readonly name: string }
function tag(box: Box<number>, sure: true): Named<number>;
function tag(box: Box<number>, sure: boolean): Box<number>;
function tag(box: Box<number>, sure: boolean): Box<number> { return box; }
console.log(String(tag({ value: 1 }, true).value));
`
	_, err := lowerSourceAs(t, source, ".a")
	var refused *Refused
	const want = "overload 1 of tag result Named<number> cannot be served by implementation result Box<number>"
	if !errors.As(err, &refused) || !strings.Contains(refused.What, want) || !strings.Contains(refused.Where, "main.a:6:") {
		t.Fatalf("want %q refused at the call, got %v", want, err)
	}
}

// A primitive brand is no overload mint: the bare Path overload lowers in .ts and .a with no report line,
// as a primitive brand's cast does (rulings.md:30).
func TestOverloadMintLeavesAPrimitiveBrand(t *testing.T) {
	t.Parallel()
	source := `type Path = string & { readonly __pathBrand: undefined };
function f(p: Path): Path;
function f(p: string): string;
function f(p: string): string {
    return p + "/";
}
const p = "a" as Path;
console.log(f(p));
`
	for _, extension := range []string{".ts", ".a"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, source, extension)
			if err != nil {
				t.Fatal(err)
			}
			if len(program.BrandMints) != 0 {
				t.Fatalf("a primitive brand was listed as a mint: %+v", program.BrandMints)
			}
		})
	}
}
