import { panic } from 'adamic';
import { allowedNames, allowedSegments, vocabulary } from './policy_data.ts';

export class Abbreviation {
    readonly form: string;
    readonly word: string;
    readonly id: string;
    readonly fields: string[];
    constructor(form: string, word: string, id: string, fields: string[]) {
        this.form = form;
        this.word = word;
        this.id = id;
        this.fields = fields;
    }
}
function upper(character: string): boolean {
    return character !== '' && character >= 'A' && character <= 'Z';
}
function lower(character: string): boolean {
    return character !== '' && character >= 'a' && character <= 'z';
}
function digit(character: string): boolean {
    return character !== '' && character >= '0' && character <= '9';
}
function capitalized(value: string): string {
    return value.slice(0, 1).toUpperCase() + value.slice(1);
}
function prefixFinding(text: string, entry: readonly string[]): Abbreviation {
    const word = entry[0] ?? panic('word');
    const expansion = entry[1] ?? '';
    const advisory = entry[5] === 'advice';
    return new Abbreviation('prefix', word, 'abbreviatedIdentifier', [
        'name',
        advisory ? word : text,
        'advice',
        advisory ? (entry[2] ?? '') : `Use "${expansion + text.slice(word.length)}".`,
    ]);
}
export function abbreviation(text: string): Abbreviation | undefined {
    if(allowedNames.includes(text)) {
        return undefined;
    }
    for(const entry of vocabulary) {
        const word = entry[0] ?? panic('word');
        const expansion = entry[1] ?? '';
        if(text === word && entry[3] !== '') {
            const style = entry[3];
            const advice =
                style === 'advice'
                    ? (entry[2] ?? '')
                    : `Use "${expansion}"${style === 'plain' ? '.' : ' or a more descriptive name.'}`;
            return new Abbreviation('whole', word, 'abbreviatedIdentifier', ['name', text, 'advice', advice]);
        }
    }
    for(const phase of ['early', 'suffix', 'late', 'segment']) {
        for(const entry of vocabulary) {
            const word = entry[0] ?? panic('word');
            const expansion = entry[1] ?? '';
            if(phase === 'early' || phase === 'late') {
                if(entry[4] !== phase || !text.startsWith(word) || !upper(text[word.length] ?? '')) {
                    continue;
                }
                let allowed = false;
                for(const segment of allowedSegments) {
                    if(text.includes(segment) && segment.toLowerCase().includes(word)) {
                        allowed = true;
                    }
                }
                if(!allowed) {
                    return prefixFinding(text, entry);
                }
            }
            if(phase === 'suffix' && entry[6] === 'true') {
                if(entry[7] === 'millisecondWord') {
                    for(let index = 1; index + 1 < text.length; index++) {
                        if(
                            text.slice(index, index + 2) === 'Ms' &&
                            lower(text[index - 1] ?? '') &&
                            (index + 2 === text.length || upper(text[index + 2] ?? ''))
                        ) {
                            return new Abbreviation('suffix', word, 'millisecondSuffix', [
                                'name',
                                text,
                                'suggestion',
                                text.slice(0, index) + (entry[8] ?? '') + text.slice(index + 2),
                            ]);
                        }
                    }
                }
                else {
                    const suffix = capitalized(word);
                    if(text.endsWith(suffix)) {
                        const replacement = entry[8] === '' ? capitalized(expansion) : (entry[8] ?? '');
                        const advice =
                            entry[9] === ''
                                ? `Use "${text.slice(0, text.length - suffix.length) + replacement}".`
                                : (entry[9] ?? '');
                        return new Abbreviation('suffix', word, 'abbreviatedSuffix', [
                            'name',
                            text,
                            'suffix',
                            suffix,
                            'advice',
                            advice,
                        ]);
                    }
                }
            }
            if(phase === 'segment' && entry[10] === 'true') {
                let allowed = false;
                for(const segment of allowedSegments) {
                    if(text.includes(segment)) {
                        allowed = true;
                    }
                }
                if(allowed) {
                    return undefined;
                }
                const segment = capitalized(word);
                let matches = false;
                let replace = -1;
                for(let index = 0; index + segment.length <= text.length; index++) {
                    if(text.slice(index, index + segment.length) !== segment) {
                        continue;
                    }
                    const before = text[index - 1] ?? '';
                    const after = text[index + segment.length] ?? '';
                    if(after !== '' && !upper(after) && !digit(after)) {
                        continue;
                    }
                    if(replace < 0) {
                        replace = index;
                    }
                    if(index === 0 || !upper(before)) {
                        matches = true;
                    }
                }
                if(matches && replace >= 0) {
                    return new Abbreviation('segment', word, 'abbreviatedWordSegment', [
                        'name',
                        text,
                        'word',
                        segment,
                        'suggestion',
                        text.slice(0, replace) + capitalized(expansion) + text.slice(replace + segment.length),
                    ]);
                }
            }
        }
    }
    return undefined;
}
