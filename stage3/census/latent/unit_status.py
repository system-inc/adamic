"""Explicit unit eligibility and deferred coverage, shared by census reports."""
DEFERRED = 'deferred_generic_instantiation'
ATTEMPTED = {'attempted', 'panic', 'speculative'}
SKIPPED = {'skipped_checker_body', 'split_checker_body'}


def attempted(unit):
    assert unit['status'] in ATTEMPTED | SKIPPED | {DEFERRED}, 'unknown unit status'
    return unit['status'] in ATTEMPTED


def body_eligible(unit, diagnostics, skipped='skipped_checker_body'):
    if unit['status'] == DEFERRED:
        assert type(unit.get('instantiated')) is bool, 'deferred call witness missing'
        return True
    return (unit['status'] == skipped) == bool(diagnostics)


def deferred_summary(units):
    deferred = [u for u in units if u['status'] == DEFERRED]
    for unit in deferred:
        assert type(unit.get('instantiated')) is bool, 'deferred call witness missing'
    spans = {}
    for unit in deferred:
        spans.setdefault(unit['where'].rsplit(':', 2)[0], []).append((unit.get('body_start', 0), unit.get('body_end', 0)))
    body_bytes = 0
    for group in spans.values():
        end = 0
        for left, right in sorted(group):
            body_bytes += max(0, right - max(end, left))
            end = max(end, right)
    return dict(deferred_body_bytes=body_bytes, deferred_generic_instantiation=len(deferred),
                instantiated_generics=sum(u['instantiated'] for u in deferred),
                never_reached=[dict(where=u['where'], name=u.get('name', ''),
                                    body_start=u.get('body_start', 0), body_end=u.get('body_end', 0))
                               for u in deferred if not u['instantiated']])


def audit_deferred(summary, units):
    # Recount attempts independently of the report helper. A report mutation
    # must not inherit its own mistaken definition through a shared counter.
    for field in ('attempted_functions', 'functions_attempted'):
        if field in summary:
            count = sum(u['kind']=='KindFunctionDeclaration' and u['status'] in ('attempted', 'panic', 'speculative') for u in units)
            assert summary[field] == count, 'deferred counted as attempted'
    if 'attempted_statements' in summary:
        count = sum(u['kind']!='KindFunctionDeclaration' and u['status'] in ('attempted', 'panic', 'speculative') for u in units)
        assert summary['attempted_statements'] == count, 'deferred counted as attempted'
    expected = deferred_summary(units)
    if expected['deferred_generic_instantiation'] or 'never_reached' in summary:
        for key, value in expected.items():
            assert summary.get(key) == value, 'deferred coverage gap'


def unit_counts(units):
    result = dict(total=len(units), functions=sum(u['kind']=='KindFunctionDeclaration' for u in units),
                  skipped_checker_body=sum(u['status']=='skipped_checker_body' for u in units),
                  attempted_functions=sum(u['kind']=='KindFunctionDeclaration' and attempted(u) for u in units),
                  attempted_statements=sum(u['kind']!='KindFunctionDeclaration' and attempted(u) for u in units),
                  panicked_units=sum(u['status']=='panic' for u in units))
    if any(u['status']==DEFERRED for u in units):
        result.update(deferred_summary(units))
    return result
