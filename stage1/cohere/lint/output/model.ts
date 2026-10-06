export interface RunFinding {
    readonly path: string;
    readonly line: number;
    readonly column: number;
    readonly severity: string;
    readonly rule: string;
    readonly messageId: string;
    readonly message: string;
    readonly description: string;
    readonly start: number;
    readonly end: number;
}

export interface ChangedFile {
    readonly path: string;
    readonly fixed: boolean;
    readonly formatted: boolean;
    readonly fixedBy: Map<string, number>;
}

export interface Phase {
    readonly name: string;
    readonly outcome: string;
    readonly seconds: number;
    readonly detail: string;
    readonly findings: number;
    readonly narrowed: string;
}

export interface CacheUse {
    readonly replayed: boolean;
    readonly filesReplayed: number;
    readonly off: boolean;
    readonly missReason: string;
    readonly findingsMissReason: string;
}

export interface RunGaps {
    readonly programFiles: number;
    readonly crashedFiles: number;
    readonly ruleCrashes: number;
    readonly rulesSkippingEverything: number;
    readonly formattingNotChecked: boolean;
    readonly nothingToCheck: boolean;
    readonly modifiedBuild: boolean;
    readonly unread: string;
}

export interface Readiness {
    readonly measured: boolean;
    readonly reason: string;
    readonly files: number;
    readonly ready: number;
    readonly unmeasured: number;
    readonly optionsMissing: readonly string[];
    readonly failingFilesByRule: Map<string, number>;
}

// Durations are captured run facts in seconds. A renderer never samples a different clock
// and claims it is the original run's measurement.
export interface RunSummary {
    readonly label: string;
    readonly seconds: number;
    readonly graphSeconds: number;
    readonly rules: number;
    readonly phases: readonly Phase[];
    readonly formattingSeconds: number;
    readonly cache: CacheUse;
    readonly typeErrors: number;
    readonly findings: number;
    readonly changed: readonly ChangedFile[];
    readonly wouldChange: number;
    readonly filesInScope: number;
    readonly checked: number;
    readonly cached: number;
    readonly nodes: number;
    readonly gaps: RunGaps;
    readonly adamic: Readiness | undefined;
}

export function failed(summary: RunSummary): boolean {
    return summary.typeErrors > 0 || summary.findings > 0 || summary.wouldChange > 0
        || summary.gaps.crashedFiles > 0 || summary.gaps.ruleCrashes > 0;
}

export function exitCode(summary: RunSummary): number { return failed(summary) ? 1 : 0; }

// Go orders UTF-8 strings by bytes; JavaScript's default sort compares UTF-16 units.
export function compareText(left: string, right: string): number {
    let a = 0;
    let b = 0;
    while (a < left.length && b < right.length) {
        const x = left.codePointAt(a) ?? 0;
        const y = right.codePointAt(b) ?? 0;
        if (x !== y) { return x < y ? -1 : 1; }
        a += x > 65535 ? 2 : 1;
        b += y > 65535 ? 2 : 1;
    }
    return a === left.length ? (b === right.length ? 0 : -1) : 1;
}

export function compareFinding(left: RunFinding, right: RunFinding): number {
    const path = compareText(left.path, right.path);
    if (path !== 0) { return path; }
    if (left.start !== right.start) { return left.start - right.start; }
    if (left.end !== right.end) { return left.end - right.end; }
    const rule = compareText(left.rule, right.rule);
    if (rule !== 0) { return rule; }
    const id = compareText(left.messageId, right.messageId);
    if (id !== 0) { return id; }
    return compareText(left.description, right.description);
}

export function orderedFindings(findings: readonly RunFinding[]): RunFinding[] {
    return [...findings].sort(compareFinding);
}

export function wrap(text: string, color: boolean, open: string, close: string): string {
    return color && text !== '' ? open + text + close : text;
}

export function dim(text: string, color: boolean): string { return wrap(text, color, '\u001b[2m', '\u001b[22m'); }
export function red(text: string, color: boolean): string { return wrap(text, color, '\u001b[31m', '\u001b[39m'); }
export function bold(text: string, color: boolean): string { return wrap(text, color, '\u001b[1m', '\u001b[22m'); }

export function colorForStdout(): boolean {
    return process.env.NO_COLOR === undefined && process.stdout.isTTY === true;
}

export function grouped(count: number): string {
    const digits = `${Math.abs(count)}`;
    let result = count < 0 ? '-' : '';
    for (let index = 0; index < digits.length; index++) {
        if (index > 0 && (digits.length - index) % 3 === 0) { result += ','; }
        result += digits[index] ?? '';
    }
    return result;
}

export function counted(count: number, singular: string, plural: string): string {
    return grouped(count) + ' ' + (count === 1 ? singular : plural);
}
