export interface Coded {
    readonly code: number;
}
export function copied(source: Coded | undefined) {
    return { ...source };
}
export function present(value: { readonly code?: number }): boolean {
    return "code" in value;
}
