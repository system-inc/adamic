// What main.ts is asked: working trees with the paths to decide in each, pattern lists with theirs, and
// globs with the text to match; and how they are read from a cases file. The test (gitignore_test.go)
// writes one from cohere's tests and git's, and sample-cases.txt is a small one to run by hand.
//
// A cases file is lines of fields separated by tabs, the first field naming the record:
//
//	tree     <name> <root>             a working tree, which the records after it fill in
//	entry    <path> File <contents>    one of its entries: a file and what it holds,
//	entry    <path> SymbolicLink <target>
//	entry    <path> Directory          or a directory, or something else (Other)
//	patterns <name>                    a pattern list, which the records after it fill in
//	line     <text>                    one of its lines
//	query    <path> <0 or 1>           a path to decide in the tree or list above, 1 for a directory
//	glob     <pattern> <text> <0 or 1> a glob to match against text, 1 to match it as a path
//
// A field writes a backslash, a tab, a newline and a carriage return as \\, \t, \n and \r.

import { panic } from 'adamic';
import type { Entry } from './gitignore.ts';

// One path to decide, and whether it is a directory.
export interface Query {
    readonly path: string;
    readonly isDirectory: boolean;
}

// A working tree: its name for the output, the root error messages show, its entries, and its queries.
export interface TreeCase {
    readonly name: string;
    readonly root: string;
    readonly entries: ReadonlyMap<string, Entry>;
    readonly queries: readonly Query[];
}

// A list of ignore-file lines that is not a file in a tree, compiled as compilePatterns does, with the
// paths to decide by it.
export interface PatternsCase {
    readonly name: string;
    readonly lines: readonly string[];
    readonly queries: readonly Query[];
}

// One glob matched against one text, as a path or as a base name.
export interface GlobCase {
    readonly pattern: string;
    readonly text: string;
    readonly path: boolean;
}

// Everything one cases file asks.
export interface Cases {
    readonly trees: readonly TreeCase[];
    readonly patterns: readonly PatternsCase[];
    readonly globs: readonly GlobCase[];
}

// unescape reads a field's escapes.
function unescape(field: string): string {
    if(!field.includes('\\')) {
        return field;
    }
    // The runs between escapes, kept whole and joined once: a field can be 100 MiB, and natively each
    // += copies the string so far, so building it a character at a time would be quadratic.
    const pieces: string[] = [];
    let start = 0;
    let backslash = field.indexOf('\\');
    while(backslash >= 0) {
        pieces.push(field.slice(start, backslash));
        const escape = field[backslash + 1] ?? panic('cases: a field ending in a lone backslash');
        switch(escape) {
            case '\\':
                pieces.push('\\');
                break;
            case 't':
                pieces.push('\t');
                break;
            case 'n':
                pieces.push('\n');
                break;
            case 'r':
                pieces.push('\r');
                break;
            default:
                panic(`cases: an unknown escape \\${escape}`);
        }
        start = backslash + 2;
        backslash = field.indexOf('\\', start);
    }
    pieces.push(field.slice(start));
    return pieces.join('');
}

// field is a record's field at index, unescaped. A record without it is a malformed cases file.
function field(fields: readonly string[], index: number, lineNumber: number): string {
    return unescape(fields[index] ?? panic(`cases: line ${lineNumber} has no field ${index}`));
}

// flag is a field that is 0 or 1.
function flag(fields: readonly string[], index: number, lineNumber: number): boolean {
    const value = field(fields, index, lineNumber);
    if(value !== '0' && value !== '1') {
        panic(`cases: line ${lineNumber} has ${value} where 0 or 1 goes`);
    }
    return value === '1';
}

// entryOf is an entry record's entry.
function entryOf(fields: readonly string[], lineNumber: number): Entry {
    const kind = field(fields, 2, lineNumber);
    switch(kind) {
        case 'File':
            return { kind: 'File', contents: field(fields, 3, lineNumber) };
        case 'SymbolicLink':
            return { kind: 'SymbolicLink', target: field(fields, 3, lineNumber) };
        case 'Directory':
            return { kind: 'Directory' };
        case 'Other':
            return { kind: 'Other' };
    }
    return panic(`cases: line ${lineNumber} has an entry of kind ${kind}`);
}

// parseCases reads a cases file. A file that doesn't follow the format is a fault in whatever wrote
// it, and panics with the line that shows it.
export function parseCases(text: string): Cases {
    const trees: TreeCase[] = [];
    const patterns: PatternsCase[] = [];
    const globs: GlobCase[] = [];

    // What the records after a tree or a pattern list fill in: its entries, its lines and its queries.
    // section says which of the two is open, so a stray record is caught.
    let section: 'none' | 'tree' | 'patterns' = 'none';
    let entries = new Map<string, Entry>();
    let lines: string[] = [];
    let queries: Query[] = [];

    const records = text.split('\n');
    for(let index = 0; index < records.length; index++) {
        const record = records[index] ?? panic('cases: a line past the end');
        const lineNumber = index + 1;
        if(record === '') {
            continue;
        }
        const fields = record.split('\t');
        const kind = fields[0] ?? panic(`cases: line ${lineNumber} is empty`);
        switch(kind) {
            case 'tree': {
                entries = new Map<string, Entry>();
                queries = [];
                trees.push({
                    name: field(fields, 1, lineNumber),
                    root: field(fields, 2, lineNumber),
                    entries,
                    queries,
                });
                section = 'tree';
                break;
            }
            case 'entry': {
                if(section !== 'tree') {
                    panic(`cases: line ${lineNumber} is an entry outside a tree`);
                }
                entries.set(field(fields, 1, lineNumber), entryOf(fields, lineNumber));
                break;
            }
            case 'patterns': {
                lines = [];
                queries = [];
                patterns.push({ name: field(fields, 1, lineNumber), lines, queries });
                section = 'patterns';
                break;
            }
            case 'line': {
                if(section !== 'patterns') {
                    panic(`cases: line ${lineNumber} is a pattern line outside a pattern list`);
                }
                lines.push(field(fields, 1, lineNumber));
                break;
            }
            case 'query': {
                if(section === 'none') {
                    panic(`cases: line ${lineNumber} is a query outside a tree or a pattern list`);
                }
                queries.push({ path: field(fields, 1, lineNumber), isDirectory: flag(fields, 2, lineNumber) });
                break;
            }
            case 'glob': {
                globs.push({
                    pattern: field(fields, 1, lineNumber),
                    text: field(fields, 2, lineNumber),
                    path: flag(fields, 3, lineNumber),
                });
                break;
            }
            default:
                panic(`cases: line ${lineNumber} is a record of kind ${kind}`);
        }
    }
    return { trees, patterns, globs };
}
