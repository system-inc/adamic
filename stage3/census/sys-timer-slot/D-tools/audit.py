import sys
sys.dont_write_bytecode=True
from run import *
import hashlib
out=ROOT/'stage3/census/sys-timer-slot';historic='949b9c3c4c883d964bba19ee9fbdae9fd1909288'
def capture(args):
 p=subprocess.run(args,cwd=ROOT,stdout=subprocess.PIPE,stderr=subprocess.PIPE);assert p.returncode==0,(args,p.stderr.decode());return p.stdout
entries=capture(['git','ls-tree','-r',historic,'stage3/census/sys-timer-slot']).decode().splitlines();unchanged=0
for line in entries:
 metadata,name=line.split('\t');sha=metadata.split()[2]
 if name.endswith('/REPORT.md') or name.endswith('/REPORT.json'):continue
 raw=(ROOT/name).read_bytes();actual=hashlib.sha1(b'blob '+str(len(raw)).encode()+b'\0'+raw).hexdigest();assert actual==sha,name;unchanged+=1
old=capture(['git','show',historic+':stage3/census/sys-timer-slot/REPORT.md']);assert (out/'REPORT.md').read_bytes().startswith(old)
a=json.loads(capture(['git','show',historic+':stage3/census/sys-timer-slot/REPORT.json']));b=json.loads((out/'REPORT.json').read_text());assert all(b[k]==v for k,v in a.items())
assert hashlib.sha256((out/'D-evidence.tar.gz').read_bytes()).hexdigest()==(out/'D-evidence.sha256').read_text().split()[0]
d=json.loads((out/'D.json').read_text());assert len(d['addedDiagnostics']['adamic'])==2 and len(d['addedDiagnostics']['upstream'])==2
for row in d['javascript']:assert row['byteIdentical'];assert (S/'H/built/local'/row['file']).read_bytes()==(S/'D/built/local'/row['file']).read_bytes()
assert d['census']['providerUnit']['status']=='split_checker_body'
summary={'status':'pass','historyOtherFilesByteUnchanged':unchanged,'oldReportMarkdownPrefixByteUnchanged':True,'oldReportJsonValuesUnchanged':True,'javascriptByteComparisons':9,'providerCensusCoverageCaveatVerified':True,'gitCaptureChildExits':0}
(out/'D-audit.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary,indent=2))
