# Pipeline ordering stop

The user authorized amending 95h's two createProgram calls, conditional on no intermediate type checking or import resolution. Inspection finds intermediate semantic work, so the user's explicit stop condition applies. No amendment or definition rewrite was made.

stage3/apply.sh delegates directly to stage3/apply.py. In apply.py's adapt(), lines 314-325, each sorted adapter runs on the previous adapter's output, followed by a Git snapshot. The loop itself does not type-check. However, the invoked adapters do:

- 97-cli-axolotl-e/adapt.cjs:154-155 calls compiler(tree, ts) and program.getTypeChecker(), then reads program.ts and performs semantic fileState checks.
- Its shared compiler helper, 47-host-errors/adapt.cjs:19-37, constructs a TypeScript program from the current compiler tsconfig with Bundler module resolution.
- 98-axolotl-c/adapt.cjs:42-43 also constructs a compiler program and gets its checker.
- 99-cli-perf-hooks-import/adapt.cjs:145-155 inspects resolved module declarations and exports, and can request getSemanticDiagnostics on sys.ts and performanceCore.ts. Its plan constructs the compiler program at lines 244-245.

All of these sort after 95-checker-axolotl-h and before 99-correctness-core-multimap. Therefore the pipeline cannot be described as checking only the final tree. These findings establish the requested stop condition; they do not claim that a particular intermediate adapter would necessarily reject the missing free-function export. No such mutant apply was run.

Read model commit: 73cdea837469a7a9e27ff692e747421646a068dd, including its fold.cjs and 95h owner checks. The two createProgram edits remain assigned to 95h by the user's decision. The free-function definition change needs an adaptation before 95 under the user's stated ordering rule.

Validation status is unchanged from README.md: no new build, apply, lane count, census, or implementation was started in this continuation. No background process was started.
