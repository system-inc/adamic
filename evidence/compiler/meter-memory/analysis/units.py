import json,sys,csv
src,out=sys.argv[1],sys.argv[2]
rows=[json.loads(l) for l in open(src)]
n=len(rows)
sec=[0.0]*n; raw=[0]*n; spread=[0.0]*n
prev_t=rows[0]['t']-0  # first unit has no earlier stamp; time unknown
for i,r in enumerate(rows):
    sec[i]=None if i==0 else (r['t']-rows[i-1]['t'])/1e9
    raw[i]=0 if i==0 else r['live_after_last_gc']-rows[i-1]['live_after_last_gc']
# spread: between changes of the live value
last_i=0
for i in range(1,n):
    if rows[i]['live_after_last_gc']!=rows[last_i]['live_after_last_gc']:
        d=rows[i]['live_after_last_gc']-rows[last_i]['live_after_last_gc']
        k=i-last_i
        for j in range(last_i+1,i+1): spread[j]=d/k
        last_i=i
w=csv.writer(open(out,'w'))
w.writerow(['n','file','where','seconds','live_delta_raw_bytes','live_delta_spread_bytes','live_after_bytes'])
for i,r in enumerate(rows):
    w.writerow([r['n'],r['file'].split('main-adapted/')[-1],r['where'].split('main-adapted/')[-1],'' if sec[i] is None else '%.3f'%sec[i],raw[i],'%.0f'%spread[i],r['live_after_last_gc']])
def rank(v):
    o=sorted(range(len(v)),key=lambda i:v[i]); r=[0]*len(v)
    for k,i in enumerate(o): r[i]=k
    return r
idx=[i for i in range(1,n)]
S=[sec[i] for i in idx]; H=[spread[i] for i in idx]; R=[raw[i] for i in idx]
def spearman(a,b):
    ra,rb=rank(a),rank(b); m=len(a); ma=sum(ra)/m; mb=sum(rb)/m
    c=sum((x-ma)*(y-mb) for x,y in zip(ra,rb)); va=sum((x-ma)**2 for x in ra); vb=sum((y-mb)**2 for y in rb)
    return c/(va*vb)**.5
print('units',n,'total seconds %.0f'%sum(S),'total live growth %.2f GiB'%((rows[-1]['live_after_last_gc']-rows[0]['live_after_last_gc'])/2**30))
print('spearman(seconds, spread heap) %.3f ; spearman(seconds, raw delta) %.3f'%(spearman(S,H),spearman(S,R)))
def line(i): return '%5d %-28s %-34s %8.2f s  %+8.1f MiB spread %+8.1f MiB raw'%(rows[i]['n'],rows[i]['file'].split('src/')[-1][:28],rows[i]['where'].split('src/')[-1].split('/')[-1][:34],sec[i],spread[i]/2**20,raw[i]/2**20)
print('\nTOP 10 BY HEAP ADDED (spread)'); 
for i in sorted(idx,key=lambda i:-spread[i])[:10]: print(line(i))
print('\nTOP 10 BY SECONDS')
for i in sorted(idx,key=lambda i:-sec[i])[:10]: print(line(i))
# top-decile overlap
k=n//10
th=set(sorted(idx,key=lambda i:-spread[i])[:k]); ts=set(sorted(idx,key=lambda i:-sec[i])[:k])
print('\ntop-decile overlap: %d of %d units are in both the top 10%% by heap and top 10%% by seconds'%(len(th&ts),k))
# share of time and heap in top decile
tt=sum(sec[i] for i in idx); print('time share of top-decile-by-seconds %.1f%%; heap share (positive spread) of top-decile-by-heap %.1f%%'%(100*sum(sec[i] for i in ts)/tt,100*sum(spread[i] for i in th)/sum(max(0,spread[i]) for i in idx)))
