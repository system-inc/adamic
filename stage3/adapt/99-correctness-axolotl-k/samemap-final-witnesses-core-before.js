import { CharacterCodes, Comparison, Debug, } from "./_namespaces/ts.js";
/* eslint-disable @typescript-eslint/prefer-for-of */
/** @internal */
export const emptyArray = [];
/** @internal */
export const emptyMap = new Map();
/** @internal */
export function length(array) {
    return array !== undefined ? array.length : 0;
}
/**
 * Iterates through 'array' by index and performs the callback on each element of array until the callback
 * returns a truthy value, then returns that value.
 * If no such value is found, the callback is applied to each element of array and undefined is returned.
 *
 * @internal
 */
export function forEach(array, callback) {
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            const result = callback(array[i], i);
            if (result) {
                return result;
            }
        }
    }
    return undefined;
}
/**
 * Like `forEach`, but iterates in reverse order.
 *
 * @internal
 */
export function forEachRight(array, callback) {
    if (array !== undefined) {
        for (let i = array.length - 1; i >= 0; i--) {
            const result = callback(array[i], i);
            if (result) {
                return result;
            }
        }
    }
    return undefined;
}
/**
 * Like `forEach`, but suitable for use with numbers and strings (which may be falsy).
 *
 * @internal
 */
export function firstDefined(array, callback) {
    if (array === undefined) {
        return undefined;
    }
    for (let i = 0; i < array.length; i++) {
        const result = callback(array[i], i);
        if (result !== undefined) {
            return result;
        }
    }
    return undefined;
}
/** @internal */
export function firstDefinedIterator(iter, callback) {
    for (const value of iter) {
        const result = callback(value);
        if (result !== undefined) {
            return result;
        }
    }
    return undefined;
}
/** @internal */
export function reduceLeftIterator(iterator, f, initial) {
    let result = initial;
    if (iterator) {
        let pos = 0;
        for (const value of iterator) {
            result = f(result, value, pos);
            pos++;
        }
    }
    return result;
}
/** @internal */
export function zipWith(arrayA, arrayB, callback) {
    const result = [];
    Debug.assertEqual(arrayA.length, arrayB.length);
    for (let i = 0; i < arrayA.length; i++) {
        result.push(callback(arrayA[i], arrayB[i], i));
    }
    return result;
}
/**
 * Creates a new array with `element` interspersed in between each element of `input`
 * if there is more than 1 value in `input`. Otherwise, returns the existing array.
 *
 * @internal
 */
export function intersperse(input, element) {
    if (input.length <= 1) {
        return input;
    }
    const result = [];
    for (let i = 0, n = input.length; i < n; i++) {
        if (i !== 0)
            result.push(element);
        result.push(input[i]);
    }
    return result;
}
export function every(array, callback) {
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            if (!callback(array[i], i)) {
                return false;
            }
        }
    }
    return true;
}
/** @internal */
export function find(array, predicate, startIndex) {
    if (array === undefined)
        return undefined;
    for (let i = startIndex ?? 0; i < array.length; i++) {
        const value = array[i];
        if (predicate(value, i)) {
            return value;
        }
    }
    return undefined;
}
/** @internal */
export function findLast(array, predicate, startIndex) {
    if (array === undefined)
        return undefined;
    for (let i = startIndex ?? array.length - 1; i >= 0; i--) {
        const value = array[i];
        if (predicate(value, i)) {
            return value;
        }
    }
    return undefined;
}
/**
 * Works like Array.prototype.findIndex, returning `-1` if no element satisfying the predicate is found.
 *
 * @internal
 */
export function findIndex(array, predicate, startIndex) {
    if (array === undefined)
        return -1;
    for (let i = startIndex ?? 0; i < array.length; i++) {
        if (predicate(array[i], i)) {
            return i;
        }
    }
    return -1;
}
/** @internal */
export function findLastIndex(array, predicate, startIndex) {
    if (array === undefined)
        return -1;
    for (let i = startIndex ?? array.length - 1; i >= 0; i--) {
        if (predicate(array[i], i)) {
            return i;
        }
    }
    return -1;
}
/** @internal */
export function contains(array, value, equalityComparer = equateValues) {
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            if (equalityComparer(array[i], value)) {
                return true;
            }
        }
    }
    return false;
}
/** @internal */
export function indexOfAnyCharCode(text, charCodes, start) {
    for (let i = start ?? 0; i < text.length; i++) {
        if (contains(charCodes, text.charCodeAt(i))) {
            return i;
        }
    }
    return -1;
}
/** @internal */
export function countWhere(array, predicate) {
    let count = 0;
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            const v = array[i];
            if (predicate(v, i)) {
                count++;
            }
        }
    }
    return count;
}
/** @internal */
export function filter(array, f) {
    if (array !== undefined) {
        const len = array.length;
        let i = 0;
        while (i < len && f(array[i]))
            i++;
        if (i < len) {
            const result = array.slice(0, i);
            i++;
            while (i < len) {
                const item = array[i];
                if (f(item)) {
                    result.push(item);
                }
                i++;
            }
            return result;
        }
    }
    return array;
}
/** @internal */
export function filterMutate(array, f) {
    let outIndex = 0;
    for (let i = 0; i < array.length; i++) {
        if (f(array[i], i, array)) {
            array[outIndex] = array[i];
            outIndex++;
        }
    }
    array.length = outIndex;
}
/** @internal */
export function clear(array) {
    array.length = 0;
}
/** @internal */
export function map(array, f) {
    let result;
    if (array !== undefined) {
        result = [];
        for (let i = 0; i < array.length; i++) {
            result.push(f(array[i], i));
        }
    }
    return result;
}
/** @internal */
export function* mapIterator(iter, mapFn) {
    for (const x of iter) {
        yield mapFn(x);
    }
}
/** @internal */
export function sameMap(array, f) {
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            const item = array[i];
            const mapped = f(item, i);
            if (item !== mapped) {
                const result = array.slice(0, i);
                result.push(mapped);
                for (i++; i < array.length; i++) {
                    result.push(f(array[i], i));
                }
                return result;
            }
        }
    }
    return array;
}
/**
 * Flattens an array containing a mix of array or non-array elements.
 *
 * @param array The array to flatten.
 *
 * @internal
 */
export function flatten(array) {
    const result = [];
    for (let i = 0; i < array.length; i++) {
        const v = array[i];
        if (v) {
            if (isArray(v)) {
                addRange(result, v);
            }
            else {
                result.push(v);
            }
        }
    }
    return result;
}
/**
 * Maps an array. If the mapped value is an array, it is spread into the result.
 *
 * @param array The array to map.
 * @param mapfn The callback used to map the result into one or more values.
 *
 * @internal
 */
