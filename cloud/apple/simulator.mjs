// Simulator selection and console receipt handling, shared with the Linux probes.
import { readFileSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

export function selectDevice(listing) {
    const devices = [];
    for (const [runtime, entries] of Object.entries(listing.devices)) {
        const match = /\.iOS-(\d+(?:-\d+)*)$/.exec(runtime);
        if (match === null) continue;
        const version = match[1].split('-').map(Number);
        if (version[0] < 17) continue;
        for (const device of entries) {
            if (device.isAvailable === true) devices.push({ ...device, version });
        }
    }
    const booted = devices.filter(device => device.state === 'Booted');
    const candidates = booted.length > 0 ? booted : devices.filter(device => device.name.startsWith('iPhone'));
    if (candidates.length === 0) throw new Error('no available iPhone simulator with iOS 17 or newer');
    candidates.sort((left, right) => {
        for (let index = 0; index < Math.max(left.version.length, right.version.length); index++) {
            const difference = (left.version[index] ?? 0) - (right.version[index] ?? 0);
            if (difference !== 0) return difference;
        }
        return left.name.localeCompare(right.name) || left.udid.localeCompare(right.udid);
    });
    return candidates.at(-1);
}

export function filterConsole(bytes) {
    const kept = [];
    let removed = false;
    for (let start = 0; start < bytes.length;) {
        const newline = bytes.indexOf(10, start);
        const end = newline === -1 ? bytes.length : newline;
        const line = bytes.subarray(start, end);
        const next = newline === -1 ? end : end + 1;
        if (!removed && /^org\.system\.adamic\.dedication: [0-9]+$/.test(line.toString('latin1'))) {
            removed = true;
        } else {
            kept.push(bytes.subarray(start, next));
        }
        start = next;
    }
    return Buffer.concat(kept);
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
    try {
        const [command, input, output] = process.argv.slice(2);
        if (command === 'select' && input && !output) {
            const device = selectDevice(JSON.parse(readFileSync(input, 'utf8')));
            const action = device.state === 'Booted' ? 'using booted' : 'booting';
            console.error(`apple: ${action} simulator ${device.name} (iOS ${device.version.join('.')}, ${device.udid})`);
            console.log(`${device.state}\t${device.udid}`);
        } else if (command === 'filter' && input && output) {
            writeFileSync(output, filterConsole(readFileSync(input)));
        } else {
            throw new Error('usage: simulator.mjs select devices.json | filter console.stdout app.stdout');
        }
    } catch (error) {
        console.error(`apple: ${error.message}`);
        process.exitCode = 1;
    }
}
