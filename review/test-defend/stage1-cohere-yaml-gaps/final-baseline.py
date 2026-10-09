exec(open('/tmp/defend-yaml/baseline.py').read().split('groups=')[0])
allgroups=json.loads((p/'groups.json').read_text());old=allgroups.pop('formatter-witnesses');groups={}
for row in old:groups['witness-'+row]=[row]
allgroups.update(groups);(p/'groups.json').write_text(json.dumps(allgroups,indent=2));groups['products']=allgroups['products']
s=open('/tmp/defend-yaml/baseline.py').read();exec(s[s.index('results=[]'):].replace("'baseline-groups.json'","'baseline-final-groups.json'"))
