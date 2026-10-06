import { bold, counted, dim, failed, grouped, red } from './model.ts';
import type { Readiness, RunSummary } from './model.ts';

// Go fmt rounds exact midpoint ties to even; Node toFixed rounds away from zero.
export function seconds(value: number): string {
    const digits = value < 0.01 ? 3 : value < 0.1 ? 2 : value < 10 ? 1 : 0;
    return fixedEven(value, digits) + 's';
}

export function fixedEven(value: number, digits: number): string {
    const sign = value < 0 ? '-' : '';
    const magnitude = Math.abs(value);
    const scale = 10 ** digits;
    const scaled = magnitude * scale;
    const lower = Math.floor(scaled);
    if (scaled - lower === 0.5 && Number.isInteger(magnitude * 2 ** (digits + 1)) && lower % 2 === 0) {
        return sign + (lower / scale).toFixed(digits);
    }
    return sign + magnitude.toFixed(digits);
}

function glyph(name: string): string {
    return name === 'fix' ? '🪄' : name === 'types' ? '🔷' : name === 'lint' ? '👑' : name === 'unused' ? '🧹' : '';
}

export function joinedOptions(names: readonly string[]): string {
    if (names.length === 1) { return names[0] ?? ''; }
    if (names.length === 2) { return `${names[0] ?? ''} and ${names[1] ?? ''}`; }
    let result = '';
    for (let index = 0; index < names.length; index++) {
        result += (index === names.length - 1 ? ', and ' : index > 0 ? ', ' : '') + (names[index] ?? '');
    }
    return result;
}

export function readinessSegment(summary: Readiness | undefined): string {
    if (summary === undefined) { return ''; }
    if (!summary.measured) { return 'Adamic readiness not measured: ' + summary.reason; }
    if (summary.files === 0) { return 'Adamic readiness not measured: no file was linted'; }
    let result = `${Math.trunc(summary.ready * 100 / summary.files)}% Adamic-ready (${grouped(summary.ready)} of ${grouped(summary.files)}`;
    if (summary.unmeasured > 0) { result += '; ' + grouped(summary.unmeasured) + ' unmeasured'; }
    if (summary.optionsMissing.length > 0) { result += '; tsconfig lacks ' + joinedOptions(summary.optionsMissing); }
    return result + ')';
}

function phaseTimes(summary: RunSummary): string[] {
    const result: string[] = [];
    if (summary.graphSeconds > 0) { result.push('🕸 ' + seconds(summary.graphSeconds)); }
    for (const name of ['fix', 'types', 'lint', 'unused']) {
        for (const phase of summary.phases) {
            if (phase.name !== name) { continue; }
            if (phase.outcome === 'ran' || phase.outcome === 'checked') {
                const elapsed = name === 'fix' ? Math.max(phase.seconds - summary.formattingSeconds, 0) : phase.seconds;
                result.push(glyph(name) + ' ' + seconds(elapsed));
            } else if (phase.outcome === 'reused') { result.push(glyph(name) + ' in 🪄'); }
        }
        if (name === 'fix' && summary.formattingSeconds > 0) { result.push('💅 ' + seconds(summary.formattingSeconds)); }
    }
    return result;
}

function uncheckedMarkers(summary: RunSummary): string[] {
    const gaps = summary.gaps;
    const result: string[] = [];
    if (gaps.nothingToCheck) { result.push('⚠ no files to check'); }
    if (gaps.programFiles > summary.filesInScope && summary.filesInScope > 0) { result.push(`⚠ only ${grouped(summary.filesInScope)} of ${grouped(gaps.programFiles)} files`); }
    for (const name of ['fix', 'types', 'lint', 'unused']) {
        for (const phase of summary.phases) {
            if (phase.name !== name) { continue; }
            if (phase.outcome === 'not reached') { result.push(`⚠ ${glyph(name)} ${name} did not run`); }
            else if (phase.outcome === 'skipped' && name !== 'fix' && name !== 'unused') { result.push(`⚠ ${glyph(name)} ${name} not checked`); }
        }
    }
    if (gaps.formattingNotChecked) { result.push('💅 formatting not checked'); }
    if (gaps.crashedFiles > 0) { result.push('⚠ ' + counted(gaps.crashedFiles, 'file crashed', 'files crashed')); }
    if (gaps.rulesSkippingEverything > 0) { result.push('⚠ ' + counted(gaps.rulesSkippingEverything, 'rule skipped every file', 'rules skipped every file')); }
    if (gaps.unread !== '') { result.push('⚠ ' + gaps.unread); }
    if (gaps.modifiedBuild) { result.push('⚠ built from a modified tree'); }
    return result;
}

export function footer(summary: RunSummary, color: boolean, phases: boolean, verbose: boolean): string {
    let result = bold((failed(summary) ? '✗ ☠️ ' : '✓ 💎 ') + seconds(summary.seconds), color);
    if (summary.typeErrors > 0) { result += ' • ' + red(counted(summary.typeErrors, 'type error', 'type errors'), color); }
    if (summary.findings > 0) { result += ' • ' + red(counted(summary.findings, 'finding', 'findings'), color); }
    if (summary.wouldChange > 0) { result += ' • ' + red(counted(summary.wouldChange, 'file would change', 'files would change'), color); }
    if (summary.changed.length > 0) { result += ' • ' + grouped(summary.changed.length) + ' cohered'; }
    if (verbose && summary.cache.replayed) { result += ' • replayed'; }
    const inside: string[] = phases && !summary.cache.replayed ? phaseTimes(summary) : [];
    if (summary.rules > 0) { inside.push(counted(summary.rules, 'rule', 'rules')); }
    if (summary.checked > 0) { inside.push(grouped(summary.checked) + ' checked'); }
    if (summary.cached > 0) { inside.push(grouped(summary.cached) + ' cached'); }
    if (verbose && summary.nodes > 0) { inside.push(counted(summary.nodes, 'node', 'nodes')); }
    if (inside.length > 0) { result += ' ' + dim('(' + inside.join(' • ') + ')', color); }
    const readiness = readinessSegment(summary.adamic);
    if (readiness !== '') { result += ' • ' + readiness; }
    const markers = uncheckedMarkers(summary);
    if (markers.length > 0) { result += ' • ' + markers.join(' • '); }
    return (summary.label === '' ? '' : summary.label + '  ') + result;
}
