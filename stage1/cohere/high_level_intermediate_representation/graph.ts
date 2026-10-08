// Adapter only: all graph algorithms are imported from the existing SSA port.
import { panic } from 'adamic';
import { construct } from '../static_single_assignment/construct.ts';
import { reversePostorder, markPredecessors, markEvaluationOrder } from '../static_single_assignment/graph.ts';
import type { GraphInterface } from '../static_single_assignment/static_single_assignment.ts';
import { HIRFunction, BasicBlock, HIRArena, blockIndex, testPlace, testBlock } from './core.ts';
import type { PlaceInterface, FunctionIndex } from './core.ts';
export const hirGraph: GraphInterface<HIRFunction, BasicBlock, PlaceInterface> = {
    entry: (fn) => fn.entry,
    blockBound: (fn) => fn.nextBlock,
    block: (fn, id) => fn.retained.has(id) ? fn.block(blockIndex(id)) : undefined,
    blocks: (fn) => fn.blockOrder.map((id) => fn.block(id)),
    setBlocks: function(fn, blocks) { fn.blockOrder = blocks.map((block) => block.id); },
    retain: function(fn, keep) { for(const id of fn.retained) { if(!keep(id)) { fn.retained.delete(id); } } },
    placeholder: function(fn, block) {
        const placeholder = new BasicBlock(block.id, { kind: 'Unreachable' }, block.kind);
        placeholder.predecessors = [...block.predecessors];
        fn.blockTable[block.id - 1] = placeholder;
        return placeholder;
    },
    id: (block) => block.id,
    predecessors: (block) => block.predecessors,
    setPredecessors: function(block, predecessors) { block.predecessors = predecessors.map((id) => blockIndex(id)); },
    phis: (block) => block.phis,
    setPhis: function(block, phis) { block.phis = phis; },
    eachEdge: function(block, visit) {
        const terminal = block.terminal;
        if(terminal.kind === 'If' || terminal.kind === 'Branch' || terminal.kind === 'Logical' || terminal.kind === 'Ternary' || terminal.kind === 'While') { visit(terminal.fallthrough ?? panic('missing fallthrough'), 'Fallthrough'); }
        if(terminal.kind === 'Goto') { visit(terminal.block ?? panic('missing goto target'), 'Real'); }
        else if(terminal.kind === 'If' || terminal.kind === 'Branch') { visit(terminal.consequent ?? panic('missing consequent'), 'Real'); visit(terminal.alternate ?? panic('missing alternate'), 'Real'); }
        else if(terminal.kind === 'Logical' || terminal.kind === 'Ternary' || terminal.kind === 'While') { visit(testBlock(terminal), 'Real'); }
    },
    endsInReturn: (block) => block.terminal.kind === 'Return',
    instructionCount: (_fn, block) => block.instructions.length,
    eachInstructionPlace: function(fn, block, index, visit) {
        const instruction = fn.instructions[block.instructions[index] ?? panic('missing instruction id')] ?? panic('missing instruction');
        instruction.lvalue = visit(instruction.lvalue, 'Define');
        if((instruction.value.kind === 'LoadLocal' || instruction.value.kind === 'LoadContext')) { instruction.value.place = visit(instruction.value.place, 'Use'); }
        if(instruction.value.kind === 'UnaryExpression') { instruction.value.value = visit(instruction.value.value, 'Use'); }
        if(instruction.value.kind === 'BinaryExpression') {
            instruction.value.left = visit(instruction.value.left, 'Use');
            instruction.value.right = visit(instruction.value.right, 'Use');
        }
        const value = instruction.value;
        if(value.kind === 'StoreGlobal' || value.kind === 'StoreLocal' || value.kind === 'StoreContext' || value.kind === 'PrefixUpdate' || value.kind === 'PostfixUpdate' || value.kind === 'PropertyStore' || value.kind === 'ComputedStore') { value.value = visit(value.value, 'Use'); }
        if(value.kind === 'PropertyLoad' || value.kind === 'ComputedLoad' || value.kind === 'PropertyStore' || value.kind === 'ComputedStore') { value.object = visit(value.object, 'Use'); }
        if(value.kind === 'ComputedLoad' || value.kind === 'ComputedStore') { value.property = visit(value.property, 'Use'); }
        if(value.kind === 'CallExpression' || value.kind === 'NewExpression') { value.callee = visit(value.callee, 'Use'); }
        if(value.kind === 'MethodCall') { value.receiver = visit(value.receiver, 'Use'); value.property = visit(value.property, 'Use'); }
        if(value.kind === 'CallExpression' || value.kind === 'NewExpression' || value.kind === 'MethodCall') { for(const argument of value.args) { argument.place = visit(argument.place, 'Use'); } }
        if(value.kind === 'FunctionExpression') { for(let index = 0; index < value.captures.length; index++) { value.captures[index] = visit(value.captures[index] ?? panic('missing capture'), 'Use'); } }
        if(value.kind === 'DeclareLocal' || value.kind === 'StoreLocal' || value.kind === 'StoreContext' || value.kind === 'PrefixUpdate' || value.kind === 'PostfixUpdate') { value.lvalue = visit(value.lvalue, 'Define'); }

    },
    isContextStore: (fn, block, index) => (fn.instructions[block.instructions[index] ?? -1] ?? panic('missing store')).value.kind === 'StoreContext',
    contextStoreDefines: (fn, block, index, place) => place.identifier === (fn.instructions[block.instructions[index] ?? -1] ?? panic('missing store')).lvalue.identifier,
    setInstructionOrder: function(fn, block, index, order) {
        const instruction = fn.instructions[block.instructions[index] ?? panic('missing instruction id')] ?? panic('missing instruction');
        instruction.order = order;
    },
    eachTerminalPlace: function(block, visit) {
        const terminal = block.terminal;
        if(terminal.kind === 'Return' || terminal.kind === 'Throw') { terminal.value = visit(terminal.value ?? panic('missing terminal value'), 'Use'); }
        else if(terminal.kind === 'If' || terminal.kind === 'Branch') { terminal.testPlace = visit(testPlace(terminal), 'Use'); }
    },
    setTerminalOrder: function(block, order) { block.terminalOrder = order; },
    params: (fn) => fn.params,
    returns: (fn) => fn.returns,
    setReturns: function(fn, place) { fn.returns = place; },
    declaration: (fn, id) => (fn.identifiers[id] ?? panic('missing identifier')).declaration,
    contextual: (fn, declaration) => fn.contextDeclarations.has(declaration),
    mint: function(fn, original) {
        const old = fn.identifiers[original] ?? panic('missing identifier');
        const id = fn.identifiers.length;
        fn.identifiers.push({ id, declaration: old.declaration, name: old.name });
        return id;
    },
    named: (fn, id) => (fn.identifiers[id] ?? panic('missing identifier')).name !== '',
    placeString: (_fn, id) => `$${id}`,
    identifierOf: (place) => place.identifier,
    withIdentifier: (place, identifier) => ({ identifier, effect: place.effect, reactive: place.reactive, start: place.start, end: place.end }),
};
export function constructHIR(arena: HIRArena, index: FunctionIndex): void {
    const fn = arena.read(index);
    reversePostorder(hirGraph, fn);
    markPredecessors(hirGraph, fn);
    markEvaluationOrder(hirGraph, fn);
    construct(hirGraph, fn);
    for(const nested of fn.functions) { constructHIR(arena, nested); }
}
