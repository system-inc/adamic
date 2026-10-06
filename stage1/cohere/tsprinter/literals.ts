// Literal paths through cohere utility_text.go and print_literal.go.
export function numberText(raw: string): string {
    if(raw.length === 1) {
        return raw;
    }
    let value = raw.toLowerCase();
    const exponent = value.indexOf('e');
    let decimalPrefix = exponent > 0;
    for(let index = 0; index < exponent; index++) {
        const char = value.slice(index, index + 1);
        if((char < '0' || char > '9') && char !== '.' && !(index === 0 && (char === '+' || char === '-'))) {
            decimalPrefix = false;
        }
    }
    if(decimalPrefix) {
        let tail = value.slice(exponent + 1);
        const negative = tail.startsWith('-');
        if(tail.startsWith('-') || tail.startsWith('+')) {
            tail = tail.slice(1);
        }
        if(tail.charCodeAt(0) >= 48 && tail.charCodeAt(0) <= 57) {
            while(tail.length > 1 && tail.startsWith('0') && tail.charCodeAt(1) >= 48 && tail.charCodeAt(1) <= 57) {
                tail = tail.slice(1);
            }
            value = value.slice(0, exponent) + (tail === '0' ? '' : `e${negative ? '-' : ''}${tail}`);
        }
    }
    if(value.startsWith('.')) {
        value = `0${value}`;
    }
    const exp = value.indexOf('e');
    const split = exp < 0 || value.startsWith('0x') ? value.length : exp;
    let mantissa = value.slice(0, split);
    const dot = mantissa.indexOf('.');
    if(dot >= 0 && !mantissa.slice(dot + 1).includes('_')) {
        while(mantissa.endsWith('0') && mantissa.length > dot + 2) {
            mantissa = mantissa.slice(0, -1);
        }
    }
    if(mantissa.endsWith('.')) {
        mantissa = mantissa.slice(0, -1);
    }
    return mantissa + value.slice(split);
}

export function stringText(raw: string, directive: boolean, preferSingle = true): string {
    const content = raw.slice(1, -1);
    if(directive)
        return content === 'use strict' || (!content.includes("'") && !content.includes('"')) ? `'${content}'` : raw;
    const single = content.split("'").length - 1;
    const double = content.split('"').length - 1;
    const quote = single > double || (single === double && !preferSingle) ? '"' : "'";
    if(raw.slice(0, 1) === quote) return raw;
    const parts: string[] = [quote];
    for(let index = 0; index < content.length; index++) {
        const char = content.slice(index, index + 1);
        if(char === '\\') {
            const next = content.slice(index + 1, index + 2);
            if(next === '"' || next === "'" || next === '\\') {
                if(next === quote || next === '\\') parts.push('\\');
                parts.push(next);
                index++;
            }
            else parts.push(char);
        }
        else {
            if(char === quote) parts.push('\\');
            parts.push(char);
        }
    }
    parts.push(quote);
    return parts.join('');
}
