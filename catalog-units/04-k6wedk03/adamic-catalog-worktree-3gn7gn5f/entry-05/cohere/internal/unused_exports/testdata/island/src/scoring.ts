// The island: an abandoned export plus the private world that existed only to serve it.
const DecayTable = [1, 2, 3];

function normalizeDecay(value: number): number {
  return value / DecayTable.length;
}

export function ScoreWindow(value: number): number {
  return normalizeDecay(value);
}

// Live, for contrast: exported and actually used from a rooted file.
export function LiveHelper(): number {
  return 42;
}

// Recursive and nothing else calls it. Referenced only by itself.
export function recursiveOrphan(n: number): number {
  return n <= 0 ? 0 : recursiveOrphan(n - 1);
}
