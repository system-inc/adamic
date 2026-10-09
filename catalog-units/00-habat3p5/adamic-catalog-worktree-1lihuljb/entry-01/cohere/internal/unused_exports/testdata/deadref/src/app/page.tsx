import {usesReachable, usesUnreachable} from '../uses';

// A framework-spared root, so both functions above are genuinely live and the
// only difference between the two constants is reachability.
export default function Page(): string {
    return usesReachable() + usesUnreachable();
}
