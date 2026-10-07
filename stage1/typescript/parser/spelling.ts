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
    for(let column = 0; column <= candidate.length; column++) {
        previous.push(column);
        current.push(0);
    }
    const big = maximum + 0.01;
    for(let row = 1; row <= name.length; row++) {
        const first = name[row - 1] ?? panic('missing spelling character');
        const minJ = Math.max(Math.ceil(row - maximum), 1);
        const maxJ = Math.min(Math.floor(maximum + row), candidate.length);
        let colMin = row;
        current[0] = colMin;
        for(let column = 1; column < minJ; column++) {
            current[column] = big;
        }
        for(let column = minJ; column <= maxJ; column++) {
            const second = candidate[column - 1] ?? panic('missing spelling candidate');
            const diagonal = previous[column - 1] ?? panic('missing spelling diagonal');
            const substitution =
                diagonal + (first.toLowerCase().codePointAt(0) === second.toLowerCase().codePointAt(0) ? 0.1 : 2);
            const value =
                first === second
                    ? diagonal
                    : Math.min(
                          (previous[column] ?? panic('missing spelling column')) + 1,
                          (current[column - 1] ?? panic('missing spelling row')) + 1,
                          substitution,
                      );
            current[column] = value;
            colMin = Math.min(colMin, value);
        }
        for(let column = maxJ + 1; column <= candidate.length; column++) {
            current[column] = big;
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
            return `${candidate} ${text.slice(candidate.length)}`;
        }
    }
    return '';
}
