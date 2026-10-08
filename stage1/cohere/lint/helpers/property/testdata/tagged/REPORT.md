# NameTagged prerequisite

Branch lint-helpers/property-name-tagged, from origin/lint-helpers/property at
`a72b9f63e66263ea34419bae20d282968e3261b6`. Explicitly authorized to fill the
missing property.NameTagged before completing classmembers. Original Go pin
`7945d102a6c18dd36adf9114a758ce646e8b2359` is unchanged.

Production source: ../../name_tagged.a. It reuses propertyName and its accept-set
and computed-key judgment; it adds the upstream number/string tag based on the
settled parser node's kind. Numeric normalization remains a parser fact.
No duplicate shared Name implementation or runtime regex compiler is introduced.

[Selected log](selected.log): 1,213 calls, 12,448 bytes identical on source Node,
emitted JavaScript and ASan/UBSan native. All actual consumer calls are captured:
core no-dupe-class-members 101, TypeScript extension 55, React JSX spread-multi
83. Another 974 calls cover Go's property tests plus explicit nested computed,
parenthesized, private, unsupported and nil-name controls under all 64 masks.
Every capture suite passes without skips. Recursion is observed as additional
actual calls, not counted as distinct upstream rule cases.

The semantic mutant changes the numeric tag to string; it compiles, executes
and differs from original Go on all three runtimes. Selected package PASS in
57.850 seconds. Tests regenerate the observations from unchanged upstream
bodies through a Go overlay; no stale answer file is certified as fresh Go.
