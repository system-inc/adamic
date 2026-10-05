# near_pio2.py: run with python3; it prints the inputs internal/oracle/testdata/trig_reduction.a uses.
# The doubles nearest n*pi/2, n up to 2^19 (rem_pio2's medium range), whose remainder is smallest
# relative to themselves: where the reduction cancels most bits and needs its third iteration.
from decimal import Decimal, getcontext
import struct, math
getcontext().prec = 80
PI = Decimal("3.14159265358979323846264338327950288419716939937510582097494459230781640628620899862803")
half = PI / 2
results = []
for n in range(1, 1 << 19):
    exact = half * n
    d = float(exact)
    for candidate in (d, math.nextafter(d, math.inf), math.nextafter(d, -math.inf)):
        r = abs(Decimal(candidate) - exact)
        if r == 0:
            continue
        lost = math.log2(candidate) - math.log2(float(r))
        results.append((lost, n, candidate))
results.sort(reverse=True)
for lost, n, candidate in results[:12]:
    bits = struct.unpack('<Q', struct.pack('<d', candidate))[0]
    print(f"{lost:.1f} bits cancelled, n={n}, x={candidate!r}, bits={bits:016x}")
