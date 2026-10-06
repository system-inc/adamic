#!/usr/bin/env python3

"""Extract regexp literals and static RegExp calls from the pinned test262 tree."""

import argparse
import json
import os
import re
import subprocess
import tempfile
arguments=argparse.ArgumentParser()
arguments.add_argument("test262")
arguments.add_argument("output")
options=arguments.parse_args()
test262=os.path.abspath(options.test262)
roots=[os.path.join(test262,'test/built-ins/RegExp'),os.path.join(test262,'test/language/literals/regexp')]
items=[]
# A lexical scan. A slash starts a regexp where an expression may start; this covers
# declarations, arguments, returns, array/object values and syntax-error fixtures.
def literals(src,path):
  i=0; can=True
  while i<len(src):
    c=src[i]
    if c in ' \t\r\n': i+=1; continue
    if src.startswith('//',i):
      j=src.find('\n',i); i=len(src) if j<0 else j+1; continue
    if src.startswith('/*',i):
      j=src.find('*/',i+2); i=len(src) if j<0 else j+2; continue
    if c in "'\"`":
      q=c;j=i+1
      while j<len(src):
        if src[j]=='\\': j+=2;continue
        if src[j]==q: j+=1;break
        j+=1
      i=j;can=False;continue
    if c=='/' and can:
      j=i+1;cls=False
      while j<len(src):
        if src[j]=='\\':j+=2;continue
        if src[j]=='\n' or src[j]=='\r':break
        if src[j]=='[':cls=True
        elif src[j]==']':cls=False
        elif src[j]=='/' and not cls:
          k=j+1
          while k<len(src) and src[k].isalpha():k+=1
          items.append((src[i+1:j],src[j+1:k],os.path.relpath(path,test262)))
          i=k;can=False;break
        j+=1
      else:i+=1
      if j>=len(src) or (j<len(src) and src[j] in '\r\n'):i+=1
      continue
    # identifiers/literals end expressions, keywords permit one
    if ('A' <= c <= 'Z') or ('a' <= c <= 'z') or c in '_$':
      m=re.match(r'[A-Za-z_$][\w$]*',src[i:]); word=m.group();i+=len(word)
      can=word in ('return','throw','case','delete','void','typeof','instanceof','in','of','yield','await','else','do')
      continue
    if c.isdigit():
      m=re.match(r'(?:0[xob][0-9a-f]+|[0-9.]+)',src[i:],re.I);i+=len(m.group());can=False;continue
    can=c in '=(:,[!&|?{};~%^*+-<>'
    i+=1
# Static constructor string arguments, evaluated by Node later.
ctor=re.compile(r'(?<![\w$])(?:new\s+)?RegExp\s*\(\s*((?:"(?:\\.|[^"\\])*"|\'(?:\\.|[^\'\\])*\'|`(?:\\.|[^`\\$])*`))\s*(?:,\s*((?:"(?:\\.|[^"\\])*"|\'(?:\\.|[^\'\\])*\'|`(?:\\.|[^`\\$])*`)))?',re.S)
raw=[]
for suite in roots:
 for dp,_,fs in os.walk(suite):
  for f in fs:
   if not f.endswith('.js'):continue
   p=os.path.join(dp,f);s=open(p,encoding='utf-8').read();literals(s,p)
   for m in ctor.finditer(s):raw.append((m.group(1),m.group(2) or "''",os.path.relpath(p,test262)))
with tempfile.TemporaryDirectory() as temporary:
 raw_path=os.path.join(temporary,"constructors.json")
 open(raw_path,'w').write(json.dumps(raw))
 program=r"""
const fs=require('fs'),vm=require('vm');
const rows=JSON.parse(fs.readFileSync(process.argv[1]));
for(const [pattern,flags,source] of rows){
 try { process.stdout.write(JSON.stringify([vm.runInNewContext(pattern),vm.runInNewContext(flags),source])+"\n"); } catch {}
}
"""
 evaluated=subprocess.run(['node','-e',program,raw_path],check=True,capture_output=True,text=True).stdout
 for line in evaluated.splitlines(): items.append(tuple(json.loads(line)))
files = []
for suite in roots:
 for directory, _, names in os.walk(suite):
  for name in names:
   path = os.path.join(directory, name)
   if name.endswith(".js") and re.search(r"(?:new\s+)?RegExp\s*\(", open(path, encoding="utf-8").read()):
    files.append(os.path.relpath(path, test262))
executor = os.path.join(os.path.dirname(__file__), "execute.js")
dynamic = subprocess.run(["node", executor, test262], input=json.dumps(files),
                         check=True, capture_output=True, text=True, timeout=180)
for line in dynamic.stdout.splitlines():
 items.append(tuple(json.loads(line)))
unique = []
seen = set()
for row in items:
 key = row[:2]
 if key not in seen:
  seen.add(key)
  unique.append(row)
open(options.output,'w').write(json.dumps(unique,separators=(',',':')))
print(f"extracted {len(unique)} distinct pattern and flag combinations")
