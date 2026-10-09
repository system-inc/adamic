import {ReachedNormally, ReachedOnlyAfterReturn} from './icons';

// The live path: this reference executes.
export function usesReachable(): string {
    return ReachedNormally;
}

// The dead path: the reference is real and resolves to the same symbol, but an
// unconditional return above it means nothing can ever evaluate it.
export function usesUnreachable(): string {
    return 'early';
    return ReachedOnlyAfterReturn;
}
