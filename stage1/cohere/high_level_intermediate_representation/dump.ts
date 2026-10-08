import { panic } from 'adamic';
import type { HIRFunction, PlaceInterface } from './core.ts';
export function placeText(place: PlaceInterface): string {
    return `${place.identifier}:${place.effect}:${place.reactive ? 1 : 0}:${place.start}:${place.end}`;
}
export function dump(fn: HIRFunction): string {
    let out = `hir-v1\nfunction ${fn.name} ${fn.kind} entry=${fn.entry} async=0 generator=0\n`;
    out += `params ${fn.params.map((place) => placeText(place)).join(',')}\ncontext ${fn.context.map((place) => placeText(place)).join(',')}\nreturns ${placeText(fn.returns)}\n`;
    for(const identifier of fn.identifiers) { out += `identifier ${identifier.id} ${identifier.declaration} ${identifier.name}\n`; }
    for(const block of fn.blocks) {
        out += `block ${block.id} block predecessors=${block.predecessors.join(',')}\n`;
        for(const id of block.instructions) {
            const instruction = fn.instructions[id] ?? panic('missing instruction');
            let payload = '';
            const value = instruction.value;
            if(value.kind === 'Primitive') { payload = value.literal; }
            else if(value.kind === 'LoadLocal') { payload = placeText(value.place); }
            else if(value.kind === 'UnaryExpression') { payload = `${value.operator} ${placeText(value.value)}`; }
            else { payload = `${placeText(value.left)} ${value.operator} ${placeText(value.right)}`; }
            out += `instruction ${instruction.id} ${instruction.order} ${placeText(instruction.lvalue)} ${instruction.start}:${instruction.end} ${instruction.value.kind} ${payload}\n`;
        }
        out += `terminal ${block.terminalOrder} Return ${placeText(block.terminal)}\n`;
    }
    out += 'scopes -\nend\n';
    return out;
}
