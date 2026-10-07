/*
 * A port of cohere's mutation_aliasing module (mutation_aliasing.go) to Adamic 0.1: React's mutation and
 * aliasing model, shared by two intermediate representations, cohere's high-level IR (the React
 * Compiler's) and Adamic's flow graph.
 *
 * It holds two things, both named after React's own passes, InferMutationAliasingEffects and
 * InferMutationAliasingRanges:
 *
 *   - The effect vocabulary: AliasingEffectInterface and its kinds, what an instruction does to the
 *     values it reads and writes. Each IR makes its own effects, cohere from React's signature table
 *     keyed by the callee's name and Adamic from its IR, where every runtime operation has one known
 *     effect. Both speak this vocabulary.
 *   - Mutable ranges (ranges.ts): the alias graph those effects build, the mutate worklist over it, and
 *     the range of evaluation over which each value is still being written.
 *
 * The ranges are React Compiler semantics, held to parity with upstream, so cohere's behaviour is the
 * default. Where the two IRs differ, the difference is a member of GraphInterface or a field of
 * OptionsInterface rather than a fork of the algorithm. GraphInterface says which member is which seam.
 *
 * It is a module of its own, beside static_single_assignment rather than inside it, because that module
 * holds no IR's meaning and this one is React's.
 *
 * Where the port differs from the Go, and why:
 *
 *   - Each type here carries the house suffix (GraphInterface, AliasingEffectKindType); its comment
 *     names the Go type it mirrors.
 *   - AliasingEffectKind and EffectValueKind are unions of their names, where Go's are integer enums.
 *     Their String methods are aliasingEffectKindName and effectValueKindName, and AliasingEffectKinds
 *     lists the kinds in Go's declaration order.
 *   - An effect's from is the place or undefined. Go's HasFrom says whether From is set, because Go
 *     has no undefined and a zero IdentifierId is a real identifier; here hasFrom is from !== undefined.
 *   - Go answers a seam that may have nothing to say with a second result, (order, ok) and the like.
 *     Here the member answers undefined: instructionOrder, returnValue, storedContextValue, and closure,
 *     whose captures and frozen answer are one ClosureInterface.
 *   - Go's Graph is a type parameter so its calls compile to direct ones. Here a pass takes the
 *     interface itself, as the single assignment port's passes do.
 */

import type {
    EvaluationOrderType,
    GraphInterface as SingleAssignmentGraphInterface,
    IdentifierIdType,
} from '../static_single_assignment/static_single_assignment.ts';

/*
 * AliasingEffectKindType (Go's AliasingEffectKind) is what one effect does. These are React's
 * `AliasingEffect` variants.
 *
 * This is the real output type of the effect inference and it is not `Effect`. See cohere's effects.go:
 * the two are different types, this one carries from and into places, and the scalar `Effect` is a
 * per-operand projection of a whole list of these, computed later from the ranges.
 *
 * The declaration order carries no meaning and there is no join over it. Unlike `ValueKind`, which
 * cohere's `immutability.go` merges with a real lattice operation, these are not merged at all: an
 * instruction produces a list of effects and they are all applied in order. A pass that tried to reduce
 * the list to one element would lose the from/into pairing that is the entire content.
 *
 *   - 'Create' makes a new value of a given kind at into. from is unset.
 *   - 'CreateFrom' makes a new value at into with the same kind as from.
 *   - 'Assign' is `into = from`, a direct assignment.
 *   - 'Alias' means mutating into implies mutating from. Direct aliasing.
 *   - 'MaybeAlias' is a potential aliasing relationship, used for unknown callees.
 *   - 'Capture' is information flow from from into into without aliasing.
 *   - 'ImmutableCapture' is data flow that escape analysis sees and mutable ranges do not. Projects to
 *     Read rather than to Capture.
 *   - 'Mutate' mutates the value and its direct aliases.
 *   - 'MutateConditionally' mutates only if the value is mutable.
 *   - 'MutateTransitive' mutates the value and everything it transitively captured.
 *   - 'MutateTransitiveConditionally' is the conditional form, and is the default applied to every
 *     operand of a call with no known signature.
 *   - 'Freeze' marks the value and its direct aliases as frozen.
 *   - 'Apply' is an unresolved call. Upstream replaces every Apply with more precise effects before the
 *     ranges pass runs, and raises an invariant if one survives; cohere's inference resolves them the
 *     same way and never emits one. Kept because `EffectGapInterproceduralParameters` is exactly the
 *     case where upstream would resolve one and cohere does not.
 */
export type AliasingEffectKindType =
    | 'Create'
    | 'CreateFrom'
    | 'Assign'
    | 'Alias'
    | 'MaybeAlias'
    | 'Capture'
    | 'ImmutableCapture'
    | 'Mutate'
    | 'MutateConditionally'
    | 'MutateTransitive'
    | 'MutateTransitiveConditionally'
    | 'Freeze'
    | 'Apply';