export function flatMap(array, mapfn) {
    let result;
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            const v = mapfn(array[i], i);
            if (v) {
                if (isArray(v)) {
                    result = addRange(result, v);
                }
                else {
                    result = append(result, v);
                }
            }
        }
    }
    return result ?? emptyArray;
}
/** @internal */
export function flatMapToMutable(array, mapfn) {
    const result = [];
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            const v = mapfn(array[i], i);
            if (v) {
                if (isArray(v)) {
                    addRange(result, v);
                }
                else {
                    result.push(v);
                }
            }
        }
    }
    return result;
}
/** @internal */
export function* flatMapIterator(iter, mapfn) {
    for (const x of iter) {
        const iter2 = mapfn(x);
        if (!iter2)
            continue;
        yield* iter2;
    }
}
/** @internal */
export function sameFlatMap(array, mapfn) {
    let result;
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            const item = array[i];
            const mapped = mapfn(item, i);
            if (result || item !== mapped || isArray(mapped)) {
                if (!result) {
                    result = array.slice(0, i);
                }
                if (isArray(mapped)) {
                    addRange(result, mapped);
                }
                else {
                    result.push(mapped);
                }
            }
        }
    }
    return result ?? array;
}
/** @internal */
export function mapAllOrFail(array, mapFn) {
    const result = [];
    for (let i = 0; i < array.length; i++) {
        const mapped = mapFn(array[i], i);
        if (mapped === undefined) {
            return undefined;
        }
        result.push(mapped);
    }
    return result;
}
/** @internal */
export function mapDefined(array, mapFn) {
    const result = [];
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            const mapped = mapFn(array[i], i);
            if (mapped !== undefined) {
                result.push(mapped);
            }
        }
    }
    return result;
}
/** @internal */
// eslint-disable-next-line no-restricted-syntax
export function* mapDefinedIterator(iter, mapFn) {
    for (const x of iter) {
        const value = mapFn(x);
        if (value !== undefined) {
            yield value;
        }
    }
}
/** @internal */
export function getOrUpdate(map, key, callback) {
    if (map.has(key)) {
        return map.get(key);
    }
    const value = callback();
    map.set(key, value);
    return value;
}
/** @internal */
export function tryAddToSet(set, value) {
    if (!set.has(value)) {
        set.add(value);
        return true;
    }
    return false;
}
/** @internal */
export function* singleIterator(value) {
    yield value;
}
/** @internal */
export function spanMap(array, keyfn, mapfn) {
    let result;
    if (array !== undefined) {
        result = [];
        const len = array.length;
        let previousKey;
        let key;
        let start = 0;
        let pos = 0;
        while (start < len) {
            while (pos < len) {
                const value = array[pos];
                key = keyfn(value, pos);
                if (pos === 0) {
                    previousKey = key;
                }
                else if (key !== previousKey) {
                    break;
                }
                pos++;
            }
            if (start < pos) {
                const v = mapfn(array.slice(start, pos), previousKey, start, pos);
                if (v) {
                    result.push(v);
                }
                start = pos;
            }
            previousKey = key;
            pos++;
        }
    }
    return result;
}
/** @internal */
export function mapEntries(map, f) {
    if (map === undefined) {
        return undefined;
    }
    const result = new Map();
    map.forEach((value, key) => {
        const [newKey, newValue] = f(key, value);
        result.set(newKey, newValue);
    });
    return result;
}
/** @internal */
export function some(array, predicate) {
    if (array !== undefined) {
        if (predicate !== undefined) {
            for (let i = 0; i < array.length; i++) {
                if (predicate(array[i])) {
                    return true;
                }
            }
        }
        else {
            return array.length > 0;
        }
    }
    return false;
}
/**
 * Calls the callback with (start, afterEnd) index pairs for each range where 'pred' is true.
 *
 * @internal
 */
export function getRangesWhere(arr, pred, cb) {
    let start;
    for (let i = 0; i < arr.length; i++) {
        if (pred(arr[i])) {
            start = start === undefined ? i : start;
        }
        else {
            if (start !== undefined) {
                cb(start, i);
                start = undefined;
            }
        }
    }
    if (start !== undefined)
        cb(start, arr.length);
}
/** @internal */
export function concatenate(array1, array2) {
    if (array2 === undefined || array2.length === 0)
        return array1;
    if (array1 === undefined || array1.length === 0)
        return array2;
    return [...array1, ...array2];
}
function selectIndex(_, i) {
    return i;
}
/** @internal */
export function indicesOf(array) {
    return array.map(selectIndex);
}
function deduplicateRelational(array, equalityComparer, comparer) {
    // Perform a stable sort of the array. This ensures the first entry in a list of
    // duplicates remains the first entry in the result.
    const indices = indicesOf(array);
    stableSortIndices(array, indices, comparer);
    let last = array[indices[0]];
    const deduplicated = [indices[0]];
    for (let i = 1; i < indices.length; i++) {
        const index = indices[i];
        const item = array[index];
        if (!equalityComparer(last, item)) {
            deduplicated.push(index);
            last = item;
        }
    }
    // restore original order
    deduplicated.sort();
    return deduplicated.map(i => array[i]);
}
function deduplicateEquality(array, equalityComparer) {
    const result = [];
    for (let i = 0; i < array.length; i++) {
        pushIfUnique(result, array[i], equalityComparer);
    }
    return result;
}
/**
 * Deduplicates an unsorted array.
 * @param equalityComparer An `EqualityComparer` used to determine if two values are duplicates.
 * @param comparer An optional `Comparer` used to sort entries before comparison, though the
 * result will remain in the original order in `array`.
 *
 * @internal
 */
export function deduplicate(array, equalityComparer, comparer) {
    return array.length === 0 ? [] :
        array.length === 1 ? array.slice() :
            comparer ? deduplicateRelational(array, equalityComparer, comparer) :
                deduplicateEquality(array, equalityComparer);
}
/**
 * Deduplicates an array that has already been sorted.
 */
function deduplicateSorted(array, comparer) {
    if (array.length === 0)
        return emptyArray;
    let last = array[0];
    const deduplicated = [last];
    for (let i = 1; i < array.length; i++) {
        const next = array[i];
        switch (comparer(next, last)) {
            // equality comparison
            case true:
            // relational comparison
            // falls through
            case Comparison.EqualTo:
                continue;
            case Comparison.LessThan:
                // If `array` is sorted, `next` should **never** be less than `last`.
                return Debug.fail("Array is unsorted.");
        }
        deduplicated.push(last = next);
    }
    return deduplicated;
}
/** @internal */
export function createSortedArray() {
    return []; // TODO: GH#19873
}
/** @internal */
export function insertSorted(array, insert, compare, equalityComparer, allowDuplicates) {
    if (array.length === 0) {
        array.push(insert);
        return true;
    }
    const insertIndex = binarySearch(array, insert, identity, compare);
    if (insertIndex < 0) {
        if (equalityComparer && !allowDuplicates) {
            const idx = ~insertIndex;
            if (idx > 0 && equalityComparer(insert, array[idx - 1])) {
                return false;
            }
            if (idx < array.length && equalityComparer(insert, array[idx])) {
                array.splice(idx, 1, insert);
                return true;
            }
        }
        array.splice(~insertIndex, 0, insert);
        return true;
    }
    if (allowDuplicates) {
        array.splice(insertIndex, 0, insert);
        return true;
    }
    return false;
}
/** @internal */
export function sortAndDeduplicate(array, comparer, equalityComparer) {
    return deduplicateSorted(toSorted(array, comparer), equalityComparer ?? comparer ?? compareStringsCaseSensitive);
}
/** @internal */
export function arrayIsEqualTo(array1, array2, equalityComparer = equateValues) {
    if (array1 === undefined || array2 === undefined) {
        return array1 === array2;
    }
    if (array1.length !== array2.length) {
        return false;
    }
    for (let i = 0; i < array1.length; i++) {
        if (!equalityComparer(array1[i], array2[i], i)) {
            return false;
        }
    }
    return true;
}
/** @internal */
export function compact(array) {
    let result;
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            const v = array[i];
            // Either the result has been initialized (and is looking to collect truthy values separately),
            // or we've hit our first falsy value and need to copy over the current stretch of truthy values.
            if (result ?? !v) {
                result ??= array.slice(0, i);
                if (v) {
                    result.push(v);
                }
            }
        }
    }
    return result ?? array;
}
/**
 * Gets the relative complement of `arrayA` with respect to `arrayB`, returning the elements that
 * are not present in `arrayA` but are present in `arrayB`. Assumes both arrays are sorted
 * based on the provided comparer.
 *
 * @internal
 */
