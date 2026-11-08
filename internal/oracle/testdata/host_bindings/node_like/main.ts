// isNodeLikeSystem's require half, a guarded read and typeof, held to Node in its module mode.
import { isNodeLikeSystem, loader, requireType } from './core.ts';

console.log(String(isNodeLikeSystem()));
console.log(loader());
console.log(requireType());