// AliasingEffectKinds are the kinds in Go's declaration order.
export const AliasingEffectKinds: readonly AliasingEffectKindType[] = [
    'Create',
    'CreateFrom',
    'Assign',
    'Alias',
    'MaybeAlias',
    'Capture',
    'ImmutableCapture',
    'Mutate',
    'MutateConditionally',
    'MutateTransitive',
    'MutateTransitiveConditionally',
    'Freeze',
    'Apply',
];

// aliasingEffectKindName is a kind's name as Go's String prints it.
export function aliasingEffectKindName(kind: AliasingEffectKindType): string {
    switch(kind) {
        case 'Create':
            return 'create';
        case 'CreateFrom':
            return 'create-from';
        case 'Assign':
            return 'assign';
        case 'Alias':
            return 'alias';
        case 'MaybeAlias':
            return 'maybe-alias';
        case 'Capture':
            return 'capture';
        case 'ImmutableCapture':
            return 'immutable-capture';
        case 'Mutate':
            return 'mutate';
        case 'MutateConditionally':
            return 'mutate-conditionally';
        case 'MutateTransitive':
            return 'mutate-transitive';
        case 'MutateTransitiveConditionally':
            return 'mutate-transitive-conditionally';
        case 'Freeze':
            return 'freeze';
        case 'Apply':
            return 'apply';
    }
}

/*
 * isMutation (Go's IsMutation) reports whether an effect of this kind widens a value's mutable range.
 *
 * This is the exact predicate the ranges pass reads to build its `mutations` list, at
 * `infer_mutation_aliasing_ranges.rs` where `mutations` is populated from effects matching
 * `Mutate | MutateConditionally | MutateTransitive | MutateTransitiveConditionally`. Exposed here rather
 * than restated there so the two passes cannot drift.
 */
export function isMutation(kind: AliasingEffectKindType): boolean {
    switch(kind) {
        case 'Mutate':
        case 'MutateConditionally':
        case 'MutateTransitive':
        case 'MutateTransitiveConditionally':
            return true;
        case 'Create':
        case 'CreateFrom':
        case 'Assign':
        case 'Alias':
        case 'MaybeAlias':
        case 'Capture':
        case 'ImmutableCapture':
        case 'Freeze':
        case 'Apply':
            return false;
    }
}

/*
 * isAliasing (Go's IsAliasing) reports whether an effect of this kind creates a from/into data-flow edge.
 *
 * These are the five variants the ranges pass treats identically when projecting to a scalar `Effect`:
 * all five give from either Capture or Read depending on whether into is still mutable, and give into
 * Store. See cohere's `ProjectEffects`.
 */
export function isAliasing(kind: AliasingEffectKindType): boolean {
    switch(kind) {
        case 'Assign':
        case 'Alias':
        case 'Capture':
        case 'CreateFrom':
        case 'MaybeAlias':
            return true;
        case 'Create':
        case 'ImmutableCapture':
        case 'Mutate':
        case 'MutateConditionally':
        case 'MutateTransitive':
        case 'MutateTransitiveConditionally':
        case 'Freeze':
        case 'Apply':
            return false;
    }
}

/*
 * AliasingEffectInterface (Go's AliasingEffect) is one effect an instruction has, over an IR's place
 * type P.
 *
 * from is the source of a data-flow edge and is undefined for the mutation and create variants, where
 * only into is set. That asymmetry is upstream's: its enum gives each variant its own fields.
 */
export interface AliasingEffectInterface<P> {
    readonly kind: AliasingEffectKindType;
    // from is the source place of a data-flow edge, or undefined when the effect has none.
    readonly from: P | undefined;
    // into is the target: the value created, mutated, or flowed into.
    readonly into: P;
    // value is the kind created, meaningful only for 'Create'.
    readonly value: EffectValueKindType;
}

// createEffect (Go's CreateEffect) is the Create effect, which every instruction producing a value emits
// first.
export function createEffect<P>(into: P, kind: EffectValueKindType): AliasingEffectInterface<P> {
    return { kind: 'Create', from: undefined, into, value: kind };
}

// flowEffect (Go's FlowEffect) is one from/into effect.
export function flowEffect<P>(kind: AliasingEffectKindType, from: P, into: P): AliasingEffectInterface<P> {
    return { kind, from, into, value: 'Mutable' };
}

// mutationEffect (Go's MutationEffect) is one mutation effect, which names only the value it writes.
export function mutationEffect<P>(kind: AliasingEffectKindType, value: P): AliasingEffectInterface<P> {
    return { kind, from: undefined, into: value, value: 'Mutable' };
}

/*
 * EffectValueKindType (Go's EffectValueKind) is the abstract kind a Create effect produces.
 *
 * This is upstream's `ValueKind` and it is deliberately not the same type as the six-element lattice
 * cohere's `immutability.go` builds for itself. That rule's lattice is a private type with a join tuned
 * to what it reports; this kind also feeds the range graph's phi refinement.
 *
 *   - 'Mutable' is a freshly created value this code may still write to.
 *   - 'Primitive' is a number, string, boolean or similar.
 *   - 'Frozen' is a value that must not be mutated from here on.
 *   - 'MaybeFrozen' and 'Global' are upstream's.
 */
