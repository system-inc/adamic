// Additional operations in the standalone Adamic library, not project TypeScript libs.

// ES2025 Set operations. Stage 0 accepts concrete library Sets with matching representations.
interface Set<T> {
	union<U>(other: ReadonlySetLike<U>): Set<T | U>;
	intersection<U>(other: ReadonlySetLike<U>): Set<T & U>;
	difference<U>(other: ReadonlySetLike<U>): Set<T>;
	symmetricDifference<U>(other: ReadonlySetLike<U>): Set<T | U>;
	isSubsetOf<U>(other: ReadonlySetLike<U>): boolean;
	isSupersetOf<U>(other: ReadonlySetLike<U>): boolean;
	isDisjointFrom<U>(other: ReadonlySetLike<U>): boolean;
}
interface ReadonlySet<T> {
	union<U>(other: ReadonlySetLike<U>): Set<T | U>;
	intersection<U>(other: ReadonlySetLike<U>): Set<T & U>;
	difference<U>(other: ReadonlySetLike<U>): Set<T>;
	symmetricDifference<U>(other: ReadonlySetLike<U>): Set<T | U>;
	isSubsetOf<U>(other: ReadonlySetLike<U>): boolean;
	isSupersetOf<U>(other: ReadonlySetLike<U>): boolean;
	isDisjointFrom<U>(other: ReadonlySetLike<U>): boolean;
}

interface ReadonlySetLike<T> {
	readonly size: number;
	readonly has: (value: T) => boolean;
	readonly keys: () => Iterator<T>;
}

