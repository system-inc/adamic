// Cohere ParseFrontMatter, including the fork's permissive closing-delimiter check.
export class ParsedFrontMatter {
    readonly present: boolean;
    readonly language: string;
    readonly explicitLanguage: string;
    readonly value: string;
    readonly startDelimiter: string;
    readonly endDelimiter: string;
    readonly raw: string;
    readonly content: string;
    constructor(
        present: boolean,
        language: string,
        explicitLanguage: string,
        value: string,
        startDelimiter: string,
        endDelimiter: string,
        raw: string,
        content: string,
    ) {
        this.present = present;
        this.language = language;
        this.explicitLanguage = explicitLanguage;
        this.value = value;
        this.startDelimiter = startDelimiter;
        this.endDelimiter = endDelimiter;
        this.raw = raw;
        this.content = content;
    }
}
function indexFrom(text: string, search: string, from: number): number {
    const index = text.slice(from).indexOf(search);
    return index < 0 ? -1 : index + from;
}
export function parseFrontMatter(text: string): ParsedFrontMatter {
    const delimiter = text.slice(0, 3);
    const empty = new ParsedFrontMatter(false, '', '', '', '', '', '', text);
    if(delimiter !== '---' && delimiter !== '+++') return empty;
    const newline = indexFrom(text, '\n', 3);
    if(newline < 0) return empty;
    const explicit = text.slice(3, newline).trim();
    const language = explicit === '' ? (delimiter === '+++' ? 'toml' : 'yaml') : explicit;
    let end = indexFrom(text, `\n${delimiter}`, newline);
    if(end < 0 && delimiter === '---' && language === 'yaml') end = indexFrom(text, '\n...', newline);
    if(end < 0) return empty;
    const raw = text.slice(0, end + 4);
    const value = text.slice(newline + 1, end);
    const blank: string[] = [];
    for(let index = 0; index < raw.length; index++) blank.push(raw.charCodeAt(index) === 10 ? '\n' : ' ');
    return new ParsedFrontMatter(
        true,
        language,
        explicit,
        value,
        delimiter,
        raw.slice(-3),
        raw,
        blank.join('') + text.slice(end + 4),
    );
}
