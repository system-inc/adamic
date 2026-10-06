import { utf8Length } from 'adamic';
import { bold, compareText, counted, dim, red } from './model.ts';
import { fixedEven, seconds } from './footer.ts';

export interface Overall {
    readonly seconds: number;
    readonly projects: Map<string, number>;
    readonly failed: readonly string[];
    readonly unfinished: number;
    readonly notChecked: readonly string[];
    readonly root: string;
}

export function overallFooter(facts: Overall, color: boolean): string {
    let total = 0;
    const engines = [...facts.projects.keys()].sort(compareText);
    for (const engine of engines) { total += facts.projects.get(engine) ?? 0; }
    const verdict = facts.failed.length > 0 || facts.unfinished > 0 || total === 0 ? '✗ ☠️ ' : '✓ 💎 ';
    let result = bold(verdict + seconds(facts.seconds), color) + ' • '
        + (total === 0 ? red('nothing checked under ' + facts.root, color) : counted(total, 'project', 'projects'));
    if (engines.length > 1) { result += ': ' + engines.map((engine) => `${facts.projects.get(engine) ?? 0} ${engine}`).join(', '); }
    if (facts.failed.length > 0) { result += ' → ' + red('☠️ ' + facts.failed.join(', '), color); }
    if (facts.unfinished > 0) { result += ' • ⚠ ' + counted(facts.unfinished, 'project did not finish', 'projects did not finish'); }
    for (const gap of facts.notChecked) { result += ' • ⚠ ' + gap; }
    return result;
}

export function crashLine(path: string, rule: string, cause: string, color: boolean): string {
    const what = rule === ''
        ? `cohere crashed on this file (${cause}). This is a bug in cohere, not in your code: nothing in this file was checked.`
        : `cohere's rule ${rule} crashed on this file (${cause}). This is a bug in cohere, not in your code: the rule's verdict on this file is missing, and the file's other rules ran.`;
    const advice = ' Please report it at https://github.com/system-inc/cohere/issues with the output of `cohere --version`, the rule, the code that crashed it, and this message.';
    return path + ' ' + red('crash', color) + ' ' + dim(rule === '' ? 'cohere' : rule, color) + ' ' + what + advice;
}

export interface LegacyFinding {
    readonly path: string;
    readonly source: string | undefined;
    readonly start: number;
    readonly rule: string;
    readonly message: string;
}

export interface Coverage {
    readonly filesChecked: number;
    readonly rulesRun: number;
    readonly graphWarm: boolean;
    readonly seconds: number;
}

function legacyPath(finding: LegacyFinding): string { return finding.source === undefined ? '<unknown>' : finding.path; }

function legacyLocation(finding: LegacyFinding): string {
    if (finding.source === undefined || finding.start < 0 || finding.start > utf8Length(finding.source)) { return '0:0'; }
    let line = 1;
    let last = 0;
    let byte = 0;
    for (const point of finding.source) {
        if (byte >= finding.start) { break; }
        byte += utf8Length(point);
        if (point === '\n' && byte <= finding.start) { line++; last = byte; }
    }
    return `${line}:${finding.start - last + 1}`;
}

export function legacyReport(findings: readonly LegacyFinding[], coverage: Coverage): string {
    const ordered = [...findings].sort((left, right) => {
        const path = compareText(legacyPath(left), legacyPath(right));
        return path === 0 ? left.start - right.start : path;
    });
    let result = '';
    for (const finding of ordered) {
        result += `${legacyPath(finding)}:${legacyLocation(finding)}\n  ${finding.rule}  ${finding.message}\n\n`;
    }
    const duration = coverage.seconds < 1 ? `${Math.trunc(coverage.seconds * 1000)}ms` : fixedEven(coverage.seconds, 2) + 's';
    const detail = `${duration} · ${coverage.filesChecked} file${coverage.filesChecked === 1 ? '' : 's'} · ${coverage.rulesRun} rule${coverage.rulesRun === 1 ? '' : 's'} · graph ${coverage.graphWarm ? 'warm' : 'cold'}`;
    return result + (ordered.length === 0 ? `✓ cohere (${detail})\n` : `✗ cohere (${detail})  ${ordered.length} finding${ordered.length === 1 ? '' : 's'}\n`);
}

export function nothingChecked(reason: string): string { return '✗ cohere checked nothing: ' + reason + '\n'; }
