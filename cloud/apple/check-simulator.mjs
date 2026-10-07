// Synthetic simctl inputs test selection and transport bytes without an Apple SDK.
import assert from 'node:assert/strict';
import { resolve } from 'node:path';
import { pathToFileURL } from 'node:url';
const { selectDevice, filterConsole } = await import(pathToFileURL(resolve(process.argv[2] ?? 'cloud/apple/simulator.mjs')));
const device = (name, udid, state = 'Shutdown', isAvailable = true) => ({ name, udid, state, isAvailable });
const listing = { devices: {
    'com.apple.CoreSimulator.SimRuntime.iOS-26-9': [device('iPhone 17', 'older')],
    'com.apple.CoreSimulator.SimRuntime.iOS-26-10': [device('iPhone 17 Pro', 'newest')],
    'com.apple.CoreSimulator.SimRuntime.iOS-27-0': [device('iPad Pro', 'ipad')],
    'com.apple.CoreSimulator.SimRuntime.iOS-28-0': [device('iPhone 18', 'unavailable', 'Shutdown', false)],
    'com.apple.CoreSimulator.SimRuntime.watchOS-29-0': [device('Apple Watch', 'watch')],
} };
assert.equal(selectDevice(listing).udid, 'newest');
listing.devices['com.apple.CoreSimulator.SimRuntime.iOS-26-9'][0].state = 'Booted';
assert.equal(selectDevice(listing).udid, 'older');
assert.throws(() => selectDevice({ devices: {} }), /no available iPhone/);
assert.throws(() => selectDevice({ devices: { 'com.apple.CoreSimulator.SimRuntime.iOS-16-0': [device('iPhone', 'old')] } }), /no available iPhone/);
const output = Buffer.from('In dedication\n');
const receipt = Buffer.from('org.system.adamic.dedication: 1234\n');
for (const parts of [[receipt, output], [output, receipt], [output.subarray(0, 3), Buffer.from('\n'), receipt, output.subarray(3)]]) {
    assert.deepEqual(filterConsole(Buffer.concat(parts)), Buffer.concat(parts.filter(part => part !== receipt)));
}
for (const bytes of [output, Buffer.from('missing newline'), Buffer.from([0, 255, 10]),
                     Buffer.from('prefix org.system.adamic.dedication: 1234\n'),
                     Buffer.from('org.system.adamic.dedication: 1234 suffix\n'),
                     Buffer.from('org.system.adamic.dedication: nope\n')]) {
    assert.deepEqual(filterConsole(bytes), bytes);
}
assert.deepEqual(filterConsole(Buffer.concat([receipt, output, receipt])), Buffer.concat([output, receipt]));
console.log('PASS: newest runtime, booted reuse, missing devices, receipt positions and unchanged bytes');
