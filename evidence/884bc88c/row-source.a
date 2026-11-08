import * as fs from 'node:fs';
const watcher = fs.watch('.');
const view: { setMaxListeners<T>(n: number | string): unknown } = watcher;
view.setMaxListeners("x");
watcher.close();
