// Cohere mdast/parse_markdown.go: the exact front-matter stage before micromark.
// Indices are UTF-16 in Adamic; the blanked prefix preserves those indices.
export interface FrontMatterInterface {
    readonly language: string;
    readonly explicitLanguage: string | undefined;
    readonly value: string;
    readonly startDelimiter: string;
    readonly endDelimiter: string;
    readonly raw: string;
}
export interface FrontMatterResultInterface {
    readonly frontMatter: FrontMatterInterface | undefined;
    readonly content: string;
}
function whitespace(code: number): boolean {
    return (
        code === 9 ||
        code === 10 ||
        code === 11 ||
        code === 12 ||
        code === 13 ||
        code === 32 ||
        code === 0xa0 ||
        code === 0x1680 ||
        code === 0x2028 ||
        code === 0x2029 ||
        code === 0x202f ||
        code === 0x205f ||
        code === 0x3000 ||
        code === 0xfeff ||
        (code >= 0x2000 && code <= 0x200a)
    );
}
function trimWhitespace(text: string): string {
    let start = 0;
    let end = text.length;
    while(start < end && whitespace(text.charCodeAt(start))) start++;
    while(end > start && whitespace(text.charCodeAt(end - 1))) end--;
    return text.slice(start, end);
}
function blankPrefix(text: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < text.length; index++) {
        if(text.charCodeAt(index) !== 10) continue;
        parts.push(' '.repeat(index - start));
        parts.push('\n');
        start = index + 1;
    }
    parts.push(' '.repeat(text.length - start));
    return parts.join('');
}
export function parseFrontMatter(text: string): FrontMatterResultInterface {
    const startDelimiter = text.slice(0, 3);
    if(startDelimiter !== '---' && startDelimiter !== '+++') return { frontMatter: undefined, content: text };
    const firstBreak = text.indexOf('\n', 3);
    if(firstBreak === -1) return { frontMatter: undefined, content: text };
    const explicit = trimWhitespace(text.slice(3, firstBreak));
    const language = explicit === '' ? (startDelimiter === '+++' ? 'toml' : 'yaml') : explicit;
    let endDelimiterIndex = text.indexOf(`\n${startDelimiter}`, firstBreak);
    if(endDelimiterIndex === -1 && startDelimiter === '---' && language === 'yaml') {
        endDelimiterIndex = text.indexOf('\n...', firstBreak);
    }
    if(endDelimiterIndex === -1) return { frontMatter: undefined, content: text };
    // Upstream's /\s?/ always matches: suffix text on the closing line is allowed.
    const raw = text.slice(0, endDelimiterIndex + 4);
    const value = firstBreak + 1 < endDelimiterIndex ? text.slice(firstBreak + 1, endDelimiterIndex) : '';
    const frontMatter: FrontMatterInterface = {
        language,
        explicitLanguage: explicit === '' ? undefined : explicit,
        value,
        startDelimiter,
        endDelimiter: raw.slice(raw.length - 3),
        raw,
    };
    return { frontMatter, content: blankPrefix(raw) + text.slice(raw.length) };
}
