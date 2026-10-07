// messageConsistencyNoConsole names the console method called.
export function messageConsistencyNoConsole(methodName: string): string {
    return `Do not call 'console.${methodName}'. Reach the tier that owns this failure and call '.log.error(identifier, data, error)' or '.log.warning(...)' for a row, or '.log.debug(message)' for a line that never becomes one.`;
}
