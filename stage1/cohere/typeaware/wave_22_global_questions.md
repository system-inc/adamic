# Global provenance questions for wave 22's third batch

Both questions use the existing exact-node inspect ABI, length-framed UTF-8 fields and program ownership. No lint verdict is returned by Go. New Go and Adamic files are named for each question; the only shared edits are two one-line dispatcher registrations. The shared registration generator and harness are unchanged.

`global-binding\nnode` and `global-binding\nvalue` accept only an Identifier. After version/mode, they return two symbol-origin records: the compiler symbol and the identifier's local merged symbol. Each record is presence, declaration count, then per declaration source-file presence and declaration-file status. Value mode selects GetShorthandAssignmentValueSymbol for shorthand assignment names. Node mode preserves GetSymbolAtLocation. Missing symbols have presence false and count zero. Extra fields, unknown modes and wrong kinds are refused.

Native no-global-assign uses value mode and requires the first declaration's file to be a declaration file. Native no-implied-eval uses node mode, requires a symbol, chooses the larger local declaration list when present, and declines any source declaration. These intentionally different policies reproduce the production Go rules, including globalThis's symbol with no declarations and unresolved timer names being declined.

`global-source` accepts only SourceFile with no suffix. After version/mode it returns the compiler's external-module-indicator presence, JavaScript source classification, and resolved AlwaysStrict compiler option. Native no-implicit-globals determines directive prologues and class/function ancestry, then reports top-level nonlexical declarations or unresolved sloppy assignment targets. Compiler-classified modules and strict TypeScript files cannot leak implicit globals; JavaScript sources run as written.

The direct checker tests cover shorthand value identity, local shadows, unresolved symbols, module/script classification, JavaScript versus TypeScript and explicit strict false. Scratch overlay mutants prove malformed binding mode and source suffix refusal. Both question modes are tested after release; the registry-retention mutant must fail the required panic expectation.

Production judgments and messages reproduce cohere at 715ba94f3608a6500086b1076ce5cb7e51b836db. The existing MIT license in this directory applies. Default options match the pinned corpus runners; no fixes or suggestions are offered by these three Go rules, and their zero repair fields are still compared byte for byte.
