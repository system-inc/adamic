import { Scope } from './one_var_scope.ts';
export class BlockScope {
    readonly kinds: Scope[] = [new Scope(), new Scope(), new Scope(), new Scope()];
}
