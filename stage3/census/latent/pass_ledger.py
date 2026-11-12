"""Read the one Go step table and distinguish body coverage from whole-pass coverage."""
from functools import lru_cache
from pathlib import Path
import re


@lru_cache(maxsize=1)
def required_steps():
    source = Path(__file__).with_name('pass_ledger.go.txt').read_text()
    marker = 'var latentPostLoweringSteps = []latent.PassStep{'
    if source.count(marker) != 1:
        raise ValueError('missing or duplicate production step table')
    block = source.split(marker, 1)[1].split('\n}', 1)[0]
    names = tuple(re.findall(r'^\s*\{Name:\s*"([^"]+)"', block, re.M))
    if not names or len(set(names)) != len(names):
        raise ValueError('empty or duplicate production step names')
    return names


def whole_passes_measured(unit):
    ledger = unit.get('pass_ledger')
    return (unit.get('status') in ('attempted', 'speculative') and isinstance(ledger, list) and
            tuple(row.get('name') for row in ledger if isinstance(row, dict)) == required_steps() and
            len(ledger) == len(required_steps()) and
            all(row.get('status') == 'executed' for row in ledger))


def body_lowers(unit, findings):
    if unit.get('status') not in ('attempted', 'speculative'):
        return False
    # The producer is authoritative when present; old streams remain readable.
    if any(isinstance(row, dict) and row.get('status') == 'not_reached_lowering_stop'
           for row in (unit.get('pass_ledger') or [])):
        return False
    return not any(finding.get('unit') == unit.get('where') and
                   finding.get('phase') in ('lowering', 'deferred', 'registration')
                   for finding in findings)


def coverage(records):
    units = [(unit, record.get('findings', [])) for record in records for unit in record.get('units', [])]
    good = [unit for unit, findings in units if body_lowers(unit, findings)]
    return {'body_lowered_units': len(good),
            'whole_program_passes_measured_units': sum(whole_passes_measured(unit) for unit in good)}


def coverage_line(summary):
    return (f"{summary['body_lowered_units']} units lower without a body stop; "
            f"{summary['whole_program_passes_measured_units']} have their whole-program passes measured")
