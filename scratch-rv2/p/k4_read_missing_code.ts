import { readFileSync } from 'node:fs';
const own = Object.prototype.hasOwnProperty;
try {
    readFileSync("/nonexistent-adamic-probe/missing.txt", "utf8");
} catch (caught) {
    if (caught instanceof Error) {
        console.log(`${String(own.call(caught, "code"))} ${String(own.call(caught, "name"))}`);
    }
}
