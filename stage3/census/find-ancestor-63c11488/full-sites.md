## Full-file inventories

These are separate full-mode measurements on the compiler directory, selecting exactly one adapted file with LATENT_ONLY_FILE. Continuation Boundary records are retained in raw streams but are not counted as refusal sites here. A deferred generic body with no finding is unexamined, not proven to lower.

### findAncestor

```typescript
export function findAncestor<T extends Node>(node: Node | undefined, callback: (element: Node) => element is T): T | undefined;
```

```typescript
export function findAncestor(node: Node | undefined, callback: (element: Node) => boolean | "quit"): Node | undefined;
```

```typescript
export function findAncestor(node: Node | undefined, callback: (element: Node) => boolean | "quit"): Node | undefined {
    while (node) {
        const result = callback(node);
        if (result === "quit") {
            return undefined;
        }
        else if (result) {
            return node;
        }
        node = node.parent;
    }
    return undefined;
}
```

| Side | Unit | Status | Instantiated |
| --- | --- | --- | --- |
| parent | compiler/utilitiesPublic.ts:786:1 | deferred_generic_instantiation | True |
| parent | compiler/utilitiesPublic.ts:787:1 | attempted | False |
| parent | compiler/utilitiesPublic.ts:788:1 | attempted | False |
| head | compiler/utilitiesPublic.ts:786:1 | deferred_generic_instantiation | True |
| head | compiler/utilitiesPublic.ts:787:1 | attempted | False |
| head | compiler/utilitiesPublic.ts:788:1 | attempted | False |

| Side | Every recorded site | Phase | Recording units |
| --- | --- | --- | --- |
| parent | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | lowering | compiler/utilitiesPublic.ts:787:1, compiler/utilitiesPublic.ts:788:1 |
| parent | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | lowering | compiler/utilitiesPublic.ts:1566:1 |
| head | No site finding recorded | | |

### Every changed site in utilitiesPublic.ts

| Category | Old | New |
| --- | --- | --- |
| lowered | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | none |
| lowered | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | none |
| new_findings | none | NotYet: checked predicate overload target ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; } at compiler/utilitiesPublic.ts:1562:1 |

### orPredicate2

```typescript
export function orPredicate2<P, R1 extends P, R2 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2): (p: P) => p is R1 | R2 {
    return p => f1(p) || f2(p);
}
```

| Side | Unit | Status | Instantiated |
| --- | --- | --- | --- |
| parent | compiler/core.ts:2462:1 | deferred_generic_instantiation | False |
| head | compiler/core.ts:2462:1 | deferred_generic_instantiation | False |

| Side | Every recorded site | Phase | Recording units |
| --- | --- | --- | --- |
| parent | No site finding recorded | | |
| head | No site finding recorded | | |

### orPredicate3

```typescript
export function orPredicate3<P, R1 extends P, R2 extends P, R3 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2, f3: (p3: P) => p3 is R3): (p: P) => p is R1 | R2 | R3 {
    return p => f1(p) || f2(p) || f3(p);
}
```

| Side | Unit | Status | Instantiated |
| --- | --- | --- | --- |
| parent | compiler/core.ts:2466:1 | deferred_generic_instantiation | False |
| head | compiler/core.ts:2466:1 | deferred_generic_instantiation | False |

| Side | Every recorded site | Phase | Recording units |
| --- | --- | --- | --- |
| parent | No site finding recorded | | |
| head | No site finding recorded | | |

### or

```typescript
export function or<T extends unknown[], U>(...fs: [(...args: T) => U, ...((...args: T) => U)[]]): (...args: T) => U;
```

```typescript
export function or<T extends unknown[], U>(...fs: ((...args: T) => U)[]): (...args: T) => U | undefined;
```

```typescript
export function or<T extends unknown[], U>(...fs: ((...args: T) => U)[]): (...args: T) => U | undefined {
    return (...args) => {
        let lastResult: U | undefined;
        for (const f of fs) {
            lastResult = f(...args);
            if (lastResult) {
                return lastResult;
            }
        }
        return lastResult;
    };
}
```

| Side | Unit | Status | Instantiated |
| --- | --- | --- | --- |
| parent | compiler/core.ts:2470:1 | deferred_generic_instantiation | False |
| parent | compiler/core.ts:2472:1 | deferred_generic_instantiation | False |
| parent | compiler/core.ts:2474:1 | deferred_generic_instantiation | False |
| head | compiler/core.ts:2470:1 | deferred_generic_instantiation | False |
| head | compiler/core.ts:2472:1 | deferred_generic_instantiation | False |
| head | compiler/core.ts:2474:1 | deferred_generic_instantiation | False |

| Side | Every recorded site | Phase | Recording units |
| --- | --- | --- | --- |
| parent | No site finding recorded | | |
| head | No site finding recorded | | |

### Every changed site in core.ts

| Category | Old | New |
| --- | --- | --- |
| None | | |

### Helper sites across entry and focused contexts

| Context | Side | Helper | Site and exact reason | Phase | Recording unit |
| --- | --- | --- | --- | --- | --- |
| entry | parent | or | NotYet: overload 1 of or with a non-array rest parameter at compiler/core.ts:2470:1 | lowering | compiler/utilities.ts:10054:1 |
| entry | head | or | NotYet: overload 1 of or with a non-array rest parameter at compiler/core.ts:2470:1 | lowering | compiler/utilities.ts:10054:1 |