export function relativeComplement(arrayA, arrayB, comparer) {
    if (!arrayB || !arrayA || arrayB.length === 0 || arrayA.length === 0)
        return arrayB;
    const result = [];
    loopB: for (let offsetA = 0, offsetB = 0; offsetB < arrayB.length; offsetB++) {
        if (offsetB > 0) {
            // Ensure `arrayB` is properly sorted.
            Debug.assertGreaterThanOrEqual(comparer(arrayB[offsetB], arrayB[offsetB - 1]), Comparison.EqualTo);
        }
        loopA: for (const startA = offsetA; offsetA < arrayA.length; offsetA++) {
            if (offsetA > startA) {
                // Ensure `arrayA` is properly sorted. We only need to perform this check if
                // `offsetA` has changed since we entered the loop.
                Debug.assertGreaterThanOrEqual(comparer(arrayA[offsetA], arrayA[offsetA - 1]), Comparison.EqualTo);
            }
            switch (comparer(arrayB[offsetB], arrayA[offsetA])) {
                case Comparison.LessThan:
                    // If B is less than A, B does not exist in arrayA. Add B to the result and
                    // move to the next element in arrayB without changing the current position
                    // in arrayA.
                    result.push(arrayB[offsetB]);
                    continue loopB;
                case Comparison.EqualTo:
                    // If B is equal to A, B exists in arrayA. Move to the next element in
                    // arrayB without adding B to the result or changing the current position
                    // in arrayA.
                    continue loopB;
                case Comparison.GreaterThan:
                    // If B is greater than A, we need to keep looking for B in arrayA. Move to
                    // the next element in arrayA and recheck.
                    continue loopA;
            }
        }
    }
    return result;
}
export function append(to, value) {
    if (value === undefined)
        return to;
    if (to === undefined)
        return [value];
    to.push(value);
    return to;
}
/** @internal */
export function combine(xs, ys) {
    if (xs === undefined)
        return ys;
    if (ys === undefined)
        return xs;
    if (isArray(xs))
        return isArray(ys) ? concatenate(xs, ys) : append(xs, ys);
    if (isArray(ys))
        return append(ys, xs);
    return [xs, ys];
}
/**
 * Gets the actual offset into an array for a relative offset. Negative offsets indicate a
 * position offset from the end of the array.
 */
function toOffset(array, offset) {
    return offset < 0 ? array.length + offset : offset;
}
/** @internal */
export function addRange(to, from, start, end) {
    if (from === undefined || from.length === 0)
        return to;
    if (to === undefined)
        return from.slice(start, end);
    start = start === undefined ? 0 : toOffset(from, start);
    end = end === undefined ? from.length : toOffset(from, end);
    for (let i = start; i < end && i < from.length; i++) {
        if (from[i] !== undefined) {
            to.push(from[i]);
        }
    }
    return to;
}
/**
 * @return Whether the value was added.
 *
 * @internal
 */
export function pushIfUnique(array, toAdd, equalityComparer) {
    if (contains(array, toAdd, equalityComparer)) {
        return false;
    }
    else {
        array.push(toAdd);
        return true;
    }
}
/**
 * Unlike `pushIfUnique`, this can take `undefined` as an input, and returns a new array.
 *
 * @internal
 */
export function appendIfUnique(array, toAdd, equalityComparer) {
    if (array !== undefined) {
        pushIfUnique(array, toAdd, equalityComparer);
        return array;
    }
    else {
        return [toAdd];
    }
}
function stableSortIndices(array, indices, comparer) {
    // sort indices by value then position
    indices.sort((x, y) => comparer(array[x], array[y]) || compareValues(x, y));
}
/**
 * Returns a new sorted array. This sort is stable, meaning elements equal to each other maintain their relative position in the array.
 *
 * @internal
 */
export function toSorted(array, comparer) {
    return (array.length === 0 ? emptyArray : array.slice().sort(comparer));
}
/** @internal */
export function* arrayReverseIterator(array) {
    for (let i = array.length - 1; i >= 0; i--) {
        yield array[i];
    }
}
/** @internal */
export function rangeEquals(array1, array2, pos, end) {
    while (pos < end) {
        if (array1[pos] !== array2[pos]) {
            return false;
        }
        pos++;
    }
    return true;
}
/**
 * Returns the element at a specific offset in an array if non-empty, `undefined` otherwise.
 * A negative offset indicates the element should be retrieved from the end of the array.
 *
 * @internal
 */
export const elementAt = !!Array.prototype.at
    ? (array, offset) => array?.at(offset)
    : (array, offset) => {
        if (array !== undefined) {
            offset = toOffset(array, offset);
            if (offset < array.length) {
                return array[offset];
            }
        }
        return undefined;
    };
/**
 * Returns the first element of an array if non-empty, `undefined` otherwise.
 *
 * @internal
 */
export function firstOrUndefined(array) {
    return array === undefined || array.length === 0 ? undefined : array[0];
}
/** @internal */
export function firstOrUndefinedIterator(iter) {
    if (iter !== undefined) {
        for (const value of iter) {
            return value;
        }
    }
    return undefined;
}
/** @internal */
export function first(array) {
    Debug.assert(array.length !== 0);
    return array[0];
}
/** @internal */
export function firstIterator(iter) {
    for (const value of iter) {
        return value;
    }
    Debug.fail("iterator is empty");
}
/**
 * Returns the last element of an array if non-empty, `undefined` otherwise.
 *
 * @internal
 */
export function lastOrUndefined(array) {
    return array === undefined || array.length === 0 ? undefined : array[array.length - 1];
}
/** @internal */
export function last(array) {
    Debug.assert(array.length !== 0);
    return array[array.length - 1];
}
/**
 * Returns the only element of an array if it contains only one element, `undefined` otherwise.
 *
 * @internal
 */
export function singleOrUndefined(array) {
    return array !== undefined && array.length === 1
        ? array[0]
        : undefined;
}
/**
 * Returns the only element of an array if it contains only one element; throws otherwise.
 *
 * @internal
 */
export function single(array) {
    return Debug.checkDefined(singleOrUndefined(array));
}
/** @internal */
export function singleOrMany(array) {
    return array !== undefined && array.length === 1
        ? array[0]
        : array;
}
/** @internal */
export function replaceElement(array, index, value) {
    const result = array.slice(0);
    result[index] = value;
    return result;
}
/**
 * Performs a binary search, finding the index at which `value` occurs in `array`.
 * If no such index is found, returns the 2's-complement of first index at which
 * `array[index]` exceeds `value`.
 * @param array A sorted array whose first element must be no larger than number
 * @param value The value to be searched for in the array.
 * @param keySelector A callback used to select the search key from `value` and each element of
 * `array`.
 * @param keyComparer A callback used to compare two keys in a sorted array.
 * @param offset An offset into `array` at which to start the search.
 *
 * @internal
 */
