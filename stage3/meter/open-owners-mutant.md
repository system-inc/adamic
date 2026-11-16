# Open-owner mutant

Mutant: skip the open-task membership check in `reason_owner` by deleting:

```python
    if task not in open_tasks:
        return 'OWNER BLANK'
```

A scratch reporter with this deletion turned
`LatentReportTests.test_closed_task_is_blank_and_in_unowned` red (exit 1):

```text
AssertionError: '#drwm5pj (Lower)' != 'OWNER BLANK'
```

The unchanged reporter passes the same case. The case also checks that the
closed task's reason appears in the Unowned table and that its top-reason
owner is blank. The mutant edits no production source and contacts no store.
