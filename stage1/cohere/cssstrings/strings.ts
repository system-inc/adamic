// CSS adjustStrings, printString, getPreferredQuote and makeString from Go cohere.
// A finite scanner implements the two upstream patterns without a regex runtime.
export function printString(raw: string, singleQuote: boolean): string {
    const content = raw.slice(1, -1);
    const preferred = singleQuote ? "'" : '"';
    const alternate = singleQuote ? '"' : "'";
    let preferredCount = 0;
    let alternateCount = 0;
    for(let index = 0; index < content.length; index++) {
        const character = content.slice(index, index + 1);
        if(character === preferred) preferredCount++;
        if(character === alternate) alternateCount++;
    }
    const quote = preferredCount > alternateCount ? alternate : preferred;
    if(raw.slice(0, 1) === quote) return raw;
    const other = quote === '"' ? "'" : '"';
    const parts: string[] = [quote];
    let start = 0;
    for(let index = 0; index < content.length; index++) {
        const character = content.slice(index, index + 1);
        const next = content.slice(index + 1, index + 2);
        if(character === '\\' && (next === '"' || next === "'" || next === '\\')) {
            parts.push(content.slice(start, index), next === other ? next : character + next);
            index++;
            start = index + 1;
        }
        else if(character === quote) {
            parts.push(content.slice(start, index), `\\${character}`);
            start = index + 1;
        }
    }
    parts.push(content.slice(start), quote);
    return parts.join('');
}

function stringEnd(value: string, start: number): number {
    const quote = value.charCodeAt(start);
    if(quote !== 34 && quote !== 39) return -1;
    for(let index = start + 1; index < value.length; index++) {
        const character = value.charCodeAt(index);
        if(character === quote) return index + 1;
        if(character === 92) index++;
    }
    return -1;
}

export function adjustStrings(value: string, singleQuote: boolean): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < value.length; index++) {
        const end = stringEnd(value, index);
        if(end < 0) continue;
        parts.push(value.slice(start, index), printString(value.slice(index, end), singleQuote));
        index = end - 1;
        start = end;
    }
    parts.push(value.slice(start));
    return parts.join('');
}
