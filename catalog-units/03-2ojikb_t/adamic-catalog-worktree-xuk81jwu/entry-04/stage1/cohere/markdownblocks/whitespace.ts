// Cohere whitespace.go and the whitespace dispatch's syntax-leading override.
import { panic } from 'adamic';
import type { DocumentArena } from './document.ts';
export interface WhitespaceTokenInterface {
    readonly present: boolean;
    readonly type: string;
    readonly value: string;
    readonly kind: string;
    readonly cj: boolean;
    readonly leading: boolean;
    readonly trailing: boolean;
}
export interface CJSpaceSampleInterface {
    readonly type: string;
    readonly value: number;
    readonly previousKind: string;
    readonly nextKind: string;
}
export interface WhitespaceFrameInterface {
    readonly value: string;
    readonly proseWrap: string;
    readonly link: boolean;
    readonly previous: WhitespaceTokenInterface;
    readonly next: WhitespaceTokenInterface;
    readonly afterNext: WhitespaceTokenInterface;
    readonly ancestorKinds: readonly string[];
    readonly ancestorSetext: readonly boolean[];
    readonly samples: readonly CJSpaceSampleInterface[];
}
function koreanPair(previous: string, next: string): boolean {
    return (previous === 'k-letter' && next === 'cj-letter') || (previous === 'cj-letter' && next === 'k-letter');
}
function nonCJ(kind: string): boolean {
    return kind === 'non-cjk' || kind === 'k-letter';
}
function asciiPunctuation(unit: number): boolean {
    return (
        (unit >= 33 && unit <= 47) ||
        (unit >= 58 && unit <= 64) ||
        (unit >= 91 && unit <= 96) ||
        (unit >= 123 && unit <= 126)
    );
}
function cjSpaces(samples: readonly CJSpaceSampleInterface[]): boolean {
    let spaces = 0;
    let empties = 0;
    for(const sample of samples) {
        if(sample.type !== 'whitespace' || (sample.value !== 0 && sample.value !== 1)) continue;
        if(
            (sample.previousKind === 'cj-letter' && sample.nextKind === 'non-cjk') ||
            (sample.previousKind === 'non-cjk' && sample.nextKind === 'cj-letter')
        ) {
            if(sample.value === 1) spaces++;
            else empties++;
        }
    }
    return spaces > empties;
}
function converts(frame: WhitespaceFrameInterface): boolean {
    if(frame.link || !frame.previous.present || !frame.next.present) return true;
    if((nonCJ(frame.previous.kind) && nonCJ(frame.next.kind)) || koreanPair(frame.previous.kind, frame.next.kind))
        return true;
    if(
        frame.previous.kind === 'cjk-punctuation' ||
        frame.next.kind === 'cjk-punctuation' ||
        (frame.previous.kind === 'cj-letter' && frame.next.kind === 'cj-letter')
    )
        return false;
    const first = frame.next.value.length === 0 ? -1 : frame.next.value.charCodeAt(0);
    const last =
        frame.previous.value.length === 0 ? -1 : frame.previous.value.charCodeAt(frame.previous.value.length - 1);
    if(asciiPunctuation(first) || asciiPunctuation(last)) return true;
    if(frame.previous.trailing || frame.next.leading) return false;
    return cjSpaces(frame.samples);
}
function breakable(frame: WhitespaceFrameInterface, proseWrap: string): boolean {
    if(proseWrap !== 'always') return false;
    for(let index = 0; index < frame.ancestorKinds.length; index++) {
        const kind = frame.ancestorKinds[index] ?? panic('whitespace ancestor');
        if(
            kind === 'tableCell' ||
            kind === 'link' ||
            kind === 'wikiLink' ||
            (kind === 'heading' && !(frame.ancestorSetext[index] ?? false))
        )
            return false;
    }
    if(frame.link) return frame.value !== '';
    if(!frame.previous.present || !frame.next.present) return true;
    if(frame.value === '') return false;
    if(koreanPair(frame.previous.kind, frame.next.kind)) return true;
    return !frame.previous.cj && !frame.next.cj;
}
function syntaxLeading(value: string): boolean {
    if(value.startsWith('>')) return true;
    if(value === '*' || value === '+' || value === '-') return true;
    if(value.length >= 1 && value.length <= 6) {
        let hashes = true;
        for(let index = 0; index < value.length; index++) if(value.charCodeAt(index) !== 35) hashes = false;
        if(hashes) return true;
    }
    if(value.length < 2 || (!value.endsWith('.') && !value.endsWith(')'))) return false;
    for(let index = 0; index < value.length - 1; index++) {
        const code = value.charCodeAt(index);
        if(code < 48 || code > 57) return false;
    }
    return true;
}
export function printWhitespace(arena: DocumentArena, frame: WhitespaceFrameInterface): number {
    let proseWrap = frame.proseWrap;
    const fakeAfterNext =
        frame.afterNext.present && frame.afterNext.type === 'whitespace' && frame.afterNext.value === '';
    const fakeSetext =
        frame.value === '\n' &&
        frame.next.value === '-' &&
        (!frame.afterNext.present || (frame.afterNext.type === 'whitespace' && frame.afterNext.value === '\n'));
    if(
        !frame.link &&
        frame.next.present &&
        syntaxLeading(frame.next.value) &&
        !fakeAfterNext &&
        !(proseWrap === 'preserve' && fakeSetext)
    )
        proseWrap = 'never';
    if(proseWrap === 'preserve' && frame.value === '\n') return arena.hardline();
    const canBeSpace = frame.value === ' ' || (frame.value === '\n' && converts(frame));
    if(breakable(frame, proseWrap)) return arena.add('h', '', 0, [], canBeSpace ? 0 : 1);
    return arena.text(canBeSpace ? ' ' : '');
}
