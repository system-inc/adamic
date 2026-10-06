import { panic } from 'adamic';
import { compareText, dim, orderedFindings, red } from './model.ts';
import type { ChangedFile, Readiness, RunFinding, RunSummary } from './model.ts';
import { footer } from './footer.ts';

export function quote(text: string): string {
    return (JSON.stringify(text) ?? panic('cannot encode a string'))
        .replaceAll('<', '\\u003c').replaceAll('>', '\\u003e').replaceAll('&', '\\u0026')
        .replaceAll('\u2028', '\\u2028').replaceAll('\u2029', '\\u2029');
}

function singleLine(text: string): string { return text.replaceAll('\n', ' '); }

export function findingLine(finding: RunFinding, color: boolean): string {
    const severity = finding.severity === 'error' ? red(finding.severity, color) : dim(finding.severity, color);
    const location = finding.path === '' ? '' : `${finding.path}:${finding.line}:${finding.column} `;
    return location + severity + ' ' + dim(finding.rule, color) + ' ' + singleLine(finding.message);
}

export function verboseFinding(finding: RunFinding): string {
    if (finding.path === '') { return `error ${finding.rule}: ${finding.message}`; }
    return `${finding.path}:${finding.line}:${finding.column} - ${singleLine(finding.message)} [${finding.rule}/${finding.messageId}]`;
}

export function findingJSON(finding: RunFinding): string {
    return `{"kind":"finding","path":${quote(finding.path)},"line":${finding.line},"column":${finding.column},"severity":${quote(finding.severity)},"rule":${quote(finding.rule)}`
        + (finding.messageId === '' ? '' : ',"messageId":' + quote(finding.messageId))
        + ',"message":' + quote(finding.message) + '}';
}

export function fixingRules(fixedBy: Map<string, number>): string {
    const names = [...fixedBy.keys()].sort((left, right) => {
        const difference = (fixedBy.get(right) ?? 0) - (fixedBy.get(left) ?? 0);
        return difference === 0 ? compareText(left, right) : difference;
    });
    return names.map((name) => name + ((fixedBy.get(name) ?? 0) > 1 ? ` ×${fixedBy.get(name) ?? 0}` : '')).join(', ');
}

export function changedFileLines(files: readonly ChangedFile[], color: boolean): string[] {
    const sorted = [...files].sort((left, right) => compareText(left.path, right.path));
    const shown = sorted.slice(0, 20);
    let width = 0;
    for (const file of shown) { width = Math.max(width, [...file.path].length); }
    const lines: string[] = [];
    for (const file of shown) {
        let line = (file.fixed ? '🪄' : '  ') + (file.formatted ? '💅' : '  ') + ' ' + file.path;
        const rules = fixingRules(file.fixedBy);
        if (rules !== '') { line += ' '.repeat(width - [...file.path].length) + ' ' + dim(rules, color); }
        lines.push(line);
    }
    if (sorted.length > shown.length) { lines.push(`… ${sorted.length - shown.length} more · --verbose for all`); }
    return lines;
}

function stringArray(values: readonly string[]): string { return '[' + values.map(quote).join(',') + ']'; }

function numberMap(values: Map<string, number>): string {
    return '{' + [...values.keys()].sort(compareText).map((name) => `${quote(name)}:${values.get(name) ?? 0}`).join(',') + '}';
}

export function changedFilesJSON(files: readonly ChangedFile[]): string[] {
    const lines: string[] = [];
    for (const file of files) {
        if (file.fixed) {
            const names = [...file.fixedBy.keys()].sort(compareText);
            lines.push('{"kind":"fixed","path":' + quote(file.path) + (names.length === 0 ? '' : ',"rules":' + stringArray(names)) + '}');
        }
        if (file.formatted) { lines.push('{"kind":"formatted","path":' + quote(file.path) + '}'); }
    }
    return lines;
}

