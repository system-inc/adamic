// Adapter only: all graph algorithms are imported from the existing SSA port.
import { panic } from 'adamic';
import { construct } from '../static_single_assignment/construct.ts';
import { reversePostorder, markPredecessors, markEvaluationOrder } from '../static_single_assignment/graph.ts';
import type { GraphInterface } from '../static_single_assignment/static_single_assignment.ts';
import { HIRFunction, BasicBlock } from './core.ts';
import type { PlaceInterface } from './core.ts';
export const hirGraph: GraphInterface<HIRFunction, BasicBlock, PlaceInterface> = {
    entry: (fn) => fn.entry,
    blockBound: (_fn) => 2,
    block: (fn, id) => fn.blocks.find((block) => block.id === id),
    blocks: (fn) => fn.blocks,
    setBlocks: function(fn, blocks) { fn.blocks = blocks; },
    retain: function(fn, keep) { fn.blocks = fn.blocks.filter((block) => keep(block.id)); },
    placeholder: (_fn, block) => block,
    id: (block) => block.id,
    predecessors: (block) => block.predecessors,
    setPredecessors: function(block, predecessors) { block.predecessors = predecessors; },
    phis: (block) => block.phis,
    setPhis: function(block, phis) { block.phis = phis; },
    eachEdge: function(_block, _visit) {},
    endsInReturn: (_block) => true,
    instructionCount: (_fn, block) => block.instructions.length,
    eachInstructionPlace: function(fn, block, index, visit) {
        const instruction = fn.instructions[block.instructions[index] ?? panic('missing instruction id')] ?? panic('missing instruction');
        instruction.lvalue = visit(instruction.lvalue, 'Define');
        if(instruction.value.kind === 'LoadLocal') { instruction.value.place = visit(instruction.value.place, 'Use'); }
        if(instruction.value.kind === 'UnaryExpression') { instruction.value.value = visit(instruction.value.value, 'Use'); }
        if(instruction.value.kind === 'BinaryExpression') {
            instruction.value.left = visit(instruction.value.left, 'Use');
            instruction.value.right = visit(instruction.value.right, 'Use');
        }
        const value = instruction.value;
        if(value.kind === 'StoreGlobal' || value.kind === 'StoreLocal' || value.kind === 'PrefixUpdate' || value.kind === 'PostfixUpdate' || value.kind === 'PropertyStore' || value.kind === 'ComputedStore') { value.value = visit(value.value, 'Use'); }
        if(value.kind === 'PropertyLoad' || value.kind === 'ComputedLoad' || value.kind === 'PropertyStore' || value.kind === 'ComputedStore') { value.object = visit(value.object, 'Use'); }
        if(value.kind === 'ComputedLoad' || value.kind === 'ComputedStore') { value.property = visit(value.property, 'Use'); }
        if(value.kind === 'DeclareLocal' || value.kind === 'StoreLocal' || value.kind === 'PrefixUpdate' || value.kind === 'PostfixUpdate') { value.lvalue = visit(value.lvalue, 'Define'); }

    },
    isContextStore: (_fn, _block, _index) => false,
    contextStoreDefines: (_fn, _block, _index, _place) => false,
    setInstructionOrder: function(fn, block, index, order) {
        const instruction = fn.instructions[block.instructions[index] ?? panic('missing instruction id')] ?? panic('missing instruction');
        instruction.order = order;
    },
    eachTerminalPlace: function(block, visit) { block.terminal = visit(block.terminal, 'Use'); },
    setTerminalOrder: function(block, order) { block.terminalOrder = order; },
    params: (fn) => fn.params,
    returns: (fn) => fn.returns,
    setReturns: function(fn, place) { fn.returns = place; },
    declaration: (fn, id) => (fn.identifiers[id] ?? panic('missing identifier')).declaration,
    contextual: (_fn, _declaration) => false,
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
export function constructHIR(fn: HIRFunction): void {
    reversePostorder(hirGraph, fn);
    markPredecessors(hirGraph, fn);
    markEvaluationOrder(hirGraph, fn);
    construct(hirGraph, fn);
}
