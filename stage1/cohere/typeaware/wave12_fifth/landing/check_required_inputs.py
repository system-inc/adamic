"""Validate actual Go correctness events: skips, failures and absent packages fail."""
import json
import sys
from pathlib import Path

EXPECTED = {
    'github.com/system-inc/adamic/stage1/typescript/scanner',
    'github.com/system-inc/adamic/stage1/typescript/parser',
    'github.com/system-inc/adamic/stage1/cohere/css',
    'github.com/system-inc/adamic/stage1/cohere/graphql',
    'github.com/system-inc/adamic/stage1/cohere/selector',
    'github.com/system-inc/adamic/stage1/cohere/values',
    'github.com/system-inc/adamic/stage1/cohere/mediaquery',
    'github.com/system-inc/adamic/stage1/cohere/json',
    'github.com/system-inc/adamic/stage1/cohere/lint',
}

PARENTS = {
    ('scanner', 'TestScannerAgreesWithTypescriptGo'),
    ('parser', 'TestCompilerExpressionsAgree'),
    ('parser', 'TestWholeCompilerAgrees'),
    ('css', 'TestThePortParsesAsGoCohereDoes'),
    ('css', 'TestCSSPrinterAgreesWithGo'),
    ('css', 'TestCSSPrinterBoundaryProofs'),
    ('graphql', 'TestThePortParsesAsGoCohereDoes'),
    ('selector', 'TestThePortParsesAsGoCohereDoes'),
    ('selector', 'TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars'),
    ('values', 'TestThePortParsesAsGoCohereDoes'),
    ('mediaquery', 'TestThePortParsesAsGoCohereDoes'),
    ('json', 'TestUpstreamNumericSeparatorGap'),
    ('json', 'TestUpstreamRepositoryCorpusParity'),
    ('json', 'TestExternalComparisonCatchesThreePrinterMutants'),
    ('lint', 'TestCompilerAndStage1Agree'),
}

def validate(path):
    events = [json.loads(line) for line in Path(path).read_text().splitlines() if line.startswith('{')]
    for event in events:
        if event.get('Action') in {'skip', 'fail'}:
            raise ValueError('required correctness check ' + event['Action'] + ': ' + event.get('Package', '') + '/' + event.get('Test', ''))
    passed = {event['Package'] for event in events if event.get('Action') == 'pass' and 'Test' not in event}
    if passed != EXPECTED:
        raise ValueError('required package results missing or unexpected: ' + repr(passed.symmetric_difference(EXPECTED)))
    tests = [event for event in events if event.get('Action') == 'pass' and event.get('Test')]
    parents = {(event['Package'].rsplit('/', 1)[-1], event['Test']) for event in tests if '/' not in event['Test']}
    if parents != PARENTS:
        raise ValueError('required parent checks missing or unexpected: ' + repr(parents.symmetric_difference(PARENTS)))
    return len(tests)

if __name__ == '__main__':
    try:
        print('Required-input checks PASS:', validate(sys.argv[1]), 'tests; nine packages; zero skips')
    except (ValueError, KeyError, IndexError) as error:
        print('Required-input checks FAIL:', error)
        raise SystemExit(1)
