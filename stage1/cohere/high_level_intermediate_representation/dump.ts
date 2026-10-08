import { panic } from 'adamic';
import type { HIRFunction, PlaceInterface, Instruction, TerminalType } from './core.ts';
export function placeText(place: PlaceInterface): string {
    return `${place.identifier}:${place.effect}:${place.reactive ? 1 : 0}:${place.start}:${place.end}`;
}
export function dump(fn: HIRFunction): string {
    let out = `hir-v1\nfunction ${fn.name} ${fn.kind} entry=${fn.entry} async=0 generator=0\n`;
    out += `params ${fn.params.map((place) => placeText(place)).join(',')}\ncontext ${fn.context.map((place) => placeText(place)).join(',')}\nreturns ${placeText(fn.returns)}\n`;
    for(const identifier of fn.identifiers) { out += `identifier ${identifier.id} ${identifier.declaration} ${identifier.name}\n`; }
    const seen = new Map<number, boolean>();
    for(const block of fn.blocks) {
        out += `block ${block.id} ${block.kind} predecessors=${block.predecessors.join(',')}\n`;
        for(const phi of block.phis) {
            out += `phi ${placeText(phi.place)}`;
            for(const operand of phi.operands) { out += ` ${operand.predecessor}=${placeText(operand.place)}`; }
            out += '\n';
        }
        for(const id of block.instructions) {
            const instruction = fn.instructions[id] ?? panic('missing instruction');
            seen.set(id, true);
            out += instructionText(instruction);
        }
        out += `terminal ${block.terminalOrder} ${block.terminal.kind} ${terminalText(block.terminal)}\n`;
    }
    for(const instruction of fn.instructions) { if(!seen.has(instruction.id)) { out += 'orphan ' + instructionText(instruction); } }
    out += 'scopes -\nend\n';
    return out;
}

// encoding/json's string alphabet, including its HTML and U+2028/U+2029 escapes.
export function quote(text: string): string {
    let out = '"';
    for(let at = 0; at < text.length; at++) {
        const code = text.charCodeAt(at);
        if(code === 34) { out += '\\"'; }
        else if(code === 92) { out += '\\\\'; }
        else if(code === 8) { out += '\\b'; }
        else if(code === 9) { out += '\\t'; }
        else if(code === 10) { out += '\\n'; }
        else if(code === 12) { out += '\\f'; }
        else if(code === 13) { out += '\\r'; }
        else if(code < 32 || code === 60 || code === 62 || code === 38 || code === 8232 || code === 8233) { out += `\\u${code.toString(16).padStart(4, '0')}`; }
        else { out += text.charAt(at); }
    }
    return out + '"';
}

function instructionText(instruction: Instruction): string {
    let payload = '';
    const value = instruction.value;
    if(value.kind === 'Primitive') { payload = value.literal; }
    else if(value.kind === 'LoadLocal') { payload = placeText(value.place); }
    else if(value.kind === 'UnaryExpression') { payload = `${value.operator} ${placeText(value.value)}`; }
    else if(value.kind === 'BinaryExpression') { payload = `${placeText(value.left)} ${value.operator} ${placeText(value.right)}`; }
    else if(value.kind === 'LoadGlobal') { payload = `{"BindingKind":${value.bindingKind},"Imported":${quote(value.imported)},"Name":${quote(value.name)},"Source":${quote(value.source)}}`; }
    else if(value.kind === 'StoreGlobal') { payload = `{"Name":${quote(value.name)},"Value":${quote(placeText(value.value))}}`; }
    else if(value.kind === 'DeclareLocal') { payload = `{"Kind":${value.declarationKind},"LValue":${quote(placeText(value.lvalue))}}`; }
    else if(value.kind === 'StoreLocal') { payload = `{"Kind":${value.declarationKind},"LValue":${quote(placeText(value.lvalue))},"Value":${quote(placeText(value.value))}}`; }
    else if(value.kind === 'PrefixUpdate' || value.kind === 'PostfixUpdate') { payload = `{"LValue":${quote(placeText(value.lvalue))},"Operation":${quote(value.operation)},"Value":${quote(placeText(value.value))}}`; }
    else if(value.kind === 'PropertyLoad') { payload = `{"Object":${quote(placeText(value.object))},"Optional":false,"Property":${quote(value.property)}}`; }
    else if(value.kind === 'ComputedLoad') { payload = `{"Object":${quote(placeText(value.object))},"Optional":false,"Property":${quote(placeText(value.property))}}`; }
    else if(value.kind === 'PropertyStore') { payload = `{"Object":${quote(placeText(value.object))},"Property":${quote(value.property)},"Value":${quote(placeText(value.value))}}`; }
    else if(value.kind === 'ComputedStore') { payload = `{"Object":${quote(placeText(value.object))},"Property":${quote(placeText(value.property))},"Value":${quote(placeText(value.value))}}`; }
    return `instruction ${instruction.id} ${instruction.order} ${placeText(instruction.lvalue)} ${instruction.start}:${instruction.end} ${instruction.value.kind} ${payload}\n`;
}

function terminalText(terminal: TerminalType): string {
    if(terminal.kind === 'Return') { return placeText(terminal.value); }
    if(terminal.kind === 'Throw') { return `{"Value":${quote(placeText(terminal.value))}}`; }
    if(terminal.kind === 'Goto') { return `{"Block":${terminal.block},"Variant":${terminal.variant}}`; }
    if(terminal.kind === 'If' || terminal.kind === 'Branch') { return `{"Alternate":${terminal.alternate},"Consequent":${terminal.consequent},"Fallthrough":${terminal.fallthrough},"Test":${quote(placeText(terminal.test))}}`; }
    if(terminal.kind === 'Logical') { return `{"Fallthrough":${terminal.fallthrough},"Operator":${quote(terminal.operator)},"Test":${terminal.test}}`; }
    if(terminal.kind === 'Ternary') { return `{"Fallthrough":${terminal.fallthrough},"Test":${terminal.test}}`; }
    if(terminal.kind === 'While') { return `{"Fallthrough":${terminal.fallthrough},"Loop":${terminal.loop},"Test":${terminal.test}}`; }
    return '{}';
}
