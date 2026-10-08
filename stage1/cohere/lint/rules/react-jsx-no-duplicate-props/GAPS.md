# Adjacent JSX recovery blocks certification

The directory is implemented but remains unpublished. All 37 ordinary upstream source/options cases, the owned witness selected/all rows, and the output mutant agree with Go on source Node, emitted JavaScript and sanitized native.

The remaining upstream case is `cohere/internal/lint/rules/react/jsx_no_duplicate_props_test.go:226`: the separate-element map control. TypeScript Go reports a parse diagnostic and recovers it; the capture carries recovery mode, so Go reports no duplicate properties. Stage 1 rejects at `stage1/typescript/parser/statements.ts:79` with `adamic: panic: parser slice expected semicolon at 10`. This is parser recovery, before the rule listener runs. Changing the rule cannot repair it, and the case has not been omitted or converted to passing input.

Local evidence: `/tmp/react-jsx-no-duplicate-props.log` (38 upstream cases, case 2 fails on the parser boundary; the other 37, witness and mutant pass). This branch requires a parser recovery dependency before it can meet the all-case parity bar and receive its one push.
