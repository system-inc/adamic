# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 1 | 1 |
| Moved by reason | 0 | 0 |
| New findings | 0 | 0 |
| New sites | 0 | 0 |
| Units whose first stop moved | 0 | 0 |
| Regressions | 0 | 0 |
| before Boundary | 65 | 65 |
| before NotYet | 63 | 63 |
| before Refused | 52 | 52 |
| after Boundary | 65 | 65 |
| after NotYet | 63 | 63 |
| after Refused | 51 | 51 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| Array as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 | 1 | 1 | 1 | 0 |
| Object.defineProperties | 3 | 3 | 3 | 3 | 0 |
| Object.setPrototypeOf | 1 | 1 | 1 | 1 | 0 |
| a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 1 | 1 | 1 | 1 | 0 |
| a cast the runtime can't check | 11 | 11 | 11 | 11 | 0 |
| a checked predicate target whose check would read past the depth bound (FlowSwitchClause) | 1 | 1 | 1 | 1 | 0 |
| a function returning SortedReadonlyArray<[number, string]> | 1 | 1 | 1 | 1 | 0 |
| a function viewed as unknown or object (dynamic function descriptors) | 1 | 1 | 1 | 1 | 0 |
| a generic instantiation whose type argument for T, constrained to Node, can't be read | 6 | 6 | 6 | 6 | 0 |
| a library method value outside a const alias, typed call/apply, or supported map callback (its receiver and callable ABI are not proven); wrap the call in an arrow | 2 | 2 | 2 | 2 | 0 |
| a member typed any ([]) related to number at the initializer site | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (fill would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (repeat would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (setPrototypeOf would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (toString would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a non-object representation passed through an opaque object parameter | 4 | 4 | 4 | 4 | 0 |
| a template interpolating an object, an array, a map, a function or undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read | 4 | 4 | 4 | 4 | 0 |
| a value of type FlowArrayMutation \| FlowAssignment \| FlowCall \| FlowCondition \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowUnreachable seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 1 | 0 | 1 | 0 | 1 |
| a value of type Record<number, FlowGraphNode> | 1 | 1 | 1 | 1 | 0 |
| a value of type SortedReadonlyArray<[number, string]> | 1 | 1 | 1 | 1 | 0 |
| a value of type T | 1 | 1 | 1 | 1 | 0 |
| a value of type never | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"CheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof CheckMode is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"EmitFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof EmitFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"FlowFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof FlowFlags is read | 2 | 2 | 2 | 2 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"ModifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ModifierFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"NodeCheckFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeCheckFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"NodeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"ObjectFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ObjectFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"RelationComparisonResult", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof RelationComparisonResult is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"ScriptKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ScriptKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"SignatureCheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureCheckMode is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"SignatureFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"SnippetKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SnippetKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"SymbolFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SymbolFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"SyntaxKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SyntaxKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"TransformFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TransformFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"TypeFacts", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFacts is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("/workspace/scratch/99n-main/src/compiler/_namespaces/ts") seen as Record<"TypeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | 1 | 1 | 1 | 1 | 0 |
| an ESM namespace object used as a value; use qualified live bindings | 18 | 18 | 18 | 18 | 0 |
| an enum inside a function or block; declare it at module scope | 2 | 2 | 2 | 2 | 0 |
| an overloaded function as a value | 3 | 3 | 3 | 3 | 0 |
| an unproven value assigned to a numeric literal or enum member slot Connection | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.SourceFile | 1 | 1 | 1 | 1 | 0 |
| assigning to a NonNullExpression | 4 | 4 | 4 | 4 | 0 |
| failed statement; state rolled back; continued at next statement or declaration | 56 | 56 | 56 | 56 | 0 |
| function signature or prologue failed; continued at next declaration | 9 | 9 | 9 | 9 | 0 |
| in that narrows a union of object types (any object with the key passes) | 1 | 1 | 1 | 1 | 0 |
| in when a possibly absent spread has no property presence descriptors | 1 | 1 | 1 | 1 | 0 |
| inherited library member prototype read as an own field | 1 | 1 | 1 | 1 | 0 |
| inherited library member repeat read as an own field | 1 | 1 | 1 | 1 | 0 |
| new an Identifier | 2 | 2 | 2 | 2 | 0 |
| overload 1 of assertEachNode parameter test cannot be served by implementation parameter test | 1 | 1 | 1 | 1 | 0 |
| overload 1 of assertNode parameter test cannot be served by implementation parameter test | 1 | 1 | 1 | 1 | 0 |
| overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test | 1 | 1 | 1 | 1 | 0 |
| reading columns | 2 | 2 | 2 | 2 | 0 |
| reading graphNode | 3 | 3 | 3 | 3 | 0 |
| reading links | 1 | 1 | 1 | 1 | 0 |
| reading members | 3 | 3 | 3 | 3 | 0 |
| returning a property or element assignment | 1 | 1 | 1 | 1 | 0 |

## Raw Findings

Moved by reason:


New findings:

None.

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

## Distinct Reads

Moved by reason:


New findings:

None.

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

The JSON retains every lowered site, move, new finding, unit stop and panic.