export function binarySearch(array, value, keySelector, keyComparer, offset) {
    return binarySearchKey(array, keySelector(value), keySelector, keyComparer, offset);
}
/**
 * Performs a binary search, finding the index at which an object with `key` occurs in `array`.
 * If no such index is found, returns the 2's-complement of first index at which
 * `array[index]` exceeds `key`.
 * @param array A sorted array whose first element must be no larger than number
 * @param key The key to be searched for in the array.
 * @param keySelector A callback used to select the search key from each element of `array`.
 * @param keyComparer A callback used to compare two keys in a sorted array.
 * @param offset An offset into `array` at which to start the search.
 *
 * @internal
 */
export function binarySearchKey(array, key, keySelector, keyComparer, offset) {
    if (!some(array)) {
        return -1;
    }
    let low = offset ?? 0;
    let high = array.length - 1;
    while (low <= high) {
        const middle = low + ((high - low) >> 1);
        const midKey = keySelector(array[middle], middle);
        switch (Math.sign(keyComparer(midKey, key))) {
            case Comparison.LessThan:
                low = middle + 1;
                break;
            case Comparison.EqualTo:
                return middle;
            case Comparison.GreaterThan:
                high = middle - 1;
                break;
        }
    }
    return ~low;
}
/** @internal */
export function reduceLeft(array, f, initial, start, count) {
    if (array && array.length > 0) {
        const size = array.length;
        if (size > 0) {
            let pos = start === undefined || start < 0 ? 0 : start;
            const end = count === undefined || pos + count > size - 1 ? size - 1 : pos + count;
            let result;
            if (arguments.length <= 2) {
                result = array[pos];
                pos++;
            }
            else {
                result = initial;
            }
            while (pos <= end) {
                result = f(result, array[pos], pos);
                pos++;
            }
            return result;
        }
    }
    return initial;
}
const hasOwnProperty = Object.prototype.hasOwnProperty;
/**
 * Indicates whether a map-like contains an own property with the specified key.
 *
 * @param map A map-like.
 * @param key A property key.
 *
 * @internal
 */
export function hasProperty(map, key) {
    return hasOwnProperty.call(map, key);
}
/**
 * Gets the value of an owned property in a map-like.
 *
 * @param map A map-like.
 * @param key A property key.
 *
 * @internal
 */
export function getProperty(map, key) {
    return hasOwnProperty.call(map, key) ? map[key] : undefined;
}
/**
 * Gets the owned, enumerable property keys of a map-like.
 *
 * @internal
 */
export function getOwnKeys(map) {
    const keys = [];
    for (const key in map) {
        if (hasOwnProperty.call(map, key)) {
            keys.push(key);
        }
    }
    return keys;
}
/** @internal */
export function getAllKeys(obj) {
    const result = [];
    do {
        const names = Object.getOwnPropertyNames(obj);
        for (const name of names) {
            pushIfUnique(result, name);
        }
    } while (obj = Object.getPrototypeOf(obj));
    return result;
}
/** @internal */
export function getOwnValues(collection) {
    const values = [];
    for (const key in collection) {
        if (hasOwnProperty.call(collection, key)) {
            values.push(collection[key]);
        }
    }
    return values;
}
/** @internal */
export function arrayOf(count, f) {
    const result = new Array(count);
    for (let i = 0; i < count; i++) {
        result[i] = f(i);
    }
    return result;
}
/** @internal */
export function arrayFrom(iterator, map) {
    const result = [];
    for (const value of iterator) {
        result.push(map ? map(value) : value);
    }
    return result;
}
/** @internal */
export function assign(t, ...args) {
    for (const arg of args) {
        if (arg === undefined)
            continue;
        for (const p in arg) {
            if (hasProperty(arg, p)) {
                t[p] = arg[p];
            }
        }
    }
    return t;
}
/**
 * Performs a shallow equality comparison of the contents of two map-likes.
 *
 * @param left A map-like whose properties should be compared.
 * @param right A map-like whose properties should be compared.
 *
 * @internal
 */
export function equalOwnProperties(left, right, equalityComparer = equateValues) {
    if (left === right)
        return true;
    if (!left || !right)
        return false;
    for (const key in left) {
        if (hasOwnProperty.call(left, key)) {
            if (!hasOwnProperty.call(right, key))
                return false;
            if (!equalityComparer(left[key], right[key]))
                return false;
        }
    }
    for (const key in right) {
        if (hasOwnProperty.call(right, key)) {
            if (!hasOwnProperty.call(left, key))
                return false;
        }
    }
    return true;
}
/** @internal */
export function arrayToMap(array, makeKey, makeValue = identity) {
    const result = new Map();
    for (let i = 0; i < array.length; i++) {
        const value = array[i];
        const key = makeKey(value);
        if (key !== undefined)
            result.set(key, makeValue(value));
    }
    return result;
}
/** @internal */
export function arrayToNumericMap(array, makeKey, makeValue = identity) {
    const result = [];
    for (let i = 0; i < array.length; i++) {
        const value = array[i];
        result[makeKey(value)] = makeValue(value);
    }
    return result;
}
/** @internal */
export function arrayToMultiMap(values, makeKey, makeValue = identity) {
    const result = createMultiMap();
    for (let i = 0; i < values.length; i++) {
        const value = values[i];
        result.add(makeKey(value), makeValue(value));
    }
    return result;
}
/** @internal */
export function group(values, getGroupId, resultSelector = identity) {
    return arrayFrom(arrayToMultiMap(values, getGroupId).values(), resultSelector);
}
export function groupBy(values, keySelector) {
    const result = {};
    if (values !== undefined) {
        for (let i = 0; i < values.length; i++) {
            const value = values[i];
            const key = `${keySelector(value)}`;
            const array = result[key] ??= [];
            array.push(value);
        }
    }
    return result;
}
/** @internal */
export function clone(object) {
    const result = {};
    for (const id in object) {
        if (hasOwnProperty.call(object, id)) {
            result[id] = object[id];
        }
    }
    return result;
}
/**
 * Creates a new object by adding the own properties of `second`, then the own properties of `first`.
 *
 * NOTE: This means that if a property exists in both `first` and `second`, the property in `first` will be chosen.
 *
 * @internal
 */
