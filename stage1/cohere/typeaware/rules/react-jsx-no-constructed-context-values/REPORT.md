Built: numeric indirect-construction and bare-provider judgments for the existing claimed context rule.
Commits: see git log for this report; based on current main c01907a7, no new claims.
Checks: owned check_indirect.py and prior check.py pass native/Go bytes and sanitizer controls; setup 36s, nproc 5.
Mutants: branch-order, provider-factory and function-kind compile exit 0 and are caught solely by Go byte comparison; previous six adapter mutants still pass.
Uncovered: live source/checker integration, memo-stability/capture analysis, class ancestry, Unicode names and full native corpus comparisons.

The new .a helper follows production Go constructionOf over supplied raw numeric tree facts. It uses the same 16-hop bound, left-first alternatives, assignment relabeling and receiver usage. Identifiers follow the last declaration only when its nearest function identity equals the usage's function identity. Findings point to the underlying construction and use the four production message variants. Provider factory recognition iterates all declarations and unwraps parentheses around the initializer, callee and receiver. No shared registration generator, parser, driver, test harness or compiler file was changed. Existing numeric rule.json kinds remain 285 and 286.

Run from the repository with /workspace/adamic-tools/env.sh sourced:

    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_indirect.py
    python3 stage1/cohere/typeaware/rules/react-jsx-fragments/testdata/check.py

Both exited 0. New source specimens are rendered from owned .jsx-source files into /tmp/wave15-indirect for unmodified production Go. Native controls supply the raw facts manually, independently of Go rule decisions; this is adapter comparison, not native parsing. Fourteen positive findings match 4,986 canonical bytes under normal and sanitizer builds. The unrelated factory specimen produces no context finding. Findings contain no fixes or suggestions, matching production Go. Prior controls still match 17 findings, two option variants, three comparison mutants and three refusal mutants.

New mutants: conditional selection always takes the second branch (caught byte 693); bare factory identifier is changed to unrelatedContext (caught byte 4640); function declarations become object constructions (caught byte 3600). All compile and exit 0, so only byte comparison catches them. ASan/UBSan/LSan runs pass. No native checker handle is owned by these adapters, so released-handle safety was not retested as part of this extension.

Production Go panics with interface conversion *ast.SatisfiesExpression, not *ast.AsExpression on the preserved satisfies specimen, exit 2. Native explicitly refuses this shape, exit 70. This mismatch is recorded separately and is not counted as byte agreement. check_indirect.py reproduces both. The production Go file was not changed.

Observed timings: native fourteen-finding supplied-fact process 1.319 ms; Go source parsing/checker/all-three-rule process 57.074 ms. These workloads differ, so no comparable native-versus-Go speed ratio is claimed. Setup readiness lines: Go 0s, clang 0s, Node 0s, submodules 0s, build cache 36s, total 36s; nproc 5, CPU quota 4.

Blockers remain explicit. Current main lacks the published JSX parser. Published harness ab70f38d4 has JSX but ParseNode.kind is still a string and RuleContext exposes no live checker declaration/type/signature interface for this numeric adapter. Memo stability additionally needs raw union type flags, resolved signature declarations and bodies, reference/shorthand symbol identity, render-scope ancestry and escape/capture traversal. This helper does not implement that analysis and refuses unresolved constructions inside components. Class ancestry and Unicode classification/Go quoting also remain unsupported. Neither this rule nor the other two claimed JSX adapters is a complete integrated port. No new rules were claimed, and no full repository gate or full native corpus run was performed for this owned extension.


Class-component extension, 2026-10-07

Fetched every origin head. Main remains c01907a7 and is already an ancestor of this branch, so no rebase was necessary. This extension changes only this owned rule directory. No new claims or bridge questions were added.

Ported the production IsEs6ComponentClass predicate over raw numeric class, heritage and expression-with-type-arguments nodes. The nearest method/accessor/constructor/property belongs to its enclosing class; arrows or function expressions whose immediate parent is a property declaration also belong to that class. Bare Component/PureComponent bases and React-qualified bases count; parentheses around the namespace receiver are unwrapped. A class declaration and a class expression both count. The classifier follows Go's heritage iteration, including its absence of an extends-token gate. Missing raw ancestry refuses explicitly. This closes the previously listed class-judgment implementation gap, while the live ancestry adapter remains absent.

The expanded independent Go controls now contain 19 positive findings (6,726 canonical bytes), including class methods, a getter, a lowercase class name, a property-bound arrow and a class expression. Plain classes and Foo.Component decline. Normal and ASan/UBSan/LSan builds match. Four compiling, exit-zero mutants are caught only by Go byte comparison: alternative-order byte 2433, provider-factory byte 6380, class-owner byte 349, function-declaration-kind byte 5340. The class-owner mutant excludes class declarations while retaining class expressions, proving the new predicate independently of the earlier provider predicate. The prior three-rule suite still passes all 17 positive adapter findings, both option comparisons, three byte mutants and three exit-70 refusal mutants.

Exact commands remain the two Python scripts above. New logs: validation/components-check.log and validation/components-regression.log. Native supplied-fact controls took 1.369 ms; Go source/checker/all-three-rule process took 53.695 ms. These workloads differ; no corpus speed ratio is claimed. The previously prepared toolchain was reused (last setup 36s, nproc 5); no setup or full repository gate was rerun for this extension. No checker handles are owned by these controls, so no new released-handle evidence is claimed. The satisfies panic/refusal remains reproducible and outside successful byte comparisons.

Remaining work: shared numeric source nodes and live declaration/type/signature facts; memo stability and escape/capture traversal; Unicode component-name classification and Go quoted-name rendering. These three JSX claims remain partial integrated ports. No new batch was claimed while those gaps remain.
