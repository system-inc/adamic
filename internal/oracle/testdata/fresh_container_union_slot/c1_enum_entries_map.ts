enum WatchFileKind { FixedPollingInterval, PriorityPollingInterval, DynamicPriorityPolling }
interface CommandLineOption {
    name: string;
    type: "string" | "number" | "boolean" | Map<string, number | string>;
}
const watchFile: CommandLineOption = {
    name: "watchFile",
    type: new Map(Object.entries({
        fixedpollinginterval: WatchFileKind.FixedPollingInterval,
        prioritypollinginterval: WatchFileKind.PriorityPollingInterval,
        dynamicprioritypolling: WatchFileKind.DynamicPriorityPolling,
    })),
};
const type = watchFile.type;
if (type instanceof Map) {
    type.set("custom", "named");
    for (const [key, value] of type) {
        console.log(key + " = " + String(value));
    }
}
