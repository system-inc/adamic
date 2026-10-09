CODE UNDER TEST: Markdown TypeScript port, with artifact cache construction and witness helpers judged separately.
ORACLE: live Go cohere and pinned Prettier fork through Node; construction expectations are self.
Fixed menu selected before observing mutant results: M1 width ASCII constant, M2 heading condition, M3 whitespace mode option, M4 splitText separator constant. Four port mutants under the rebuild exception. Empty-answer probes P1 stringWidth, P2 splitText, P3 printDocument, P4 printWhitespace. Setup S1 artifact key loses sanitize option. Witness W1 textOutputDifference returns nil and W2 whitespaceLayoutBytesEqual returns true.

Before any mutant run, code inspection changed M4 from the outer space predicate 32 -> 31 to the newline classification constant 10 -> 9. The earlier plan could leave the inner loop unable to advance at U+001F. No result influenced this revision.

Construction edits S2/S3/S4 drop the complete sync.Once.Do construction statement while preserving the subsequent setup guards. They do not disable those guards. S1 changes only the artifact key sanitize option. W1/W2 explicitly weaken witness comparisons.
