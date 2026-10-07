# Structure component and NetworkService predicates

Three separate .a files port isLikelyReactComponent(nodes, index, looksLikeJsx), typeReferenceName(nodes, index), and isNetworkServiceHookCall(nodes, index, methods). model.a is an immutable flat projection of actual parser facts, carrying boolean classifications and direct child identities, not numeric kinds. No rule dispatch or registry is authored here.

isLikelyReactComponent accepts only FunctionDeclaration, FunctionExpression, ArrowFunction and MethodDeclaration. Its first parameter's identifier name properties or props wins before body inspection. Otherwise it examines only direct return statements in a block, or the expression body itself, via the supplied returnArgumentLooksLikeJsx predicate. The owned driver composes the proved batch32 helper against a parallel same-index JSX projection; it never supplies a Go boolean answer as the dependency. Nested returns, later parameters, destructured names and accessors do not gain acceptance. Nil is -1 and returns false.

typeReferenceName accepts only a TypeReference whose TypeName is an Identifier, and returns its exact text. Nil, other type shapes and qualified names return the empty string. It does not pick the last segment of a qualified name.

isNetworkServiceHookCall has the original precondition: the handed node is an actual CallExpression with valid callee/receiver edges. It unwraps parentheses around the callee, requires PropertyAccessExpression, unwraps the receiver, requires its Identifier spelling networkService, and looks up the property name in the supplied method map. Optional-chain flags do not change this helper's answer. Element access, asserted receivers, differently cased names and other objects do not match. Non-call inputs fail explicitly in the port rather than reaching invalid layout access. Go's mutable NetworkServiceHookMethods map is the explicit dependency: its initial four true entries are useGraphQlQuery, useGraphQlMutation, graphQlRequest and useSuspenseGraphQlQuery. Callers must preserve values, not just key membership; an existing false-valued key remains false. The tests add one false entry only in the oracle process and export that actual map to the driver, without changing any original function or source file.

The valid projection contract is acyclic, immutable parser nodes with valid indices and consistent classifications; -1 represents absent edges. Out-of-range indices fail explicitly. Malformed/cyclic projections and nil/non-call arguments to the network helper are outside the Go helper's domain. These helpers need no regex.

The oracle parses every string literal from every inventory consumer test file with pinned typescript-go. It queries every node and nil for component/type modes, and every actual call expression for network mode. Controls exercise first/later/destructured parameter names, methods/accessors, direct and nested returns, JSX/conditional/assertion boundaries, qualified and other type forms, all four network methods, false/missing map entries, optional calls, element accesses and callee/receiver parentheses through depth 64. Shape flags and edges are factual, with no predicted helper verdicts. Source Node, emitted JavaScript and sanitized native compare to unchanged Go.

With the setup environment sourced:

```
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch33 -count=1 -v -timeout=20m > /tmp/lint05-batch33-helpers.log 2>&1
```

Set ADAMIC_SLOT05_BATCH33_EVIDENCE to an existing directory for raw generated corpus/verdict/coverage files. Published evidence is compressed losslessly. This proves helper behavior on actual Go parser geometry; independent Adamic parsing and whole-rule findings remain integration work.
