export interface Coded {
    readonly code: number;
}
export function copied(source: Coded | undefined) {
    return { ...source, label: "copy" };
}
