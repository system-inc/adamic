/*
 * construct.go: single static assignment form. Rename every definition so each value is written
 * exactly once, and insert phi nodes where control from several predecessors rejoins.
 *
 * This is Braun et al., "Simple and Efficient Construction of Static Single Assignment Form" (CC 2013),
 * in the sealed-block form upstream uses. React's own specification of the pass is at
 * `compiler/packages/babel-plugin-react-compiler/docs/passes/02-enterSSA.md`, and the shape here follows
 * it closely enough that the two can be read side by side.
 *
 * Why Braun rather than Cytron: the textbook construction computes the iterated dominance frontier,
 * places empty phis at every block in it, then renames in a second walk over the dominator tree. Braun
 * places a phi only when a lookup actually crosses a join with disagreeing predecessors, which means it
 * never places one that redundancy elimination would immediately remove, and it needs no dominance
 * frontier at all. Dominance is still what correctness is stated against (see verifySSA), but it is
 * computed there, for checking rather than for construction.
 *
 * The one property everything rests on: a block may be processed only after every predecessor that is
 * not a back edge, so that a lookup into a predecessor finds a finished answer. The block array is in
 * reverse postorder, which is exactly that guarantee, and an IR's finalize establishes it. A caller that
 * restructured the graph without re-running it gets silent nonsense, which is why an IR's construct
 * re-runs it first.
 *
 * Loops, which is where a construction that looks right on straight-line code breaks: a loop header is
 * reached from before the loop and from the back edge, and the back edge's block has not been processed
 * when the header is. Braun's answer is the incomplete phi: when a lookup reaches a block with
 * unprocessed predecessors, mint the phi's result immediately, record it as the block's definition so
 * the loop body's reads bind to it, and leave the operands to be filled when the last predecessor
 * lands. Recording the definition before recursing is also what stops the lookup recursing forever
 * around the cycle.
 *
 * What is deliberately not here: redundant-phi elimination is a separate pass, eliminateRedundantPhis,
 * in eliminate.ts, which construct runs. Recursion into nested functions is each IR's own, around this.
 */

import { panic } from 'adamic';
import { eliminateRedundantPhis } from './eliminate.ts';
import { edgesOf } from './graph.ts';
import { withPhiOperand } from './phi.ts';
import type {
    BlockIdType,
    DeclarationIdType,
    GraphInterface,
    IdentifierIdType,
    PhiOperandInterface,
    RoleType,
} from './static_single_assignment.ts';

// SingleAssignmentState is one block's view: what each original binding currently resolves to, and the
// phis whose operands are still waiting on an unprocessed predecessor.
interface SingleAssignmentStateInterface<P> {
    /*
     * definitions maps a binding to the value it holds on entry to, or within, this block.
     *
     * The key is the DeclarationId rather than the IdentifierId, and that is the single most important
     * decision in this file. Lowering already mints a fresh IdentifierId for every store to a variable,
     * so `let y` reassigned twice arrives as three distinct identifiers sharing one declaration. Keying
     * on the identifier would make each store a different variable, and a lookup would never find a
     * predecessor's definition because the predecessor stored under a different key.
     */
    readonly definitions: Map<DeclarationIdType, IdentifierIdType>;

    // incompletePhis are phis minted before every predecessor was processed. Their result is already
    // bound in definitions; only the operands are outstanding.
    incompletePhis: IncompletePhiInterface<P>[];
}

// newSingleAssignmentState is a block's view before anything is defined in it.
function newSingleAssignmentState<P>(): SingleAssignmentStateInterface<P> {
    return { definitions: new Map<DeclarationIdType, IdentifierIdType>(), incompletePhis: [] };
}

// IncompletePhi is a phi awaiting operands: original is the pre-assignment place being merged, the key
// a lookup uses, and renamed the phi's result, already minted and already visible in definitions.
interface IncompletePhiInterface<P> {
    readonly original: P;
    readonly renamed: P;
}

// Walking is which of run's three visits a place is in: 'Uses' renames an instruction's uses and passes
// over its definitions, 'Definitions' the reverse, and 'Terminal' renames a terminal's places, uses and
// definitions both.
type WalkingType = 'Uses' | 'Definitions' | 'Terminal';

class SingleAssignmentBuilder<F, B, P> {
    private readonly graph: GraphInterface<F, B, P>;
    private readonly fn: F;

