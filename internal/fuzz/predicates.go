package fuzz

import "fmt"

// Cadences cover the cross product without waiting for a rare random draw. Values
// still come from the seeded stream. Each scene is independent of the corpus.
func (g *generator) helperPredicateSource() string {
	value := g.pick("abc", "hello", "xy")
	switch g.seed % 4 {
	case 1:
		return fmt.Sprintf(`function predHelper(y?: unknown, x?: unknown): x is undefined { return typeof x === "undefined"; }
function predGuard(v: unknown): v is undefined { return predHelper(v); }
function predUse(v: string | undefined): void {
 if (predGuard(v)) { console.log("undefined? " + String(v)); const w: undefined = v; console.log(w === undefined ? "yes" : "no"); }
 else { console.log("string " + v.toUpperCase()); }
}
predUse(%q); predUse(undefined);`, value)
	case 2:
		return fmt.Sprintf(`function predHelper(y: unknown, x: unknown): x is string { return typeof x === "string"; }
function predGuard(v: unknown): v is string { return predHelper(0, v); }
function predUse(v: string | number): void { if (predGuard(v)) console.log("string " + v); else console.log("number " + String(v * 2)); }
predUse(%q); predUse(21);`, value)
	case 3:
		return fmt.Sprintf(`function predHelper(y?: unknown, x?: unknown): x is string { return typeof x === "string"; }
function predGuard(v: unknown): v is string { return predHelper(0, v); }
function predUse(v: string | number): void { if (predGuard(v)) console.log("string " + v); else console.log("number " + String(v * 2)); }
predUse(%q); predUse(21);`, value)
	default:
		return fmt.Sprintf(`function predHelper(y: unknown, x: unknown): x is string { return typeof x === "string"; }
function predGuard(v: unknown): v is string { return predHelper(v, 0); }
function predUse(v: string | number): void { if (predGuard(v)) console.log("string " + v); else console.log("number " + String(v * 2)); }
predUse(%q); predUse(21);`, value)
	}
}

func (g *generator) arrayPredicateSource() string {
	methods := []string{"filter", "find", "findIndex", "findLast", "some", "every"}
	scene := (g.seed - 1) % 36
	method := methods[scene%6]
	family := scene / 12
	inferred := scene/6%2 == 1
	element, target, values, condition := "string | number", "number", fmt.Sprintf(`["a", %d, "bc", 4]`, 2+g.random.IntN(4)), `typeof x === "number"`
	prefix := ""
	switch family {
	case 1:
		element, target, values, condition = "string | undefined", "string", `["a", undefined, "bc"]`, `x !== undefined`
	case 2:
		prefix = `type PredA = { readonly kind: "a"; readonly n: number }; type PredB = { readonly kind: "b"; readonly n: number };` + "\n"
		element, target, values, condition = "PredA | PredB", "PredA", `[{kind: "a", n: 2}, {kind: "b", n: 4}]`, `x.kind === "a"`
	}
	callback := "predIsTarget"
	if inferred {
		callback = "(x) => " + condition
	} else {
		prefix += fmt.Sprintf("function predIsTarget(x: %s): x is %s { return %s; }\n", element, target, condition)
	}
	prefix += fmt.Sprintf("const predItems: (%s)[] = %s;\nconst predResult = predItems.%s(%s);\n", element, values, method, callback)
	if method == "filter" {
		prefix += `console.log(String(predResult.length));` + "\n"
		if family == 2 {
			prefix += `for (const item of predResult) console.log(String(item.n));`
		} else {
			prefix += `console.log(String(predResult[0])); console.log(predResult.join(","));`
		}
	} else if family == 2 && (method == "find" || method == "findLast") {
		prefix += `console.log(predResult === undefined ? "missing" : String(predResult.n));`
	} else {
		prefix += `console.log(String(predResult));`
	}
	return prefix
}
