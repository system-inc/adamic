# Code and oracle, established before mutation runs

Starting main: 76c59c81e8617cea1892a01841494895927a712c. Audit starting main: f91994f019703ba25d2918cf529c0e0b0c05d93c. Every requested name exists; none moved to another file. Current list has 275 top-level tests. The full matrix includes additions since the audit.

Code under test is Adamic production Go lowering. No test, comparison harness, TypeScript checker, cohere rule implementation, Node, fixture, or expected value is mutated. Oracles for these selected assertions are self-written. External execution in other package rows remains intact.

| Row | Code under test | Oracle | Difference investigated |
|---|---|---|---|
| TestModuleNamespaceInitializedReadProof | Lower, proveModuleReads, checkedModuleRead, namespaceExpression | Executable output and console-write guard; absence of value ReferenceError in emitted C | Qualified ESM property access, whereas subsumer's explicit elimination checks use plain imported identifiers |
| TestCallableNamespaceReceiverStaysLoud | namespaceRefusal, namespaceOwnThis | Refused type and different receivers text | Function declaration merged with namespace, whereas arrow subsumer uses a namespace member function |
| TestNamespaceClassEarlyConstructionStaysLoud | namespaceInitialization, namespaceCallGraph.discover, namespaceCallable | NotYet and namespace initialization text | Construction via new and namespace class, rather than ordinary callable namespace export reads |
| TestNamespaceAmbientHostInitialization | namespaceInitialization, node host lowering, ownership analysis through Lower | Executable early-read refusal, nil ambient preflight errors, node: error text; cwd may accept independent cycle-capable Refused | Current assertion now catches the previously unchecked early-read precondition; host inputs have exclusive host and ownership coverage |
| TestNamespaceAmbientContextsDoNotExecute | namespaceRuntime | Ambient false, executable true, declaration-file false | Manually cleared inherited flag with declaration-file authority still present |
| TestNamespaceCallGraphLinearWork | namespaceCallGraph.reach/discover | Exactly depth+1 body expansions before and after repeat queries | Empty reach sets and duplicate acyclic edges; enum memoization and cycle union have nonempty sets |
| TestNamespaceCallGraphCycleUnion | namespaceCallGraph.reach/discover | Two specific namespace identities for all three cycle members and three expansions | Three-member namespace-only cycle, compared with two-member enum and namespace cycle; no exclusive covered line |
| TestNamespaceEnumInitializationIndependentOfModuleAnalysis | namespaceInitialization and shared reach graph | NotYet and enum initialization diagnostic | Calls namespace preflight directly, bypassing Lower's independent enum preflight |
| TestParserFactoryBindingHoisting | declareModule and namespaceBody | Two specific NamespaceVar names exactly once each and nonempty namespace body | Assertions added since audit; direct metadata observation rather than only generated output |
| TestTscNamespaceDeclarationShapes | Lower and namespace declaration/body handling | Self-written success/refusal outcomes, overload expectation derived from our own output | TypeScript-derived shapes, optional uninitialized locals, erased declarations and overload forms; no tsc execution |

The baseline passes in 41.705 seconds. Two whole rows skip for external project inputs, and one subcase skips for deferred implementation. Scope and complete baseline log retain the evidence. Missing external corpus trees are not toolchain installation problems.

Coverage compares each subsumed row with its named subsumer. Each formerly untrue row is compared with all other current top-level tests, explicitly excluding itself. All coverage commands passed. Profiles and commands are retained. Covered-line differences are inclusive line projections of Go coverage blocks, useful leads rather than path or column exclusivity proofs.

Seven mutants are the maximum for a full-package matrix near one minute under this brief. D01 through D06 prioritize distinct branches, inputs, or observable metadata. D07 tests the shared component union, also reached by early construction; it is a nonunique fallback candidate. Failure to establish a unique mutant within this cap does not authorize deletion.

Observed correction: the empty-reach distinction does not separate linear work from every other package row. TestEnumInitializationReach/call-graph-24 uses the same empty graph and reaches the 90-second timeout under D06. The target detects it at depth 12, but is not uniquely defended. D05 also reaches both TestEnumInitializationReach and TestNamespaceLimitsStayLoud, as namespace initialization now owns shared enum reachability.

Further observed correction: D06-enum-alone passes the expensive enum fixture within 90 seconds. The combined exclusion timeout is not an independent assertion failure. The complete enum row is rerun separately to resolve uniqueness; final rows.json and REPORT.md use that result, not the earlier provisional inference.
