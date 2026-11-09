#!/usr/bin/env python3
"""The daily meter's outcomes: what run.sh records from the meter's exit, and which of them submit.sh commits.

    python3 -I outcome.py meter <exit> <report.json>   prints the outcome; for a relation budget outcome, its why next
    python3 -I outcome.py submittable <outcome>        exits 0 for an outcome that is committed to main, else 1

twice-daily.sh exits 3 when the run's record is whole but its relation budget isn't ok (#ny0c3gb): red is
OverBudget, and unknown, or a verdict that is missing or can't be read, is BudgetUnknown. Neither is green, and both
are committed, so the day's record lands on main. Any other non-zero exit is Failed and commits nothing.
"""
import json
import sys

# The meter's exit for a whole record whose relation budget isn't ok (stage3/meter/twice-daily.sh's last step).
BUDGET_EXIT = 3
# Every outcome whose run is committed to main; Running and Failed never are.
SUBMITTABLE = ('Succeeded', 'OverBudget', 'BudgetUnknown')


def meter(code, report):
    """The outcome for the meter's exit, and for a budget outcome the budget's why."""
    if code == '0':
        return 'Succeeded', ''
    if code != str(BUDGET_EXIT):
        return 'Failed', ''
    try:
        with open(report) as source:
            budget = json.load(source).get('relation_budget')
    except (OSError, ValueError, AttributeError) as error:
        return 'BudgetUnknown', f'no relation budget verdict readable in {report}: {error}'
    if not isinstance(budget, dict):
        return 'BudgetUnknown', f'no relation budget verdict in {report}'
    why = str(budget.get('why', ''))
    if budget.get('status') == 'red':
        return 'OverBudget', why
    # Unknown, and any status a 3 shouldn't carry, ok among them, is never read as green.
    return 'BudgetUnknown', why if budget.get('status') == 'unknown' else f"exit 3 with status {budget.get('status')!r}; {why}"


def main(arguments):
    if len(arguments) == 3 and arguments[0] == 'meter':
        outcome, why = meter(arguments[1], arguments[2])
        print(outcome)
        if outcome in ('OverBudget', 'BudgetUnknown'):
            print(why.replace('\n', ' '))
        return 0
    if len(arguments) == 2 and arguments[0] == 'submittable':
        return 0 if arguments[1] in SUBMITTABLE else 1
    print('usage: outcome.py meter <exit> <report.json> | outcome.py submittable <outcome>', file=sys.stderr)
    return 2


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
