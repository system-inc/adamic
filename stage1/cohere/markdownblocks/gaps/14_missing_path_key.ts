import { AstArena } from '../astArena.ts';
import { AstPath, pathIndex, pathProperty } from '../astPath.ts';
const arena = new AstArena();
const root = arena.add('root');
arena.node(root).parent = true;
const path = new AstPath(arena, root);
console.log(path.call(current => `key="${current.key()}" present=${current.key() !== ''}`, [
    pathProperty('absent'), pathProperty('children'), pathIndex(0),
]));
