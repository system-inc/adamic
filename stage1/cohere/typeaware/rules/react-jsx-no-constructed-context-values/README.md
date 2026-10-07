This is a supplied-input judgment adapter, not an integrated source-to-rule port. `rule.json` lists typescript-go AST kind names and declares supplied-node visitation. `visit` receives its input directly; it does not traverse a file or refetch the entry node. Checker declaration fields are raw facts, supplied manually by positive controls while shared JSX source and checker-fact adapters are unavailable. Missing checker facts refuse rather than pretending to be resolved. The shared generator is not wired to this adapter API.

See ../react-jsx-fragments/REPORT.md for exact scope, independent production Go controls, compiling mutants, sanitizer results, options, limits and commands.


The numeric adapter now also judges parenthesized/as expressions, conditional and logical branches, assignments, property receivers, local variable/function declarations and identifier chains. Input.values is a flat raw tree; Input.valueIndex selects the value expression. Value.first/second are child indexes, Value.declarations contains raw declaration indexes in checker order, and Value.nearestFunction is a function identity. Input.tagIndex enables bare providers initialized by createContext or React.createContext. Missing declarations refuse explicitly. See REPORT.md for independent Go evidence and remaining gaps.

Class judgments now use Input.classFactsReady, Input.classIndex and Input.nearestFunctionParentKind. Raw class/heritage/type nodes are linked by Value.children and Value.first. Methods/accessors/constructors and property-bound arrows consult the enclosing class; its name does not determine component status.

Component function names now use the JS RegExp literal /^\p{Lu}/u, matching Go unicode.IsUpper across every code point. Quoted non-ASCII variable names in identifier messages remain a separate unsupported case.
