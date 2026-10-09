// Shape from ahra modules/os/AhraOsCommandLineInterface.ts:120, the domain command loaders at boot.
// Each loader is a dynamic import plus a pure table read, so the loads are independent. The
// Object.assign into the accumulator depends on order only where two domains export the same verb,
// and assigning the Promise.all results in sequence keeps that precedence exactly.
// Expected: 1 finding.
interface CommandLineInterfaceCommandInterface {
    description: string;
}
type DomainCommandLoaderType = () => Promise<Record<string, CommandLineInterfaceCommandInterface>>;
declare const domainCommandLoaders: Array<[string, DomainCommandLoaderType]>;
declare function loadDomainCommands(
    domainName: string,
    load: DomainCommandLoaderType,
): Promise<Record<string, CommandLineInterfaceCommandInterface>>;

export async function composeCommands(): Promise<Record<string, CommandLineInterfaceCommandInterface>> {
    const composedCommands: Record<string, CommandLineInterfaceCommandInterface> = {};
    for(const [domainName, load] of domainCommandLoaders) {
        Object.assign(composedCommands, await loadDomainCommands(domainName, load));
    }
    return composedCommands;
}
