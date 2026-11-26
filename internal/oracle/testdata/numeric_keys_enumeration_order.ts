enum PollingInterval {
    High = 2000,
    Medium = 500,
    Low = 250,
}

interface Levels {
    Low: number;
    Medium: number;
    High: number;
}

function createPollingIntervalBasedLevels(levels: Levels) {
    return {
        [PollingInterval.Low]: levels.Low,
        [PollingInterval.Medium]: levels.Medium,
        [PollingInterval.High]: levels.High,
    };
}

const table = createPollingIntervalBasedLevels({ Low: 32, Medium: 64, High: 256 });
console.log(Object.keys(table).join(",") + ` ${table[PollingInterval.Low]} ${table[PollingInterval.Medium]} ${table[PollingInterval.High]}`);
