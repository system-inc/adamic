export function statistics(values) {
  const sorted = [...values].sort((left, right) => left - right);
  if (!sorted.length) throw new Error('no samples');
  const middle = Math.floor(sorted.length / 2);
  return { best: sorted[0], median: sorted.length % 2 ? sorted[middle] : (sorted[middle - 1] + sorted[middle]) / 2 };
}

export function subtractStartup(measured, startup, requests) {
  if (!(requests > 0)) throw new Error('per-request denominator must be positive');
  return {
    wallMsPerRequest: (measured.wallMs - startup.wallMs) / requests,
    cpuMsPerRequest: (measured.cpuMs - startup.cpuMs) / requests,
    rawWallMsPerRequest: measured.wallMs / requests,
    rawCpuMsPerRequest: measured.cpuMs / requests,
    startupWallMs: startup.wallMs,
    startupCpuMs: startup.cpuMs,
    maxRssBytes: measured.maxRssBytes,
    startupMaxRssBytes: startup.maxRssBytes,
  };
}