    private readonly states = new Map<BlockIdType, SingleAssignmentStateInterface<P>>();

    // unsealedPredecessors counts, per block, how many predecessors have not yet been processed. A block
    // whose count reaches zero is sealed and its incomplete phis can be filled. Absent from the map
    // means "not yet decremented", which is not the same as zero; the read sites start from the
    // predecessor count on first touch for exactly that reason.
    private readonly unsealedPredecessors = new Map<BlockIdType, number>();

    // unknown holds bindings a lookup walked off the entry block without finding. In cohere they are
    // globals, or captures from an enclosing function; in Adamic, which keeps both out of its graph, a
    // variable read before anything defines it. They are left un-renamed: there is no definition in this
    // function to merge, so a phi over them would be an invention.
    private readonly unknown = new Map<DeclarationIdType, boolean>();

    // visited records blocks already processed, so sealing only fills phis for a block whose body has
    // actually been walked.
    private readonly visited = new Map<BlockIdType, boolean>();

    // walking, walkingBlock and walkingContextStore are where run's walk is, read by visitPlace.
    private walking: WalkingType = 'Uses';
    private walkingBlock: BlockIdType = 0;
    private walkingContextStore = false;

    constructor(graph: GraphInterface<F, B, P>, fn: F) {
        this.graph = graph;
        this.fn = fn;
    }

    run(): void {
        // Parameters are definitions in the entry block. Renaming them is what makes a reassigned
        // parameter behave like any other binding rather than like a global.
        const entryId = this.graph.entry(this.fn);
        if(this.graph.block(this.fn, entryId) === undefined) {
            return;
        }
        this.states.set(entryId, newSingleAssignmentState<P>());
        const params = this.graph.params(this.fn);
        for(let index = 0; index < params.length; index++) {
            params[index] = this.defineIn(entryId, params[index] ?? panic('a parameter index past its end'));
        }

        // One visitor for every place, which reads what to do from this.walking.
        const visit = (place: P, role: RoleType): P => this.visitPlace(place, role);

        for(const block of this.graph.blocks(this.fn)) {
            const blockId = this.graph.id(block);
            this.walkingBlock = blockId;
            if(!this.states.has(blockId)) {
                this.states.set(blockId, newSingleAssignmentState<P>());
            }
            this.visited.set(blockId, true);

            const count = this.graph.instructionCount(this.fn, block);
            for(let index = 0; index < count; index++) {
                // Uses first, then definitions. `x = x + 1` must read the old x before the store mints the
                // new one; visiting in the other order would make the increment read itself.
                this.walking = 'Uses';
                this.graph.eachInstructionPlace(this.fn, block, index, visit);
                // Only a context store reuses its binding's definition. The guard keys on the declaration,
                // and a function expression assigned to the same declaration would otherwise reuse the
                // identifier too, measured as a duplicate definition of a function value on
                // `let n = 1; const g = () => n; n = 2`.
                this.walking = 'Definitions';
                this.walkingContextStore = this.graph.isContextStore(this.fn, block, index);
                this.graph.eachInstructionPlace(this.fn, block, index, visit);
            }

            this.walking = 'Terminal';
            this.graph.eachTerminalPlace(block, visit);

            // Seal each successor that this block was the last unprocessed predecessor of.
            for(const edge of edgesOf(this.graph, block)) {
                if(edge.edge === 'Fallthrough') {
                    continue;
                }
                const successor = this.graph.block(this.fn, edge.successor);
                if(successor === undefined) {
                    continue;
                }
                const remaining =
                    (this.unsealedPredecessors.get(edge.successor) ?? this.graph.predecessors(successor).length) - 1;
                this.unsealedPredecessors.set(edge.successor, remaining);
                if(remaining === 0 && this.visited.get(edge.successor) === true) {
                    this.fixIncompletePhis(edge.successor);
                }
            }
        }

        // The function's return value is a definition like any other and is read by nothing inside the
        // graph, so it is renamed to whatever reaches the end rather than left pointing at the original.
        this.renameReturns();
    }

    // visitPlace is run's one visitor, doing what this.walking says for the block this.walkingBlock.
    private visitPlace(place: P, role: RoleType): P {
        switch(this.walking) {
            case 'Uses':
                return role === 'Define' ? place : this.useIn(this.walkingBlock, place);
            case 'Definitions':
                return role === 'Define'
                    ? this.defineInMaybeContext(this.walkingBlock, place, this.walkingContextStore)
                    : place;
            case 'Terminal':
                return role === 'Define'
                    ? this.defineIn(this.walkingBlock, place)
                    : this.useIn(this.walkingBlock, place);
        }
    }

