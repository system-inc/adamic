// Keyword suggestions follow typescript-go's weighted spelling distance.
import { panic } from 'adamic';
import { keywords } from '../scanner/tokens.ts';

function characters(text: string): string[] {
    const result: string[] = [];
    for(const character of text) {
        result.push(character);
    }
    return result;
}

function distance(name: readonly string[], candidate: readonly string[], maximum: number): number {
    let previous: number[] = [];
    let current: number[] = [];
    for(let j = 0; j <= candidate.length; j++) {
        previous.push(j);
        current.push(0);
    }
    const big = maximum + 0.01;
    for(let i = 1; i <= name.length; i++) {
        const first = name[i - 1] ?? panic('missing spelling character');
        const minJ = Math.max(Math.ceil(i - maximum), 1);
        const maxJ = Math.min(Math.floor(maximum + i), candidate.length);
        let colMin = i;
        current[0] = colMin;
        for(let j = 1; j < minJ; j++) {
            current[j] = big;
        }
        for(let j = minJ; j <= maxJ; j++) {
            const second = candidate[j - 1] ?? panic('missing spelling candidate');
            const diagonal = previous[j - 1] ?? panic('missing spelling diagonal');
            const substitution =
                diagonal + (first.toLowerCase().codePointAt(0) === second.toLowerCase().codePointAt(0) ? 0.1 : 2);
            const value =
                first === second
                    ? diagonal
                    : Math.min(
                          (previous[j] ?? panic('missing spelling column')) + 1,
                          (current[j - 1] ?? panic('missing spelling row')) + 1,
                          substitution,
                      );
            current[j] = value;
            colMin = Math.min(colMin, value);
        }
        for(let j = maxJ + 1; j <= candidate.length; j++) {
            current[j] = big;
        }
        if(colMin > maximum) {
            return -1;
        }
        const saved = previous;
        previous = current;
        current = saved;
    }
    const result = previous[candidate.length] ?? panic('missing spelling result');
    return result > maximum ? -1 : result;
}

export function keywordSuggestion(text: string): string {
    const name = characters(text);
    const maximumDifference = Math.max(2, Math.floor(name.length * 0.34));
    let bestDistance = Math.floor(name.length * 0.4) + 0.9;
    let best = '';
    for(const candidate of keywords.keys()) {
        if(
            candidate.length <= 2 ||
            candidate === text ||
            Math.abs(candidate.length - name.length) > maximumDifference
        ) {
            continue;
        }
        const value = distance(name, characters(candidate), bestDistance);
        if(value >= 0 && (value < bestDistance || best === '' || candidate < best)) {
            bestDistance = value;
            best = candidate;
        }
    }
    if(best !== '') {
        return best;
    }
    for(const candidate of keywords.keys()) {
        if(candidate.length > 2 && text.length > candidate.length + 2 && text.startsWith(candidate)) {
            return candidate + ' ' + text.slice(candidate.length);
        }
    }
    return '';
}
