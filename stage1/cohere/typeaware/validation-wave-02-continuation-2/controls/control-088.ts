import type {Process} from 'NodeJS';declare const fake:Process;fake.stdout.write('x');fake.exit();
export {};
