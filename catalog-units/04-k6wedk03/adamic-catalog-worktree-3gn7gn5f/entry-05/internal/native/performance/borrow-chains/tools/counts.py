"""Compare every recorded fixture, refusing any retain or release rise."""
from pathlib import Path
import csv,json,sys

def table(path):
 result={}
 for line in Path(path).read_text().splitlines():
  cells=[cell.strip() for cell in line.split('|')[1:-1]]
  if len(cells)==7 and cells[1].isdigit():result[cells[0]]=list(map(int,cells[1:]))
 return result
before=table(sys.argv[1]);after=table(sys.argv[2]);assert before.keys()==after.keys()
rows=[];changed=0
for fixture,values in before.items():
 current=after[fixture]
 assert current[2]<=values[2] and current[3]<=values[3],fixture
 assert [current[i] for i in (0,1,4,5)]==[values[i] for i in (0,1,4,5)],fixture
 if current!=values:changed+=1
 rows.append([fixture,values[2],current[2],values[3],current[3]])
with open(sys.argv[3],'w') as output:
 writer=csv.writer(output);writer.writerow(['fixture','retains_before','retains_after','releases_before','releases_after']);writer.writerows(rows)
summary={'fixtures':len(rows),'changed':changed,'retains_removed':sum(row[1]-row[2] for row in rows),'releases_removed':sum(row[3]-row[4] for row in rows),'rises':0,'allocation_free_peak_region_changes':0}
print(json.dumps(summary,indent=2))
