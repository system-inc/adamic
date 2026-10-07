// Ordered Thompson threads retain Go regexp's first-alternative and greedy priority.
import { panic } from 'adamic';
import type { WidthInstructionInterface } from './widthTables.ts';
export function inRanges(code: number, ranges: readonly number[]): boolean {
    let low = 0;
    let high = ranges.length / 2;
    while(low < high) {
        const middle = Math.floor((low + high) / 2);
        if(code < (ranges[middle * 2] ?? panic('range start'))) high = middle;
        else if(code > (ranges[middle * 2 + 1] ?? panic('range end'))) low = middle + 1;
        else return true;
    }
    return false;
}
export function mappedUnit(code: number): number {
    return code >= 0xd800 && code <= 0xdfff ? 0x100000 + code - 0xd800 : code;
}
export class EmojiMatcher {
    readonly seen: number[];
    generation = 0;
    readonly instructions: readonly WidthInstructionInterface[];
    readonly entry: number;
    constructor(instructions: readonly WidthInstructionInterface[], entry: number) {
        this.instructions = instructions;
        this.entry = entry;
        this.seen = Array.from({ length: instructions.length }, () => 0);
    }
    closure(seeds: readonly number[], position: number, start: number, end: number): number[] {
        this.generation++;
        const stack = seeds.slice().reverse();
        const active: number[] = [];
        while(stack.length > 0) {
            const id = stack.pop() ?? panic('matcher stack');
            if(this.seen[id] === this.generation) continue;
            this.seen[id] = this.generation;
            const instruction = this.instructions[id] ?? panic('matcher instruction');
            if(instruction.kind === 'a') {
                stack.push(instruction.alternate);
                stack.push(instruction.out);
            }
            else if(instruction.kind === 'n') stack.push(instruction.out);
            else if(instruction.kind === 'e') {
                if(
                    ((instruction.assertion & 4) === 0 || position === start) &&
                    ((instruction.assertion & 8) === 0 || position === end)
                )
                    stack.push(instruction.out);
            }
            else if(instruction.kind === 'r' || instruction.kind === 'm') active.push(id);
        }
        return active;
    }
    match(text: string, start: number, end: number): number {
        let position = start;
        let active = this.closure([this.entry], position, start, end);
        let best = -1;
        while(active.length > 0) {
            const seeds: number[] = [];
            for(const id of active) {
                const instruction = this.instructions[id] ?? panic('active instruction');
                if(instruction.kind === 'm') {
                    best = position;
                    break;
                }
                if(position < end && inRanges(mappedUnit(text.charCodeAt(position)), instruction.ranges))
                    seeds.push(instruction.out);
            }
            if(seeds.length === 0) break;
            position++;
            active = this.closure(seeds, position, start, end);
        }
        return best;
    }
}
