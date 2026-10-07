declare var process: NodeJS.Process;
declare module "node:process" {
    global {
        var process: NodeJS.Process;
        namespace NodeJS {
            interface WritableStream { write(chunk: string | Uint8Array, callback?: (error?: Error | null) => void): boolean }
            interface WriteStream extends WritableStream { columns: number }
            interface ReadStream { isTTY?: boolean; setRawMode(mode: boolean): this; resume(): this; pause(): this }
            interface Process {
                stdout: WriteStream & { fd: 1 };
                stderr: WriteStream & { fd: 2 };
                stdin: ReadStream & { fd: 0 };
                exitCode: number | string | null | undefined;
                exit(code?: number | string | null): never;
                on(event: string, listener: (...args: any[]) => void): this;
            }
        }
    }
    export = process;
}
declare module "process" {
    import process = require("node:process");
    export = process;
}
declare module "node:console" {
    namespace console {
        interface Console {
            log(...data: any[]): void;
            info(...data: any[]): void;
            debug(...data: any[]): void;
            warn(...data: any[]): void;
            error(...data: any[]): void;
            trace(...data: any[]): void;
            table(tabularData?: any, properties?: string[]): void;
            dir(item?: any): void;
            dirxml(...data: any[]): void;
            group(...data: any[]): void;
            time(label?: string): void;
        }
    }
    var console: console.Console;
    export = console;
}
declare namespace NodeJS {interface Timeout{unref():this;ref():this;}}
declare function require(name:string):unknown;
declare const flag:boolean;declare const ms:number;declare const asyncItems:AsyncIterable<string>;
declare function run():Promise<void>;declare function work():Promise<string>;
declare function use(value:unknown):void;declare function onEvent(callback:()=>void):void;
declare class Host {constructor(callback:()=>void);}
declare module 'node:timers/promises' {export function setTimeout(delay:number):Promise<void>;}
declare module 'NodeJS' {export interface Process {stdout:{write(text:string):void};exit():never;}}