    // defineIn mints a fresh value for a definition, records it as the block's current answer, and
    // returns the place renamed to it.
    private defineIn(blockId: BlockIdType, place: P): P {
        return this.defineInMaybeContext(blockId, place, false);
    }

    /*
     * defineInMaybeContext is defineIn, told whether this definition is a context store.
     *
     * A context binding is defined once and every later write reuses that definition. Upstream refuses
     * every rename after the first for a binding it has marked as context (`SSA/EnterSSA.ts:124`), and
     * its two writes to a reassigned-and-captured binding name the same identifier. That shared identity
     * is what puts the declaration and the reassignment in one disjoint class, so a later write reaches
     * the value a closure already captured. Gated on the instruction being a context store: the set is
     * keyed by declaration, and a function expression assigned to the same declaration would otherwise
     * reuse the identifier too, which is a real duplicate definition.
     */
    private defineInMaybeContext(blockId: BlockIdType, place: P, contextStore: boolean): P {
        const binding = this.graph.declaration(this.fn, this.graph.identifierOf(place));
        const state = this.states.get(blockId) ?? panic('a definition in a block run never entered');
        if(contextStore && this.graph.contextual(this.fn, binding)) {
            const existing = state.definitions.get(binding);
            if(existing !== undefined) {
                return this.graph.withIdentifier(place, existing);
            }
        }
        const renamed = this.graph.mint(this.fn, this.graph.identifierOf(place));
        state.definitions.set(binding, renamed);
        return this.graph.withIdentifier(place, renamed);
    }

    // useIn is a use renamed to whatever value reaches this block.
    private useIn(blockId: BlockIdType, place: P): P {
        return this.graph.withIdentifier(place, this.valueAt(place, blockId));
    }

    /*
     * valueAt is Braun's lookup: which value does this binding hold on entry to this block.
     *
     * The order of the cases is load-bearing and is Braun's:
     *
     *  1. Defined here already: the local answer wins.
     *  2. No predecessors: the entry block, and the binding was never defined. It is a global or a
     *     capture; record it and hand back the original rather than inventing a definition.
     *  3. Some predecessor unprocessed: a loop. Mint an incomplete phi. Recording it in definitions
     *     before recursing is what terminates the walk around the cycle.
     *  4. Exactly one predecessor: no merge, so recurse and cache. A phi here would be redundant by
     *     construction.
     *  5. Several predecessors: a real join. Mint the result, record it, then collect operands; a
     *     predecessor whose lookup comes back around to this block must find the result already bound.
     */
    private valueAt(place: P, blockId: BlockIdType): IdentifierIdType {
        const original = this.graph.identifierOf(place);
        const binding = this.graph.declaration(this.fn, original);
        if(this.unknown.get(binding) === true) {
            return original;
        }

        let state = this.states.get(blockId);
        if(state === undefined) {
            state = newSingleAssignmentState<P>();
            this.states.set(blockId, state);
        }
        const defined = state.definitions.get(binding);
        if(defined !== undefined) {
            return defined;
        }

        const block = this.graph.block(this.fn, blockId);
        if(block === undefined) {
            return original;
        }

        const predecessors = this.graph.predecessors(block);
        if(predecessors.length === 0) {
            this.unknown.set(binding, true);
            return original;
        }

        const remaining = this.unsealedPredecessors.get(blockId) ?? predecessors.length;
        if(remaining > 0) {
            const renamed = this.graph.mint(this.fn, original);
            state.definitions.set(binding, renamed);
            state.incompletePhis.push({ original: place, renamed: this.graph.withIdentifier(place, renamed) });
            return renamed;
        }

        if(predecessors.length === 1) {
            const renamed = this.valueFromPredecessor(place, predecessors[0] ?? panic('one predecessor and none'));
            state.definitions.set(binding, renamed);
            return renamed;
        }

        const renamed = this.graph.mint(this.fn, original);
        state.definitions.set(binding, renamed);
        this.addPhi(blockId, place, this.graph.withIdentifier(place, renamed));
        return renamed;
    }

