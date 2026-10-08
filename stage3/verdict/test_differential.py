"""Selection must not silently pass a missing function or stale provenance."""
import copy
import gzip
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch
from coverage_worker import Workers
import coverage_worker
import io
from differential import identity, load_map, save_map, select, validate_source_map, sha, output_hash, compact_map, covered_indices, record_coverage


def example():
    rows=[{'source':f'{index}.ts','configuration':''} for index in range(20)]
    inventory={'files':{'src/compiler/a.ts':{'sha256':'old','outside_sha256':'same'}},
               'functions':[{'id':'a','index':0,'sha256':'old'},{'id':'b','index':1,'sha256':'old'}]}
    mapping={'inventory':inventory,'configurations':{identity(row):{'functions':[0] if i==0 else [1]} for i,row in enumerate(rows)}}
    current=copy.deepcopy(inventory);current['functions'][0]['sha256']='new'
    return mapping,current,rows


class Differential(unittest.TestCase):
    def test_fresh_compiler_globals_and_utf8_mutant(self):
        with tempfile.TemporaryDirectory() as scratch:
            root=Path(scratch);bundle=root/'fixture.cjs'
            bundle.write_text("let calls=0; exports.verdictHits=new Uint8Array(2); exports.verdictSystem={}; exports.verdictExecute=(sys)=>{exports.verdictHits[0]=1; if(calls===0)exports.verdictHits[1]=1; calls++;sys.write(String(calls)+'\\ud800');sys.exit(0);};")
            def observations(folder):
                folder.mkdir();values=[]
                with Workers(bundle,folder/'workers') as workers:
                    for index in range(2):
                        cwd=folder/str(index);cwd.mkdir();stdout=io.BytesIO();stderr=io.BytesIO()
                        workers.execute(['fixture'],cwd,stdout,stderr)
                        values.append((stdout.getvalue(),json.loads((cwd/'.verdict-coverage.json').read_text())))
                return values
            env={'TSC_JOBS':'1'}
            with patch.dict(__import__('os').environ,env):
                expected=[(b'1\xef\xbf\xbd',[0,1])]*2
                self.assertEqual(observations(root/'good'),expected)
                text=(coverage_worker.ROOT/'coverage_worker.cjs').read_text()
                text=text.replace('function freshCompiler() {','let cached; function freshCompiler() { if(cached) return cached;',1)
                text=text.replace('return instance.exports;','return cached=instance.exports;',1)
                (root/'coverage_worker.cjs').write_text(text)
                with patch.object(coverage_worker,'ROOT',root):
                    self.assertNotEqual(observations(root/'mutant'),expected)

    def test_changed_function_and_deterministic_sample(self):
        mapping,current,rows=example();affected,sample,report=select(mapping,current,rows)
        self.assertEqual(affected,[rows[0]])
        self.assertEqual(len(sample),8)
        self.assertNotIn(rows[0],sample)
        self.assertEqual(select(mapping,current,list(reversed(rows)))[1],sample)
        self.assertEqual(report['coverage_gaps'],[])
        # Mutant: the selector ignores the function's changed body.
        with self.assertRaises(AssertionError):self.assertEqual([],affected)

    def test_missing_function_map_is_a_gap(self):
        mapping,current,rows=example()
        mapping['configurations'][identity(rows[0])]['functions']=[]
        affected,sample,report=select(mapping,current,rows)
        self.assertEqual(affected,[])
        self.assertEqual(report['coverage_gaps'],['a'])

    def test_compact_map_preserves_selection_and_gaps(self):
        mapping,current,rows=example()
        expected=select(mapping,current,rows)
        compact_map(mapping)
        self.assertEqual(select(mapping,current,rows),expected)
        self.assertEqual(covered_indices(mapping['configurations'][identity(rows[0])]),[0])
        mapping['configurations'][identity(rows[0])]['functions_bitmap']=__import__('differential').pack_functions([],2)
        self.assertEqual(select(mapping,current,rows)[2]['coverage_gaps'],['a'])

    def test_global_change_is_a_gap(self):
        mapping,current,rows=example();current['files']['src/compiler/a.ts']['outside_sha256']='new'
        self.assertEqual(select(mapping,current,rows)[2]['outside_function_changes'],['src/compiler/a.ts'])

    def test_added_function_is_a_gap(self):
        mapping,current,rows=example();current['functions'].append({'id':'new','index':2,'sha256':'new'})
        self.assertIn('new',select(mapping,current,rows)[2]['coverage_gaps'])

    def test_deleted_function_selects_old_callers(self):
        mapping,current,rows=example();current['functions'].pop(0)
        self.assertEqual(select(mapping,current,rows)[0],[rows[0]])

    def test_map_checksum_mutant(self):
        with tempfile.TemporaryDirectory() as scratch:
            path=Path(scratch)/'map.gz';save_map(path,{'schema':1})
            self.assertEqual(load_map(path),{'schema':1})
            with gzip.open(path,'rt') as stream:data=json.load(stream)
            data['schema']=2
            with gzip.open(path,'wt') as stream:json.dump(data,stream)
            with self.assertRaisesRegex(RuntimeError,'checksum mismatch'):load_map(path)

    def test_stale_source_hash_with_valid_map_checksum(self):
        with tempfile.TemporaryDirectory() as scratch:
            root=Path(scratch);(root/'src/compiler').mkdir(parents=True);(root/'src/tsc').mkdir()
            (root/'src/compiler/input.a').write_text('fixture only')
            (root/'src/tsc/.keep').write_text('')
            # Use a real Git blob at a .ts path, without authoring a .ts program.
            subprocess.run(['git','init','-q',str(root)],check=True)
            subprocess.run(['git','-C',str(root),'add','src/tsc/.keep'],check=True)
            blob=subprocess.check_output(['git','-C',str(root),'hash-object','-w','src/compiler/input.a'],text=True).strip()
            subprocess.run(['git','-C',str(root),'update-index','--add','--cacheinfo',f'100644,{blob},src/compiler/a.ts'],check=True)
            subprocess.run(['git','-C',str(root),'-c','user.name=Fixture','-c','user.email=fixture@example.invalid','commit','-qm','Fixture'],check=True)
            files={'src/compiler/a.ts':{'sha256':sha(b'fixture only')}}
            aggregate=sha(json.dumps([['src/compiler/a.ts',files['src/compiler/a.ts']['sha256']]],separators=(',',':')).encode())
            mapping={'inventory':{'files':files,'source_sha256':aggregate}}
            validate_source_map(mapping,root,'HEAD')
            mapping['inventory']['files']['src/compiler/a.ts']['sha256']=sha(b'stale')
            path=root/'map.gz';save_map(path,mapping)
            with self.assertRaisesRegex(RuntimeError,'base source hash mismatch'):validate_source_map(load_map(path),root,'HEAD')

    def test_resumed_source_inventory_mutant(self):
        with tempfile.TemporaryDirectory() as scratch:
            root=Path(scratch);binary=root/'tsc';binary.write_text('unused')
            data={'functions':[],'files':{},'source_sha256':'same'}
            (root/'inventory.json').write_text(json.dumps(data))
            resume=root/'resume';resume.mkdir();(resume/'source-inventory.json').write_text('{"stale":true}')
            output=root/'output';output.mkdir()
            with patch('differential.inventory',return_value=data):
                with self.assertRaisesRegex(RuntimeError,'stale resumed coverage'):
                    record_coverage(binary,root,root/'map.gz',output,resume)

    def test_expected_stream_hash_mutants(self):
        wanted={'stdout':b'expected','stderr':b'','exit':b'0\n'}
        original=output_hash(wanted)
        for stream in wanted:
            mutant=dict(wanted);mutant[stream]+=b'x'
            self.assertNotEqual(output_hash(mutant),original)


if __name__=='__main__':unittest.main()
