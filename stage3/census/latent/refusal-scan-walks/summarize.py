"""Render exact rule and file-root counts, without discarding zero-cost rules."""
import csv,json,pathlib,sys
root=pathlib.Path(sys.argv[1])
rows=[]; files=[]; sites=[]
for name in ('utilities','checker'):
    path=root/('counts-'+name+'.json')
    data=json.loads(path.read_text())
    plain=json.loads((root/('plain-'+name+'.json')).read_text())
    assert data['findings']==plain['findings']
    assert data['findings_sha256']==plain['findings_sha256']
    for rule in [*data['rules'],data['outer']]:
        is_outer = rule is data['outer']
        rows.append([name,rule['Name'],'' if is_outer else rule['Calls'],rule['FileWalks'],rule['ChildWalks'],rule['Children'],'' if is_outer else f"{rule['Seconds']:.9f}"])
        for file,count in sorted((rule.get('Files') or {}).items()):
            files.append([name,rule['Name'],file,count])
        for site,counts in sorted((rule.get('Sites') or {}).items()):
            sites.append([name,rule['Name'],site,counts['FileWalks'],counts['ChildWalks'],counts['Children']])
for filename,header,data in [
    ('rules.csv',['scan','rule','calls','file_root_entries','child_walk_calls','children_entered','instrumented_inclusive_seconds'],rows),
    ('files.csv',['scan','rule','file','root_entries'],files),
    ('walk-sites.csv',['scan','rule','site','file_root_entries','child_walk_calls','children_entered'],sites)]:
    with (root/filename).open('w') as f:
        writer=csv.writer(f,lineterminator="\n");writer.writerow(header);writer.writerows(data)