export function extend(first, second) {
    const result = {};
    for (const id in second) {
        if (hasOwnProperty.call(second, id)) {
            result[id] = second[id];
        }
    }
    for (const id in first) {
        if (hasOwnProperty.call(first, id)) {
            result[id] = first[id];
        }
    }
    return result;
}
/** @internal */
export function copyProperties(first, second) {
    for (const id in second) {
        if (hasOwnProperty.call(second, id)) {
            first[id] = second[id];
        }
    }
}
/** @internal */
export function maybeBind(obj, fn) {
    return fn?.bind(obj);
}
/** @internal */
export function createMultiMap() {
    const map = new Map();
    map.add = multiMapAdd;
    map.remove = multiMapRemove;
    return map;
}
function multiMapAdd(key, value) {
    let values = this.get(key);
    if (values !== undefined) {
        values.push(value);
    }
    else {
        this.set(key, values = [value]);
    }
    return values;
}
function multiMapRemove(key, value) {
    const values = this.get(key);
    if (values !== undefined) {
        unorderedRemoveItem(values, value);
        if (!values.length) {
            this.delete(key);
        }
    }
}
/** @internal */
export function createQueue(items) {
    const elements = items?.slice() ?? [];
    let headIndex = 0;
    function isEmpty() {
        return headIndex === elements.length;
    }
    function enqueue(...items) {
        elements.push(...items);
    }
    function dequeue() {
        if (isEmpty()) {
            throw new Error("Queue is empty");
        }
        const result = elements[headIndex];
        elements[headIndex] = undefined; // Don't keep referencing dequeued item
        headIndex++;
        // If more than half of the queue is empty, copy the remaining elements to the
        // front and shrink the array (unless we'd be saving fewer than 100 slots)
        if (headIndex > 100 && headIndex > (elements.length >> 1)) {
            const newLength = elements.length - headIndex;
            elements.copyWithin(/*target*/ 0, /*start*/ headIndex);
            elements.length = newLength;
            headIndex = 0;
        }
        return result;
    }
    return {
        enqueue,
        dequeue,
        isEmpty,
    };
}
/**
 * Creates a Set with custom equality and hash code functionality.  This is useful when you
 * want to use something looser than object identity - e.g. "has the same span".
 *
 * If `equals(a, b)`, it must be the case that `getHashCode(a) === getHashCode(b)`.
 * The converse is not required.
 *
 * To facilitate a perf optimization (lazy allocation of bucket arrays), `TElement` is
 * assumed not to be an array type.
 *
 * @internal
 */
export function createSet(getHashCode, equals) {
    const multiMap = new Map();
    let size = 0;
    function* getElementIterator() {
        for (const value of multiMap.values()) {
            if (isArray(value)) {
                yield* value;
            }
            else {
                yield value;
            }
        }
    }
    const set = {
        has(element) {
            const hash = getHashCode(element);
            if (!multiMap.has(hash))
                return false;
            const candidates = multiMap.get(hash);
            if (isArray(candidates))
                return contains(candidates, element, equals);
            return equals(candidates, element);
        },
        add(element) {
            const hash = getHashCode(element);
            if (multiMap.has(hash)) {
                const values = multiMap.get(hash);
                if (isArray(values)) {
                    if (!contains(values, element, equals)) {
                        values.push(element);
                        size++;
                    }
                }
                else {
                    const value = values;
                    if (!equals(value, element)) {
                        multiMap.set(hash, [value, element]);
                        size++;
                    }
                }
            }
            else {
                multiMap.set(hash, element);
                size++;
            }
            return this;
        },
        delete(element) {
            const hash = getHashCode(element);
            if (!multiMap.has(hash))
                return false;
            const candidates = multiMap.get(hash);
            if (isArray(candidates)) {
                for (let i = 0; i < candidates.length; i++) {
                    if (equals(candidates[i], element)) {
                        if (candidates.length === 1) {
                            multiMap.delete(hash);
                        }
                        else if (candidates.length === 2) {
                            multiMap.set(hash, candidates[1 - i]);
                        }
                        else {
                            unorderedRemoveItemAt(candidates, i);
                        }
                        size--;
                        return true;
                    }
                }
            }
            else {
                const candidate = candidates;
                if (equals(candidate, element)) {
                    multiMap.delete(hash);
                    size--;
                    return true;
                }
            }
            return false;
        },
        clear() {
            multiMap.clear();
            size = 0;
        },
        get size() {
            return size;
        },
        forEach(action) {
            // NOTE: arrayFrom means that if the callback mutates the underlying collection,
            //       we won't have an accurate set of values
            for (const elements of arrayFrom(multiMap.values())) {
                if (isArray(elements)) {
                    for (const element of elements) {
                        action(element, element, set);
                    }
                }
                else {
                    const element = elements;
                    action(element, element, set);
                }
            }
        },
        keys() {
            return getElementIterator();
        },
        values() {
            return getElementIterator();
        },
        *entries() {
            for (const value of getElementIterator()) {
                yield [value, value];
            }
        },
        [Symbol.iterator]: () => {
            return getElementIterator();
        },
        [Symbol.toStringTag]: multiMap[Symbol.toStringTag],
    };
    return set;
}
/**
 * Tests whether a value is an array.
 *
 * @internal
 */
export function isArray(value) {
    // See: https://github.com/microsoft/TypeScript/issues/17002
    return Array.isArray(value);
}
/** @internal */
export function toArray(value) {
    return isArray(value) ? value : [value];
}
/**
 * Tests whether a value is string
 *
 * @internal
 */
export function isString(text) {
    return typeof text === "string";
}
/** @internal */
export function isNumber(x) {
    return typeof x === "number";
}
/** @internal */
export function tryCast(value, test) {
    return value !== undefined && test(value) ? value : undefined;
}
/** @internal */
export function cast(value, test) {
    if (value !== undefined && test(value))
        return value;
    return Debug.fail(`Invalid cast. The supplied value ${value} did not pass the test '${Debug.getFunctionName(test)}'.`);
}
/**
 * Does nothing.
 *
 * @internal
 */
export function noop(_) { }
/**
 * Do nothing and return false
 *
 * @internal
 */
export function returnFalse() {
    return false;
}
/**
 * Do nothing and return true
 *
 * @internal
 */
export function returnTrue() {
    return true;
}
/**
 * Do nothing and return undefined
 *
 * @internal
 */
export function returnUndefined() {
    return undefined;
}
/**
 * Returns its argument.
 *
 * @internal
 */
export function identity(x) {
    return x;
}
/**
 * Returns lower case string
 */
function toLowerCase(x) {
    return x.toLowerCase();
}
// We convert the file names to lower case as key for file name on case insensitive file system
// While doing so we need to handle special characters (eg \u0130) to ensure that we dont convert
// it to lower case, fileName with its lowercase form can exist along side it.
// Handle special characters and make those case sensitive instead
//
// |-#--|-Unicode--|-Char code-|-Desc-------------------------------------------------------------------|
// | 1. | i        | 105       | Ascii i                                                                |
// | 2. | I        | 73        | Ascii I                                                                |
// |-------- Special characters ------------------------------------------------------------------------|
// | 3. | \u0130   | 304       | Upper case I with dot above                                            |
// | 4. | i,\u0307 | 105,775   | i, followed by 775: Lower case of (3rd item)                           |
// | 5. | I,\u0307 | 73,775    | I, followed by 775: Upper case of (4th item), lower case is (4th item) |
// | 6. | \u0131   | 305       | Lower case i without dot, upper case is I (2nd item)                   |
// | 7. | \u00DF   | 223       | Lower case sharp s                                                     |
//
// Because item 3 is special where in its lowercase character has its own
// upper case form we cant convert its case.
// Rest special characters are either already in lower case format or
// they have corresponding upper case character so they dont need special handling
//
// But to avoid having to do string building for most common cases, also ignore
// a-z, 0-9, \u0131, \u00DF, \, /, ., : and space
const fileNameLowerCaseRegExp = /[^\u0130\u0131\u00DFa-z0-9\\/:\-_. ]+/g;
/**
 * Case insensitive file systems have descripencies in how they handle some characters (eg. turkish Upper case I with dot on top - \u0130)
 * This function is used in places where we want to make file name as a key on these systems
 * It is possible on mac to be able to refer to file name with I with dot on top as a fileName with its lower case form
 * But on windows we cannot. Windows can have fileName with I with dot on top next to its lower case and they can not each be referred with the lowercase forms
 * Technically we would want this function to be platform sepcific as well but
 * our api has till now only taken caseSensitive as the only input and just for some characters we dont want to update API and ensure all customers use those api
 * We could use upper case and we would still need to deal with the descripencies but
 * we want to continue using lower case since in most cases filenames are lowercasewe and wont need any case changes and avoid having to store another string for the key
 * So for this function purpose, we go ahead and assume character I with dot on top it as case sensitive since its very unlikely to use lower case form of that special character
 *
 * @internal
 */