function readinessJSON(summary: Readiness | undefined): string {
    if (summary === undefined) {
        return '{"measured":false,"reason":"lint did not run","files":0,"ready":0,"unmeasured":0,"optionsMissing":[],"failingFilesByRule":{}}';
    }
    return `{"measured":${summary.measured}` + (summary.reason === '' ? '' : ',"reason":' + quote(summary.reason))
        + `,"files":${summary.files},"ready":${summary.ready},"unmeasured":${summary.unmeasured},"optionsMissing":`
        + stringArray(summary.optionsMissing) + ',"failingFilesByRule":' + numberMap(summary.failingFilesByRule) + '}';
}

export function summaryJSON(summary: RunSummary): string {
    let fixed = 0;
    let formatted = 0;
    for (const file of summary.changed) { if (file.fixed) { fixed++; } if (file.formatted) { formatted++; } }
    const phases = summary.phases.map((phase) => `{"name":${quote(phase.name)},"outcome":${quote(phase.outcome)},"seconds":${phase.seconds}`
        + (phase.detail === '' ? '' : ',"detail":' + quote(phase.detail)) + '}');
    const cache = summary.cache;
    const gaps = summary.gaps;
    return '{"kind":"summary","schemaVersion":1' + (summary.label === '' ? '' : ',"label":' + quote(summary.label))
        + ',"verdict":' + quote(summary.typeErrors > 0 || summary.findings > 0 || summary.wouldChange > 0 || gaps.crashedFiles > 0 || gaps.ruleCrashes > 0 ? 'fail' : 'pass')
        + `,"seconds":${summary.seconds},"graphSeconds":${summary.graphSeconds},"rules":${summary.rules},"phases":[${phases.join(',')}]`
        + (summary.formattingSeconds === 0 ? '' : `,"formattingSeconds":${summary.formattingSeconds}`)
        + `,"cache":{"replayed":${cache.replayed},"filesReplayed":${cache.filesReplayed},"off":${cache.off}`
        + (cache.missReason === '' ? '' : ',"missReason":' + quote(cache.missReason))
        + (cache.findingsMissReason === '' ? '' : ',"findingsMissReason":' + quote(cache.findingsMissReason)) + '}'
        + `,"typeErrors":${summary.typeErrors},"findings":${summary.findings},"filesFixed":${fixed},"filesFormatted":${formatted},"wouldChange":${summary.wouldChange}`
        + `,"cohered":${summary.changed.length},"filesInScope":${summary.filesInScope},"checked":${summary.checked},"cached":${summary.cached},"nodes":${summary.nodes}`
        + `,"gaps":{"programFiles":${gaps.programFiles},"crashedFiles":${gaps.crashedFiles},"ruleCrashes":${gaps.ruleCrashes},"rulesSkippingEverything":${gaps.rulesSkippingEverything}`
        + `,"formattingNotChecked":${gaps.formattingNotChecked},"nothingToCheck":${gaps.nothingToCheck},"modifiedBuild":${gaps.modifiedBuild}`
        + (gaps.unread === '' ? '' : ',"unread":' + quote(gaps.unread)) + '},"adamic":' + readinessJSON(summary.adamic) + '}';
}

export function renderFindings(findings: readonly RunFinding[], mode: string, color: boolean): string[] {
    return orderedFindings(findings).map((finding) => mode === 'json' ? findingJSON(finding) : mode === 'verbose' ? verboseFinding(finding) : findingLine(finding, color));
}

export function renderSummary(summary: RunSummary, mode: string, color: boolean, phases: boolean): string {
    return mode === 'json' ? summaryJSON(summary) : footer(summary, color, phases, mode === 'verbose');
}

export function crashJSON(path: string, rule: string, cause: string): string {
    return '{"kind":"crash","path":' + quote(path) + (rule === '' ? '' : ',"rule":' + quote(rule)) + ',"cause":' + quote(cause) + '}';
}
