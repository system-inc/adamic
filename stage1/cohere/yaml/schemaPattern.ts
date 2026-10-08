// The fixed YAML tag tests need alternation, concatenation, ASCII classes and repetition.
// Matching returns possible end offsets so a later branch can backtrack without callbacks.
import { panic } from 'adamic';
import { PatternNode } from './patternNode.ts';
// The borrowed node arena is stable throughout recursive matching. Returning
// a node through an accessor would give each recursive frame another owner.
function matchPattern(nodes: readonly PatternNode[], index: number, source: string, position: number): number[] {
    const result: number[] = [];
    const node = nodes[index] ?? panic('missing schema pattern node');
    if(node.kind === 'class') {
        const character = source.slice(position, position + 1);
        if(character !== '' && node.characters.includes(character)) result.push(position + 1);
    }
    else if(node.kind === 'alternative') {
        for(const child of node.children)
            for(const end of matchPattern(nodes, child, source, position)) if(!result.includes(end)) result.push(end);
    }
    else if(node.kind === 'concat') {
        let positions: number[] = [position];
        for(const child of node.children) {
            const next: number[] = [];
            for(const at of positions)
                for(const end of matchPattern(nodes, child, source, at)) if(!next.includes(end)) next.push(end);
            positions = next;
            if(positions.length === 0) break;
        }
        for(const end of positions) result.push(end);
    }
    else {
        let positions: number[] = [position];
        if(node.minimum === 0) result.push(position);
        const maximum = node.maximum < 0 ? source.length + 1 : node.maximum;
        for(let count = 1; count <= maximum; count++) {
            const next: number[] = [];
            for(const at of positions)
                for(const end of matchPattern(nodes, node.children[0] ?? -1, source, at))
                    if(!next.includes(end)) next.push(end);
            if(next.length === 0) break;
            positions = next;
            if(count >= node.minimum) for(const end of positions) if(!result.includes(end)) result.push(end);
        }
    }
    return result;
}
export class SchemaPattern {
    pattern: string;
    offset = 0;
    nodes: PatternNode[] = [];
    root = -1;
    firstCharacters = '';
    firstMasks: number[] = [0, 0, 0, 0];
    constructor(pattern: string) {
        this.pattern = pattern.slice(1, -1);
        this.root = this.alternative();
        if(this.offset !== this.pattern.length) panic('unconsumed schema pattern');
        this.firstCharacters = this.leading(this.root);
        for(let index = 0; index < this.firstCharacters.length; index++) {
            const code = this.firstCharacters.charCodeAt(index);
            if(code < 128) {
                const word = Math.floor(code / 32);
                this.firstMasks[word] = (this.firstMasks[word] ?? 0) | (1 << (code % 32));
            }
        }
    }
    character(): string {
        return this.pattern.slice(this.offset, this.offset + 1);
    }
    make(kind: string): number {
        this.nodes.push(new PatternNode(kind));
        return this.nodes.length - 1;
    }
    node(index: number): PatternNode {
        return this.nodes[index] ?? panic('missing schema pattern node');
    }
    escaped(): string {
        const character = this.character();
        this.offset++;
        return character === 't' ? '\t' : character === 'n' ? '\n' : character === 'r' ? '\r' : character;
    }
    atom(): number {
        let result: number;
        const character = this.character();
        this.offset++;
        if(character === '(') {
            if(this.pattern.slice(this.offset, this.offset + 2) === '?:') this.offset += 2;
            result = this.alternative();
            if(this.character() !== ')') panic('unclosed schema pattern group');
            this.offset++;
        }
        else if(character === '[') {
            result = this.make('class');
            while(this.character() !== ']') {
                if(this.character() === '') panic('unclosed schema pattern class');
                let first = this.character();
                this.offset++;
                if(first === '\\') first = this.escaped();
                if(this.character() === '-' && this.pattern.slice(this.offset + 1, this.offset + 2) !== ']') {
                    this.offset++;
                    let last = this.character();
                    this.offset++;
                    if(last === '\\') last = this.escaped();
                    for(let code = first.charCodeAt(0); code <= last.charCodeAt(0); code++)
                        this.node(result).characters += String.fromCodePoint(code);
                }
                else this.node(result).characters += first;
            }
            this.offset++;
        }
        else {
            result = this.make('class');
            if(character === '\\') {
                if(this.character() === 'd') {
                    this.node(result).characters = '0123456789';
                    this.offset++;
                }
                else this.node(result).characters = this.escaped();
            }
            else this.node(result).characters = character;
        }
        const quantifier = this.character();
        if(quantifier === '?' || quantifier === '*' || quantifier === '+' || quantifier === '{') {
            this.offset++;
            const repeated = this.make('repeat');
            this.node(repeated).children.push(result);
            this.node(repeated).minimum = quantifier === '+' ? 1 : 0;
            this.node(repeated).maximum = quantifier === '?' ? 1 : -1;
            if(quantifier === '{') {
                const start = this.offset;
                while(this.character() !== '}') {
                    if(this.character() === '') panic('unclosed schema pattern repeat');
                    this.offset++;
                }
                const limits = this.pattern.slice(start, this.offset).split(',');
                this.node(repeated).minimum = Number(limits[0] ?? '0');
                this.node(repeated).maximum =
                    limits.length === 1
                        ? this.node(repeated).minimum
                        : (limits[1] ?? '') === ''
                          ? -1
                          : Number(limits[1] ?? '0');
                this.offset++;
            }
            result = repeated;
        }
        return result;
    }
    concatenate(): number {
        const index = this.make('concat');
        while(this.character() !== '' && this.character() !== ')' && this.character() !== '|')
            this.node(index).children.push(this.atom());
        return index;
    }
    alternative(): number {
        const index = this.make('alternative');
        this.node(index).children.push(this.concatenate());
        while(this.character() === '|') {
            this.offset++;
            this.node(index).children.push(this.concatenate());
        }
        return index;
    }
    match(index: number, source: string, position: number): number[] {
        return matchPattern(this.nodes, index, source, position);
    }
    // Only nullable concatenation prefixes contribute possible first units.
    leading(index: number): string {
        const node = this.node(index);
        if(node.kind === 'class') return node.characters;
        if(node.kind === 'repeat') return this.leading(node.children[0] ?? -1);
        const parts: string[] = [];
        for(const child of node.children) {
            parts.push(this.leading(child));
            if(node.kind === 'concat' && !this.match(child, '', 0).includes(0)) break;
        }
        return parts.join('');
    }
    test(source: string): boolean {
        if(source !== '') {
            const code = source.charCodeAt(0);
            if(code < 128) {
                if(((this.firstMasks[Math.floor(code / 32)] ?? 0) & (1 << (code % 32))) === 0) return false;
            }
            else if(!this.firstCharacters.includes(source.slice(0, 1))) return false;
        }
        return this.match(this.root, source, 0).includes(source.length);
    }
}
