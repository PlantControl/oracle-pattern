import copy
import gzip
import json
import math
import os
from pathlib import Path
import shutil
import sys
import tempfile
import unittest
from unittest import mock

import sb03md as oracle


class SB03MDBaselineTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.manifest,cls.records=oracle.load_baseline(oracle.BASELINE)
        cls.by_id={r['case']['id']:r for r in cls.records}

    def test_corrupt_missing_empty_duplicate_policy_trace_and_observations_fail(self):
        with tempfile.TemporaryDirectory() as work:
            root=Path(work)
            with self.assertRaises(FileNotFoundError):oracle.load_baseline(root)
            for mode in ('checksum','empty','duplicate','policy','trace','observation','nonfinite','missing-svd'):
                with self.subTest(mode=mode):
                    records=copy.deepcopy(self.records[:1])
                    manifest=copy.deepcopy(self.manifest)
                    if mode=='empty':records=[]
                    if mode=='duplicate':records*=2
                    if mode=='policy':manifest['policy']='unknown'
                    if mode=='trace':records[0]['trace']['caller']='unknown'
                    if mode=='observation':records[0]['expected']['scale']='2'
                    if mode=='nonfinite':records[0]['expected']['scale']='nan'
                    if mode=='missing-svd':del records[0]['expected']['true_sep']
                    (root/'calls.jsonl.gz').write_bytes(gzip.compress(''.join(json.dumps(r)+'\n' for r in records).encode()))
                    manifest.update(count=len(records),calls_sha256=oracle.common.digest(root/'calls.jsonl.gz'))
                    if mode=='checksum':manifest['calls_sha256']='invalid'
                    (root/'manifest.json').write_text(json.dumps(manifest))
                    with self.assertRaises(ValueError):oracle.load_baseline(root)

    def test_schur_checks_reject_overlapping_and_real_blocks(self):
        for a in ([1,1,0,-1,1,1,0,-1,1],[1,1,0,1,1,0,0,0,2]):
            c=oracle.case('bad-schur',3,fact='N',job='S')
            c['a']=oracle.common.strings(a)
            o=dict(a=c['a'],u=oracle.flat(oracle.identity(3)))
            self.assertTrue(oracle.schur_checks(c,o))

    def test_no_overwrites_or_empty_selection(self):
        with mock.patch.object(oracle,'build_reference',side_effect=AssertionError('unexpected build')):
            with self.assertRaises(FileExistsError):oracle.capture(oracle.BASELINE,[42],0)
        with self.assertRaisesRegex(ValueError,'overwrite'):
            oracle.evaluate(oracle.BASELINE,oracle.BASELINE/'manifest.json')
        with self.assertRaises(ValueError):oracle.common.select(self.records,'missing')

    def test_comparator_rejects_wrong_solution_condition_state_and_spectrum(self):
        for name in ('grid-CBFN-3','grid-DBNN-3','rhs-CF-1e-150','grid-CBFN-0','invalid-1-3'):
            r=self.by_id[name]
            for key in ('info','padding_ok','scale','sep','ferr','a','u','c','wr','wi'):
                with self.subTest(case=name,key=key):
                    o=copy.deepcopy(r['expected'])
                    if isinstance(o[key],list):
                        if not o[key]:continue
                        o[key][0]='9999'
                    elif key=='info':o[key]=-99
                    elif key=='padding_ok':o[key]=False
                    else:o[key]='9999'
                    self.assertTrue(oracle.compare(r['case'],r['expected'],o)[0])
        r=self.by_id['rhs-CF-1e-150'];o=copy.deepcopy(r['expected']);o['c']=['0']*len(o['c'])
        self.assertTrue(oracle.compare(r['case'],r['expected'],o)[0])


