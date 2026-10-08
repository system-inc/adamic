package javascript

// The native runtime has no ICU linkage. Refuse exactly the same supported-zone
// boundary here, while using V8's intrinsic formatting for admitted dates.
const dateStringJavaScript = `(date => {
 const time = Date.prototype.getTime.call(date);
 if (Number.isNaN(time)) return 'Invalid Date';
 let zone = process.env.TZ;
 if (zone === undefined) { try { const target = adamicNodeFSFile.readlinkSync('/etc/localtime'); const at = target.indexOf('/zoneinfo/'); if (at >= 0) zone = target.slice(at + 10); } catch {} }
 let reason;
 if (zone !== 'UTC' && zone !== 'Etc/UTC' && zone !== 'America/Denver') reason = 'Date.toString: exact ICU long zone name unavailable for runtime TZ';
 else if (zone === 'America/Denver' && (time < 0 || Math.floor(time / 1000) > 2147483647)) reason = 'Date.toString: America/Denver outside the supported 1970-2038 zone transition range';
 if (reason !== undefined) { const error = new Error(reason); error.name = 'DateStringNotYet'; throw error; }
 return Date.prototype.toString.call(date);
})`
