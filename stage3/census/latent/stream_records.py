"""Adapt completed header/files/relation-trailer streams for census readers."""


def census_records(rows):
    """Keep file rows separate; expose only relation answers actually observed."""
    reports = [index for index, row in enumerate(rows)
               if isinstance(row, dict) and row.get('status') == 'relation_report']
    if not reports:
        return rows
    if reports != [len(rows) - 1]:
        raise ValueError('relation_report must occur exactly once, at the end')
    if not isinstance(rows[0], dict) or rows[0].get('status') != 'measurement':
        raise ValueError('relation_report requires the original program header')
    trailer = rows[-1]
    relation = trailer['relation_overflow']
    if not isinstance(relation.get('after_load'), str) or not isinstance(relation.get('end'), str):
        raise ValueError('completed relation_report must contain both endpoint answers')
    header = dict(rows[0], relation_overflow=relation, findings=trailer['findings'])
    return [header, *rows[1:-1]]
