// Generated from pinned Go command flag declarations by testdata/definitions.go.
export interface Definition {
    readonly name: string;
    readonly kind: string;
    readonly value: string;
    readonly usage: string;
}
export const definitions: Definition[] = [
    {
        name: 'cache-dump',
        kind: 'Bool',
        value: 'false',
        usage: 'print what the cache table for this project holds, and exit',
    },
    {
        name: 'coverage',
        kind: 'Bool',
        value: 'false',
        usage: 'name every rule once under the coverage fact that describes it, rather than only counting them',
    },
    {
        name: 'directory',
        kind: 'String',
        value: '',
        usage: 'the project root, which every relative path resolves against (default: the directory of the nearest tsconfig.json or Package.swift at or above the working directory)',
    },
    {
        name: 'explain',
        kind: 'String',
        value: '',
        usage: 'report what every rule did on one file, and why it did or did not run, writing nothing',
    },
    {
        name: 'fix',
        kind: 'Bool',
        value: 'false',
        usage: 'apply fixes and format, running no other phase (--no-format to leave formatting out)',
    },
    {
        name: 'fix-passes',
        kind: 'Int',
        value: '10',
        usage: 'how many times a file may be re-linted while fixes keep landing',
    },
    {
        name: 'format',
        kind: 'Bool',
        value: 'false',
        usage: 'format the files not on record as formatted, or the paths named; a bare run, --fix and --no-fix already do, so naming it there changes nothing',
    },
    {
        name: 'format-all',
        kind: 'Bool',
        value: 'false',
        usage: 'format every file, not only the ones not on record as formatted (implies --format)',
    },
    {
        name: 'format-only',
        kind: 'Bool',
        value: 'false',
        usage: "format only, proposing no fixes and running no other phase (implies --format); with --no-fix, the commit gate's format check: every file formatting would change here or in a nested repository is a finding, and so is one the formatter could not read, and the exit is 0 only when there are none",
    },
    {
        name: 'json',
        kind: 'Bool',
        value: 'false',
        usage: 'print newline-delimited JSON for a program to read, described by schema/CohereOutput.schema.json',
    },
    {
        name: 'lint',
        kind: 'Bool',
        value: 'false',
        usage: "run the lint rules only, reporting their findings without fixing them and without TypeScript's diagnostics",
    },
    {
        name: 'lint-config',
        kind: 'String',
        value: 'CohereSettings.json',
        usage: 'the settings file that says which rules apply to which files, relative to --directory or else to where you typed it; unnamed, the file of this name at the project root',
    },
    {
        name: 'no-cache',
        kind: 'Bool',
        value: 'false',
        usage: "read nothing from and write nothing to this project's caches (the cache table and the tsconfig's incremental build info), so every phase computes from source",
    },
    {
        name: 'no-fix',
        kind: 'Bool',
        value: 'false',
        usage: "mutate no source in the checked project: report what would change, fixes and formatting both, without writing a byte of it, and exit nonzero if anything would (cohere still keeps its own cache in the project's .cache/cohere, unless --no-cache)",
    },
    {
        name: 'no-format',
        kind: 'Bool',
        value: 'false',
        usage: 'leave formatting out of a bare run, --fix or --no-fix: fix, type-check and lint only',
    },
    {
        name: 'phases',
        kind: 'Bool',
        value: 'false',
        usage: "put where the time went first in the footer's parentheses",
    },
    {
        name: 'print-config',
        kind: 'Bool',
        value: 'false',
        usage: "print, as JSON in ESLint's --print-config shape, every registered rule's resolved severity and options for one file (the path given, else index.ts at the project root), and exit",
    },
    {
        name: 'profile',
        kind: 'String',
        value: '',
        usage: 'write a Go CPU profile of the run to `file`, for go tool pprof',
    },
    {
        name: 'rules',
        kind: 'Bool',
        value: 'false',
        usage: "print the rules this cohere implements for the project's language, one per line, and exit",
    },
    {
        name: 'rules-enabled',
        kind: 'Bool',
        value: 'false',
        usage: 'print the rules the lint config resolves for one file (the path given, else index.ts at the project root), with severity, and exit',
    },
    { name: 'single-threaded', kind: 'Bool', value: 'false', usage: 'use one checker instead of several' },
    {
        name: 'stdin-filepath',
        kind: 'String',
        value: '',
        usage: "with --fix, read one file's text from stdin and print what --fix (and --format, if named) would write for the file at this path, writing nothing to disk",
    },
    {
        name: 'timing',
        kind: 'Bool',
        value: 'false',
        usage: 'report what building the graph cost and the CPU each rule cost, most expensive rule first',
    },
    {
        name: 'tsconfig',
        kind: 'String',
        value: 'tsconfig.json',
        usage: 'the tsconfig that defines the program, relative to --directory or else to where you typed it; unnamed, the file of this name in --directory, or else the nearest one at or above the working directory',
    },
    {
        name: 'types',
        kind: 'Bool',
        value: 'false',
        usage: "report TypeScript's own diagnostics only, running no rules and writing no fixes",
    },
    {
        name: 'unused',
        kind: 'Bool',
        value: 'false',
        usage: 'report code that was written and never used: unreferenced exports, and statements nothing can reach',
    },
    {
        name: 'unused-all',
        kind: 'Bool',
        value: 'false',
        usage: 'list the unused findings already marked cohere-keep rather than only counting them (implies --unused)',
    },
    {
        name: 'unused-deep',
        kind: 'Bool',
        value: 'false',
        usage: 'also compute the transitive closure of unused code and group it into islands (implies --unused)',
    },
    {
        name: 'verbose',
        kind: 'Bool',
        value: 'false',
        usage: 'print everything a run can say: each phase, the coverage summary, overrides, skips, notes, memory and the total',
    },
    {
        name: 'version',
        kind: 'Bool',
        value: 'false',
        usage: 'print the version, what this binary was built from, and the Swift contract it speaks, and exit',
    },
];
export const renameDefinitions: Definition[] = [
    {
        name: 'directory',
        kind: 'String',
        value: '',
        usage: "the working directory paths resolve against (default: the process's own)",
    },
    {
        name: 'dry-run',
        kind: 'Bool',
        value: 'false',
        usage: 'print the plan and write nothing (the default; accepted so it can be stated explicitly)',
    },
    { name: 'single-threaded', kind: 'Bool', value: 'false', usage: 'use one checker instead of several' },
    {
        name: 'tsconfig',
        kind: 'String',
        value: 'tsconfig.json',
        usage: 'the tsconfig that defines the program, relative to --directory',
    },
    { name: 'write', kind: 'Bool', value: 'false', usage: 'apply the rename; without it nothing is written' },
];
