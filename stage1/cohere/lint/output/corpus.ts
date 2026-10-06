import { panic, readTextFile, utf8Length } from 'adamic';
import { Parser } from '../../../typescript/parser/parser.ts';
import { Scanner } from '../../../typescript/scanner/scanner.ts';
import { Linter } from '../lint.ts';
import { Settings } from '../settings.ts';
import { Locations } from './location.ts';
import type { ChangedFile, RunFinding } from './model.ts';

export interface Corpus {
    readonly findings: RunFinding[];
    readonly changed: ChangedFile[];
    readonly files: number;
    readonly wouldChange: number;
}

// Fixes happen before findings, as in cohere's pipeline. A dry run counts files that would
// change and reports original findings. No source file is written by this rendering driver.
export function lintFiles(rows: readonly string[], fix: boolean, root: string): Corpus {
    const findings: RunFinding[] = [];
    const changed: ChangedFile[] = [];
    let wouldChange = 0;
    let files = 0;
    for (const row of rows) {
        if (row === '') { continue; }
        files++;
        const fields = row.split('\t');
        const path = fields[0] ?? panic('missing source path');
        const read = readTextFile(path);
        if (read.kind === 'Error') { panic(read.message); }
        const parser = new Parser(read.text, path);
        const scanner = new Scanner(read.text);
        const settings = new Settings();
        settings.load(fields[5] ?? '');
        const linter = new Linter(read.text, parser, scanner, fields[1] ?? 'all', fields[2] ?? '', fields[3] ?? '', fields[4] === 'true', settings);
        linter.run();
        const rewritten = linter.fixed();
        let reported = linter.findings;
        let source = read.text;
        if (rewritten !== read.text) {
            if (fix) {
                const relative = path.startsWith(root + '/') ? path.slice(root.length + 1) : path;
                changed.push({ path: relative, fixed: true, formatted: false, fixedBy: linter.fixesByRule });
                source = rewritten;
                const nextParser = new Parser(source, path);
                const nextScanner = new Scanner(source);
                const next = new Linter(source, nextParser, nextScanner, fields[1] ?? 'all', fields[2] ?? '', fields[3] ?? '', fields[4] === 'true', settings);
                next.run();
                reported = next.findings;
            } else { wouldChange++; }
        }
        const locations = new Locations(source);
        for (const finding of reported) {
            const location = locations.at(finding.start);
            findings.push({ path, line: location.line, column: location.column, severity: 'error', rule: finding.rule, messageId: finding.id, message: finding.message.replaceAll('\n', ' '), description: finding.message, start: location.byte, end: utf8Length(source.slice(0, finding.end)) });
        }
    }
    return { findings, changed, files, wouldChange };
}