    /*
     * valueFromPredecessor is the value a binding holds when control arrives at blockId from predecessor.
     *
     * Every lookup across an edge comes through here, and today it is the predecessor's value at its end
     * whatever the edge, which is right for every edge but an exceptional one. An exceptional edge leaves
     * its block from the throwing instruction, before that instruction's definitions happen, so the
     * handler should see the definitions from just before it. That is #2yz9ra9, and its fix lands here.
     * The Go passes the block control arrives at too, for that fix; nothing reads it yet, so it's left out.
     */
    private valueFromPredecessor(place: P, predecessor: BlockIdType): IdentifierIdType {
        return this.valueAt(place, predecessor);
    }

    // addPhi collects one operand per predecessor and attaches the phi to the block.
    //
    // original is the place a lookup reads, and each operand is that place naming the predecessor's
    // value, so whatever else the place carries rides along on every operand.
    private addPhi(blockId: BlockIdType, original: P, renamed: P): void {
        const block = this.graph.block(this.fn, blockId);
        if(block === undefined) {
            return;
        }

        let operands: PhiOperandInterface<P>[] = [];
        for(const predecessorId of this.graph.predecessors(block)) {
            operands = withPhiOperand(
                operands,
                predecessorId,
                this.graph.withIdentifier(original, this.valueFromPredecessor(original, predecessorId)),
            );
        }

        this.graph.setPhis(block, [...this.graph.phis(block), { place: renamed, operands }]);
    }

    // fixIncompletePhis fills in the operands of every phi minted before this block was sealed.
    private fixIncompletePhis(blockId: BlockIdType): void {
        const state = this.states.get(blockId);
        if(state === undefined) {
            return;
        }
        const pending = state.incompletePhis;
        state.incompletePhis = [];
        for(const phi of pending) {
            this.addPhi(blockId, phi.original, phi.renamed);
        }
    }

    /*
     * renameReturns rewrites the function's returns place to the value that reaches the exit.
     *
     * Every `return` stores into one identifier, so after renaming the stores there are several values
     * and the returns place still names the original. It is resolved against the blocks that actually
     * end in a return, which is where the value is observable. An IR with no returns place skips this.
     */
    private renameReturns(): void {
        const returns = this.graph.returns(this.fn);
        if(returns === undefined) {
            return;
        }
        const binding = this.graph.declaration(this.fn, this.graph.identifierOf(returns));
        if(this.unknown.get(binding) === true) {
            return;
        }
        for(const block of this.graph.blocks(this.fn)) {
            if(!this.graph.endsInReturn(block)) {
                continue;
            }
            const renamed = this.states.get(this.graph.id(block))?.definitions.get(binding);
            if(renamed !== undefined) {
                this.graph.setReturns(this.fn, this.graph.withIdentifier(returns, renamed));
                return;
            }
        }
    }
}

/*
 * construct converts one function to single static assignment form, in place, then eliminates the
 * redundant phis that placement produces. It does not re-establish the graph's invariants or recurse
 * into nested functions; an IR's own construct does both around it.
 *
 * After it returns: every identifier that a source binding takes is written exactly once, every use
 * names the definition that actually reaches it, and each block's phis hold a phi wherever a binding's
 * value depends on which predecessor control arrived from.
 */
export function construct<F, B, P>(graph: GraphInterface<F, B, P>, fn: F): void {
    /*
     * Phis from a previous run are dropped, because this pass appends them and cannot reconcile what it
     * did not mint. A caller that restructures the graph and runs construction again is the case:
     * inlining an immediately invoked function expression does exactly that. The phis left behind name
     * identifiers from before the renumbering, so they define a value nothing produces. Braun's
     * algorithm derives every phi it needs from the graph, so nothing is lost by discarding the previous
     * answer, and eliminateRedundantPhis below then sees only phis this run placed.
     *
     * The Go skips a nil block here; a block array here holds blocks, never nothing.
     */
    for(const block of graph.blocks(fn)) {
        graph.setPhis(block, []);
    }

    const builder = new SingleAssignmentBuilder(graph, fn);
    builder.run();

    // Elimination is part of construction rather than an optional follow-up, because Braun's placement
    // cannot avoid producing redundant phis and their share is not marginal: measured over 1,945
    // functions of real TypeScript, 24,418 phis before elimination and 2,746 after.
    eliminateRedundantPhis(graph, fn);
}