export function toFileNameLowerCase(x) {
    return fileNameLowerCaseRegExp.test(x) ?
        x.replace(fileNameLowerCaseRegExp, toLowerCase) :
        x;
}
/**
 * Throws an error because a function is not implemented.
 *
 * @internal
 */
export function notImplemented() {
    throw new Error("Not implemented");
}
/** @internal */
export function memoize(callback) {
    let value;
    return () => {
        if (callback) {
            value = callback();
            callback = undefined;
        }
        return value;
    };
}
/**
 * A version of `memoize` that supports a single primitive argument
 *
 * @internal
 */
export function memoizeOne(callback) {
    const map = new Map();
    return (arg) => {
        const key = `${typeof arg}:${arg}`;
        let value = map.get(key);
        if (value === undefined && !map.has(key)) {
            value = callback(arg);
            map.set(key, value);
        }
        return value;
    };
}
/** @internal */
export var AssertionLevel;
(function (AssertionLevel) {
    AssertionLevel[AssertionLevel["None"] = 0] = "None";
    AssertionLevel[AssertionLevel["Normal"] = 1] = "Normal";
    AssertionLevel[AssertionLevel["Aggressive"] = 2] = "Aggressive";
    AssertionLevel[AssertionLevel["VeryAggressive"] = 3] = "VeryAggressive";
})(AssertionLevel || (AssertionLevel = {}));
/** @internal */
export function equateValues(a, b) {
    return a === b;
}
/**
 * Compare the equality of two strings using a case-sensitive ordinal comparison.
 *
 * Case-sensitive comparisons compare both strings one code-point at a time using the integer
 * value of each code-point after applying `toUpperCase` to each string. We always map both
 * strings to their upper-case form as some unicode characters do not properly round-trip to
 * lowercase (such as `ẞ` (German sharp capital s)).
 *
 * @internal
 */
export function equateStringsCaseInsensitive(a, b) {
    return a === b
        || a !== undefined
            && b !== undefined
            && a.toUpperCase() === b.toUpperCase();
}
/**
 * Compare the equality of two strings using a case-sensitive ordinal comparison.
 *
 * Case-sensitive comparisons compare both strings one code-point at a time using the
 * integer value of each code-point.
 *
 * @internal
 */
export function equateStringsCaseSensitive(a, b) {
    return equateValues(a, b);
}
export function compareComparableValues(a, b) {
    return a === b ? Comparison.EqualTo :
        a === undefined ? Comparison.LessThan :
            b === undefined ? Comparison.GreaterThan :
                a < b ? Comparison.LessThan :
                    Comparison.GreaterThan;
}
/**
 * Compare two numeric values for their order relative to each other.
 * To compare strings, use any of the `compareStrings` functions.
 *
 * @internal
 */
export function compareValues(a, b) {
    return compareComparableValues(a, b);
}
/**
 * Compare two TextSpans, first by `start`, then by `length`.
 *
 * @internal
 */
export function compareTextSpans(a, b) {
    return compareValues(a?.start, b?.start) || compareValues(a?.length, b?.length);
}
/** @internal */
export function maxBy(arr, init, mapper) {
    for (let i = 0; i < arr.length; i++) {
        init = Math.max(init, mapper(arr[i]));
    }
    return init;
}
/** @internal */
export function min(items, compare) {
    return reduceLeft(items, (x, y) => compare(x, y) === Comparison.LessThan ? x : y);
}
/**
 * Compare two strings using a case-insensitive ordinal comparison.
 *
 * Ordinal comparisons are based on the difference between the unicode code points of both
 * strings. Characters with multiple unicode representations are considered unequal. Ordinal
 * comparisons provide predictable ordering, but place "a" after "B".
 *
 * Case-insensitive comparisons compare both strings one code-point at a time using the integer
 * value of each code-point after applying `toUpperCase` to each string. We always map both
 * strings to their upper-case form as some unicode characters do not properly round-trip to
 * lowercase (such as `ẞ` (German sharp capital s)).
 *
 * @internal
 */
export function compareStringsCaseInsensitive(a, b) {
    if (a === b)
        return Comparison.EqualTo;
    if (a === undefined)
        return Comparison.LessThan;
    if (b === undefined)
        return Comparison.GreaterThan;
    a = a.toUpperCase();
    b = b.toUpperCase();
    return a < b ? Comparison.LessThan : a > b ? Comparison.GreaterThan : Comparison.EqualTo;
}
/**
 * `compareStringsCaseInsensitive` transforms letters to uppercase for unicode reasons,
 * while eslint's `sort-imports` rule transforms letters to lowercase. Which one you choose
 * affects the relative order of letters and ASCII characters 91-96, of which `_` is a
 * valid character in an identifier. So if we used `compareStringsCaseInsensitive` for
 * import sorting, TypeScript and eslint would disagree about the correct case-insensitive
 * sort order for `__String` and `Foo`. Since eslint's whole job is to create consistency
 * by enforcing nitpicky details like this, it makes way more sense for us to just adopt
 * their convention so users can have auto-imports without making eslint angry.
 *
 * @internal
 */
export function compareStringsCaseInsensitiveEslintCompatible(a, b) {
    if (a === b)
        return Comparison.EqualTo;
    if (a === undefined)
        return Comparison.LessThan;
    if (b === undefined)
        return Comparison.GreaterThan;
    a = a.toLowerCase();
    b = b.toLowerCase();
    return a < b ? Comparison.LessThan : a > b ? Comparison.GreaterThan : Comparison.EqualTo;
}
/**
 * Compare two strings using a case-sensitive ordinal comparison.
 *
 * Ordinal comparisons are based on the difference between the unicode code points of both
 * strings. Characters with multiple unicode representations are considered unequal. Ordinal
 * comparisons provide predictable ordering, but place "a" after "B".
 *
 * Case-sensitive comparisons compare both strings one code-point at a time using the integer
 * value of each code-point.
 *
 * @internal
 */
export function compareStringsCaseSensitive(a, b) {
    return compareComparableValues(a, b);
}
/** @internal */
export function getStringComparer(ignoreCase) {
    return ignoreCase ? compareStringsCaseInsensitive : compareStringsCaseSensitive;
}
/**
 * Creates a string comparer for use with string collation in the UI.
 */