class SB03MDExecutableTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.temp=tempfile.TemporaryDirectory(prefix='sb03md-validation-');cls.work=Path(cls.temp.name)
        cls.exe,_=oracle.build_candidate(cls.work)
        cls.manifest,records=oracle.load_baseline(oracle.BASELINE)
        cls.by_id={r['case']['id']:r for r in records}

    @classmethod
    def tearDownClass(cls):cls.temp.cleanup()

    def observe(self,c,exe=None):
        return json.loads(oracle.common.invoke([str(exe or self.exe)],data=json.dumps(c)))

    def test_independent_anchors(self):
        for name,x,sep,ferr in [('scalar-CN--1.0',-1,2,oracle.EPS/2),('scalar-DN-0.0',-2,1,0)]:
            r=self.by_id[name]
            for o in (r['expected'],self.observe(r['case'])):
                self.assertEqual(list(map(float,o['c'])),[x]);self.assertEqual(float(o['sep']),sep);self.assertEqual(float(o['ferr']),ferr)
        for d in 'CD':
            r=self.by_id[f'identity-anchor-{d}']
            for o in (r['expected'],self.observe(r['case'])):
                for x,y in zip(map(float,o['c']),[1,0,0,1]):self.assertLess(abs(x-y),1e-14)

    def test_block_scaling_and_pivot_status(self):
        for d in 'CD':
            for tr in 'NT':
                for label in ('block-scaling','early-pivot','final-pivot','near'):
                    name=f'{label}-{d}{tr}'+('-4' if label=='block-scaling' else '')
                    r=self.by_id[name];o=self.observe(r['case'])
                    self.assertEqual(oracle.compare(r['case'],r['expected'],o)[0],[])
                    if label=='block-scaling':self.assertLess(float(o['scale']),1)
                    if label=='early-pivot':self.assertEqual(o['info'],3)
                    if label=='final-pivot':self.assertEqual(o['info'],0)

    def test_reference_independent_replay_and_identity(self):
        report=self.work/'replay.json'
        before={p.name:oracle.common.digest(p) for p in oracle.BASELINE.iterdir()}
        with mock.patch.object(oracle,'build_reference',side_effect=AssertionError('legacy build attempted')):
            with mock.patch.dict(os.environ,{'FC':'/missing','ORACLE_BLAS_LIBS':'-lmissing','CGO_ENABLED':'0','GOPROXY':'off'}):
                self.assertEqual(oracle.evaluate(oracle.BASELINE,report,'grid-DBFN-6'),0)
        result=json.loads(report.read_text())
        self.assertEqual(result['baseline']['seeds'],self.manifest['seeds'])
        self.assertEqual(result['baseline']['calls_sha256'],self.manifest['calls_sha256'])
        self.assertEqual(before,{p.name:oracle.common.digest(p) for p in oracle.BASELINE.iterdir()})
        with self.assertRaises(RuntimeError):
            oracle.common.invoke([sys.executable,'oracle/sb03md.py','verify','--case','missing','--report',str(report)])

    def test_invalid_arguments_and_malformed_protocol(self):
        for r in self.by_id.values():
            if r['case']['invalid']:
                self.assertEqual(oracle.compare(r['case'],r['expected'],self.observe(r['case']))[0],[])
        c=self.by_id['grid-CBFN-3']['case']
        for text in ('',json.dumps(dict(c,a=[])),json.dumps(c)+' {}',json.dumps(dict(c,n=-1))):
            with self.assertRaises(RuntimeError):oracle.common.invoke([str(self.exe)],data=text)

    def test_compiled_negative_controls(self):
        source=oracle.ROOT/'lyapunov/solve.go';original=source.read_text()
        edits={
            'ignore-transpose':('tr != blas.NoTrans, n,','false, n,'),
            'drop-scaling':('out.Scale, out.NearSingular = solveSchur','_, out.NearSingular = solveSchur'),
            'signed-zero-estimator':('x[i] = 0','x[i] = value'),
            'corrupt-factors':('return out, nil\n}\n\nfunc square','a.Data[0] += 1\n\treturn out, nil\n}\n\nfunc square'),
        }
        selected=['grid-CBFN-3','grid-CBFT-3','grid-DBFN-6','scaled-CN']
        for name,(old,new) in edits.items():
            with self.subTest(name=name):
                self.assertEqual(original.count(old),1,'mutation site changed')
                mutated=self.work/(name+'.go');mutated.write_text(original.replace(old,new))
                overlay=self.work/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(source):str(mutated)}}))
                exe=self.work/name
                oracle.common.invoke(['go','build','-overlay',str(overlay),'-o',str(exe),'./cmd/sb03md-go'])
                failures=[]
                for key in selected:
                    r=self.by_id[key];failures += oracle.compare(r['case'],r['expected'],self.observe(r['case'],exe))[0]
                self.assertTrue(failures,'accepted defective candidate')


if __name__=='__main__':unittest.main()
