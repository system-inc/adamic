// One type identity and its checker metadata. Links are numeric identities.
export class TypeFact {
    readonly id: number;
    readonly flags: number;
    readonly name: string;
    readonly error: boolean;
    readonly target: number;
    readonly array: boolean;
    readonly tuple: number;
    readonly constraint: number;
    readonly element: number;
    readonly parts: readonly number[];
    readonly arguments: readonly number[];
    constructor(
        id: number,
        flags: number,
        name: string,
        error: boolean,
        target: number,
        array: boolean,
        tuple: number,
        constraint: number,
        element: number,
        parts: readonly number[],
        arguments_: readonly number[],
    ) {
        this.id = id;
        this.flags = flags;
        this.name = name;
        this.error = error;
        this.target = target;
        this.array = array;
        this.tuple = tuple;
        this.constraint = constraint;
        this.element = element;
        this.parts = parts;
        this.arguments = arguments_;
    }
}