const createUIStringComparer = (() => {
    return createIntlCollatorStringComparer;
    function compareWithCallback(a, b, comparer) {
        if (a === b)
            return Comparison.EqualTo;
        if (a === undefined)
            return Comparison.LessThan;
        if (b === undefined)
            return Comparison.GreaterThan;
        const value = comparer(a, b);
        return value < 0 ? Comparison.LessThan : value > 0 ? Comparison.GreaterThan : Comparison.EqualTo;
    }
    function createIntlCollatorStringComparer(locale) {
        // Intl.Collator.prototype.compare is bound to the collator. See NOTE in
        // http://www.ecma-international.org/ecma-402/2.0/#sec-Intl.Collator.prototype.compare
        const comparer = new Intl.Collator(locale, { usage: "sort", sensitivity: "variant", numeric: true }).compare;
        return (a, b) => compareWithCallback(a, b, comparer);
    }
})();
let uiComparerCaseSensitive;
let uiLocale;
/** @internal */
export function getUILocale() {
    return uiLocale;
}
/** @internal */
export function setUILocale(value) {
    if (uiLocale !== value) {
        uiLocale = value;
        uiComparerCaseSensitive = undefined;
    }
}
/**
 * Compare two strings in a using the case-sensitive sort behavior of the UI locale.
 *
 * Ordering is not predictable between different host locales, but is best for displaying
 * ordered data for UI presentation. Characters with multiple unicode representations may
 * be considered equal.
 *
 * Case-sensitive comparisons compare strings that differ in base characters, or
 * accents/diacritic marks, or case as unequal.
 *
 * @internal
 */
export function compareStringsCaseSensitiveUI(a, b) {
    uiComparerCaseSensitive ??= createUIStringComparer(uiLocale);
    return uiComparerCaseSensitive(a, b);
}
/** @internal */
export function compareProperties(a, b, key, comparer) {
    return a === b ? Comparison.EqualTo :
        a === undefined ? Comparison.LessThan :
            b === undefined ? Comparison.GreaterThan :
                comparer(a[key], b[key]);
}
/**
 * True is greater than false.
 *
 * @internal
 */
export function compareBooleans(a, b) {
    return compareValues(a ? 1 : 0, b ? 1 : 0);
}
/**
 * Given a name and a list of names that are *not* equal to the name, return a spelling suggestion if there is one that is close enough.
 * Names less than length 3 only check for case-insensitive equality.
 *
 * find the candidate with the smallest Levenshtein distance,
 *    except for candidates:
 *      * With no name
 *      * Whose length differs from the target name by more than 0.34 of the length of the name.
 *      * Whose levenshtein distance is more than 0.4 of the length of the name
 *        (0.4 allows 1 substitution/transposition for every 5 characters,
 *         and 1 insertion/deletion at 3 characters)
 *
 * @internal
 */
export function getSpellingSuggestion(name, candidates, getName) {
    const maximumLengthDifference = Math.max(2, Math.floor(name.length * 0.34));
    let bestDistance = Math.floor(name.length * 0.4) + 1; // If the best result is worse than this, don't bother.
    let bestCandidate;
    for (const candidate of candidates) {
        const candidateName = getName(candidate);
        if (candidateName !== undefined && Math.abs(candidateName.length - name.length) <= maximumLengthDifference) {
            if (candidateName === name) {
                continue;
            }
            // Only consider candidates less than 3 characters long when they differ by case.
            // Otherwise, don't bother, since a user would usually notice differences of a 2-character name.
            if (candidateName.length < 3 && candidateName.toLowerCase() !== name.toLowerCase()) {
                continue;
            }
            const distance = levenshteinWithMax(name, candidateName, bestDistance - 0.1);
            if (distance === undefined) {
                continue;
            }
            Debug.assert(distance < bestDistance); // Else `levenshteinWithMax` should return undefined
            bestDistance = distance;
            bestCandidate = candidate;
        }
    }
    return bestCandidate;
}
function levenshteinWithMax(s1, s2, max) {
    let previous = new Array(s2.length + 1);
    let current = new Array(s2.length + 1);
    /** Represents any value > max. We don't care about the particular value. */
    const big = max + 0.01;
    for (let i = 0; i <= s2.length; i++) {
        previous[i] = i;
    }
    for (let i = 1; i <= s1.length; i++) {
        const c1 = s1.charCodeAt(i - 1);
        const minJ = Math.ceil(i > max ? i - max : 1);
        const maxJ = Math.floor(s2.length > max + i ? max + i : s2.length);
        current[0] = i;
        /** Smallest value of the matrix in the ith column. */
        let colMin = i;
        for (let j = 1; j < minJ; j++) {
            current[j] = big;
        }
        for (let j = minJ; j <= maxJ; j++) {
            // case difference should be significantly cheaper than other differences
            const substitutionDistance = s1[i - 1].toLowerCase() === s2[j - 1].toLowerCase()
                ? (previous[j - 1] + 0.1)
                : (previous[j - 1] + 2);
            const dist = c1 === s2.charCodeAt(j - 1)
                ? previous[j - 1]
                : Math.min(/*delete*/ previous[j] + 1, /*insert*/ current[j - 1] + 1, /*substitute*/ substitutionDistance);
            current[j] = dist;
            colMin = Math.min(colMin, dist);
        }
        for (let j = maxJ + 1; j <= s2.length; j++) {
            current[j] = big;
        }
        if (colMin > max) {
            // Give up -- everything in this column is > max and it can't get better in future columns.
            return undefined;
        }
        const temp = previous;
        previous = current;
        current = temp;
    }
    const res = previous[s2.length];
    return res > max ? undefined : res;
}
/** @internal */
export function endsWith(str, suffix, ignoreCase) {
    const expectedPos = str.length - suffix.length;
    return expectedPos >= 0 && (ignoreCase
        ? equateStringsCaseInsensitive(str.slice(expectedPos), suffix)
        : str.indexOf(suffix, expectedPos) === expectedPos);
}
/** @internal */
export function removeSuffix(str, suffix) {
    return endsWith(str, suffix) ? str.slice(0, str.length - suffix.length) : str;
}
/** @internal */
export function tryRemoveSuffix(str, suffix) {
    return endsWith(str, suffix) ? str.slice(0, str.length - suffix.length) : undefined;
}
/**
 * Takes a string like "jquery-min.4.2.3" and returns "jquery"
 *
 * @internal
 */
export function removeMinAndVersionNumbers(fileName) {
    // We used to use the regex /[.-]((min)|(\d+(\.\d+)*))$/ and would just .replace it twice.
    // Unfortunately, that regex has O(n^2) performance because v8 doesn't match from the end of the string.
    // Instead, we now essentially scan the filename (backwards) ourselves.
    let end = fileName.length;
    for (let pos = end - 1; pos > 0; pos--) {
        let ch = fileName.charCodeAt(pos);
        if (ch >= CharacterCodes._0 && ch <= CharacterCodes._9) {
            // Match a \d+ segment
            do {
                --pos;
                ch = fileName.charCodeAt(pos);
            } while (pos > 0 && ch >= CharacterCodes._0 && ch <= CharacterCodes._9);
        }
        else if (pos > 4 && (ch === CharacterCodes.n || ch === CharacterCodes.N)) {
            // Looking for "min" or "min"
            // Already matched the 'n'
            --pos;
            ch = fileName.charCodeAt(pos);
            if (ch !== CharacterCodes.i && ch !== CharacterCodes.I) {
                break;
            }
            --pos;
            ch = fileName.charCodeAt(pos);
            if (ch !== CharacterCodes.m && ch !== CharacterCodes.M) {
                break;
            }
            --pos;
            ch = fileName.charCodeAt(pos);
        }
        else {
            // This character is not part of either suffix pattern
            break;
        }
        if (ch !== CharacterCodes.minus && ch !== CharacterCodes.dot) {
            break;
        }
        end = pos;
    }
    // end might be fileName.length, in which case this should internally no-op
    return end === fileName.length ? fileName : fileName.slice(0, end);
}
/**
 * Remove an item from an array, moving everything to its right one space left.
 *
 * @internal
 */