export type EffectValueKindType = 'Mutable' | 'Primitive' | 'Frozen' | 'MaybeFrozen' | 'Global';

// EffectValueKinds are the kinds in Go's declaration order.
export const EffectValueKinds: readonly EffectValueKindType[] = [
    'Mutable',
    'Primitive',
    'Frozen',
    'MaybeFrozen',
    'Global',
];

// effectValueKindName is a kind's name as Go's String prints it.
export function effectValueKindName(kind: EffectValueKindType): string {
    switch(kind) {
        case 'Primitive':
            return 'primitive';
        case 'Frozen':
            return 'frozen';
        case 'MaybeFrozen':
            return 'maybe-frozen';
        case 'Global':
            return 'global';
        case 'Mutable':
            return 'mutable';
    }
}

// ClosureInterface is what GraphInterface.closure answers for a closure: its captures, which a later
// Freeze of the closure freezes, and whether React creates the closure Frozen.
export interface ClosureInterface<P> {
    readonly captures: readonly P[];
    readonly frozen: boolean;
}

/*
 * GraphInterface (Go's Graph) is what the alias graph and the ranges need of an intermediate
 * representation: everything single assignment needs, and one member more per seam where cohere's IR
 * and Adamic's differ.
 *
 * An IR implements it on a value of its own, which carries the IR's effect table, and calls a pass with
 * that value and its function. The passes hold no IR type; every read of one goes through here.
 */
export interface GraphInterface<F, B, P> extends SingleAssignmentGraphInterface<F, B, P> {
    // instructionOrder is the evaluation order of a block's instruction at index, and undefined when the
    // block names an instruction the function does not hold, which every pass skips.
    readonly instructionOrder: (fn: F, block: B, index: number) => EvaluationOrderType | undefined;
    // terminalOrder is the evaluation order of a block's terminal.
    readonly terminalOrder: (block: B) => EvaluationOrderType;

    // effects are the effects of a block's instruction at index, in the order they apply. This is the
    // effects input: cohere's come from its signature-table inference, Adamic's from its own.
    readonly effects: (fn: F, block: B, index: number) => readonly AliasingEffectInterface<P>[];

    // parametersFrozen reports whether the function's parameters arrive Frozen. React's answer is a
    // component's or a hook's, since React owns the arguments it passes them; Adamic's is never.
    readonly parametersFrozen: (fn: F) => boolean;
    // context are the values a nested function captures, defined on entry like its parameters. An IR
    // that keeps captured variables out of its graph has none.
    readonly context: (fn: F) => readonly P[];
    // returnValue is the value a block's Return terminal returns, which the pass aliases into the
    // returns place, and undefined for a block that does not return one. Asked only of an IR whose
    // returns gives a place.
    readonly returnValue: (block: B) => P | undefined;
    // storedContextValue is the value a block's instruction at index stores into a captured binding,
    // when it is such a store: React widens its range from the instruction's shape alone.
    readonly storedContextValue: (fn: F, block: B, index: number) => IdentifierIdType | undefined;
    /*
     * closure is asked at every Create of a block's instruction at index whose into is into. When the
     * instruction makes a closure whose value is into, it answers the closure's captures, which a later
     * Freeze of the closure freezes, and whether React creates the closure Frozen: every capture
     * immutable and its body only reading them. kinds are the abstract kinds the pass holds at that
     * point, an immutable kind for each value that is not Mutable, and are only read. An IR with no such
     * rule answers undefined.
     */
    readonly closure: (
        fn: F,
        block: B,
        index: number,
        into: IdentifierIdType,
        kinds: ReadonlyMap<IdentifierIdType, EffectValueKindType>,
    ) => ClosureInterface<P> | undefined;
}

// OptionsInterface (Go's Options) is what a caller sets on the pass, rather than what its IR answers.
export interface OptionsInterface {
    /*
     * parametersDefinedOnEntry opens every parameter's range at the function's first instruction, rather
     * than at its first read, and gives one nothing widened that instruction alone.
     *
     * Adamic's rule (its 4748a636), and false by default, so cohere keeps React's: there, a component's or
     * hook's parameters are Frozen and nothing the body does mutates them. An Adamic parameter is an owned
     * or borrowed value its body may mutate, and one mutated before its first read (through another value
     * that reaches it: this.list and a node of it, say) would otherwise be mutated outside its range. A
     * parameter then always has a range, so a mutation the graph missed is outside it rather than on a
     * range the pass declined to set.
     */
    readonly parametersDefinedOnEntry: boolean;

    // contextKinds seeds the abstract kind of captured values, for a caller that knows them: cohere's
    // read-only-closure probe passes the kinds its enclosing function holds for the captures. A value
    // absent here, or Mutable, is created Mutable.
    readonly contextKinds: ReadonlyMap<IdentifierIdType, EffectValueKindType>;
}