export function orderedRemoveItem(array, item) {
    for (let i = 0; i < array.length; i++) {
        if (array[i] === item) {
            orderedRemoveItemAt(array, i);
            return true;
        }
    }
    return false;
}
/**
 * Remove an item by index from an array, moving everything to its right one space left.
 *
 * @internal
 */
export function orderedRemoveItemAt(array, index) {
    // This seems to be faster than either `array.splice(i, 1)` or `array.copyWithin(i, i+ 1)`.
    for (let i = index; i < array.length - 1; i++) {
        array[i] = array[i + 1];
    }
    array.pop();
}
function unorderedRemoveItemAt(array, index) {
    // Fill in the "hole" left at `index`.
    array[index] = array[array.length - 1];
    array.pop();
}
/**
 * Remove the *first* occurrence of `item` from the array.
 *
 * @internal
 */
export function unorderedRemoveItem(array, item) {
    return unorderedRemoveFirstItemWhere(array, element => element === item);
}
/** Remove the *first* element satisfying `predicate`. */
function unorderedRemoveFirstItemWhere(array, predicate) {
    for (let i = 0; i < array.length; i++) {
        if (predicate(array[i])) {
            unorderedRemoveItemAt(array, i);
            return true;
        }
    }
    return false;
}
/** @internal */
export function createGetCanonicalFileName(useCaseSensitiveFileNames) {
    return useCaseSensitiveFileNames ? identity : toFileNameLowerCase;
}
/** @internal */
export function patternText({ prefix, suffix }) {
    return `${prefix}*${suffix}`;
}
/**
 * Given that candidate matches pattern, returns the text matching the '*'.
 * E.g.: matchedText(tryParsePattern("foo*baz"), "foobarbaz") === "bar"
 *
 * @internal
 */
export function matchedText(pattern, candidate) {
    Debug.assert(isPatternMatch(pattern, candidate));
    return candidate.substring(pattern.prefix.length, candidate.length - pattern.suffix.length);
}
/**
 * Return the object corresponding to the best pattern to match `candidate`.
 *
 * @internal
 */
export function findBestPatternMatch(values, getPattern, candidate) {
    let matchedValue;
    // use length of prefix as betterness criteria
    let longestMatchPrefixLength = -1;
    for (let i = 0; i < values.length; i++) {
        const v = values[i];
        const pattern = getPattern(v);
        if (pattern.prefix.length > longestMatchPrefixLength && isPatternMatch(pattern, candidate)) {
            longestMatchPrefixLength = pattern.prefix.length;
            matchedValue = v;
        }
    }
    return matchedValue;
}
/** @internal */
export function startsWith(str, prefix, ignoreCase) {
    return ignoreCase
        ? equateStringsCaseInsensitive(str.slice(0, prefix.length), prefix)
        : str.lastIndexOf(prefix, 0) === 0;
}
/** @internal */
export function removePrefix(str, prefix) {
    return startsWith(str, prefix) ? str.substr(prefix.length) : str;
}
/** @internal */
export function tryRemovePrefix(str, prefix, getCanonicalFileName = identity) {
    return startsWith(getCanonicalFileName(str), getCanonicalFileName(prefix)) ? str.substring(prefix.length) : undefined;
}
/** @internal */
export function isPatternMatch({ prefix, suffix }, candidate) {
    return candidate.length >= prefix.length + suffix.length &&
        startsWith(candidate, prefix) &&
        endsWith(candidate, suffix);
}
/** @internal */
export function and(f, g) {
    return (arg) => f(arg) && g(arg);
}
/** @internal */
export function or(...fs) {
    return (...args) => {
        let lastResult;
        for (const f of fs) {
            lastResult = f(...args);
            if (lastResult) {
                return lastResult;
            }
        }
        return lastResult;
    };
}
/** @internal */
export function not(fn) {
    return (...args) => !fn(...args);
}
/** @internal */
export function assertType(_) { }
/** @internal */
export function singleElementArray(t) {
    return t === undefined ? undefined : [t];
}
/** @internal */
export function enumerateInsertsAndDeletes(newItems, oldItems, comparer, inserted, deleted, unchanged) {
    unchanged ??= noop;
    let newIndex = 0;
    let oldIndex = 0;
    const newLen = newItems.length;
    const oldLen = oldItems.length;
    let hasChanges = false;
    while (newIndex < newLen && oldIndex < oldLen) {
        const newItem = newItems[newIndex];
        const oldItem = oldItems[oldIndex];
        const compareResult = comparer(newItem, oldItem);
        if (compareResult === Comparison.LessThan) {
            inserted(newItem);
            newIndex++;
            hasChanges = true;
        }
        else if (compareResult === Comparison.GreaterThan) {
            deleted(oldItem);
            oldIndex++;
            hasChanges = true;
        }
        else {
            unchanged(oldItem, newItem);
            newIndex++;
            oldIndex++;
        }
    }
    while (newIndex < newLen) {
        inserted(newItems[newIndex++]);
        hasChanges = true;
    }
    while (oldIndex < oldLen) {
        deleted(oldItems[oldIndex++]);
        hasChanges = true;
    }
    return hasChanges;
}
/** @internal */
export function cartesianProduct(arrays) {
    const result = [];
    cartesianProductWorker(arrays, result, /*outer*/ undefined, 0);
    return result;
}
function cartesianProductWorker(arrays, result, outer, index) {
    for (const element of arrays[index]) {
        let inner;
        if (outer) {
            inner = outer.slice();
            inner.push(element);
        }
        else {
            inner = [element];
        }
        if (index === arrays.length - 1) {
            result.push(inner);
        }
        else {
            cartesianProductWorker(arrays, result, inner, index + 1);
        }
    }
}
export function takeWhile(array, predicate) {
    if (array !== undefined) {
        const len = array.length;
        let index = 0;
        while (index < len && predicate(array[index])) {
            index++;
        }
        return array.slice(0, index);
    }
}
/** @internal */
export function skipWhile(array, predicate) {
    if (array !== undefined) {
        const len = array.length;
        let index = 0;
        while (index < len && predicate(array[index])) {
            index++;
        }
        return array.slice(index);
    }
}
/** @internal */
export function isNodeLikeSystem() {
    // This is defined here rather than in sys.ts to prevent a cycle from its
    // use in performanceCore.ts.
    return typeof process !== "undefined"
        && !!process.nextTick
        && !process.browser
        && typeof require !== "undefined";
}
