"""Native-toolchain checks; excluded from ordinary offline verification."""
import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock

import sb03md as oracle
import capture_sb03mx


class SB03MDCaptureTests(unittest.TestCase):
    def test_capture_without_go_then_audit(self):
        with tempfile.TemporaryDirectory() as work:
            target=Path(work)/'baseline'
            with mock.patch.object(oracle,'build_candidate',side_effect=AssertionError('legacy operation built Go')):
                oracle.capture(target,[91],0,'identity-anchor-C')
                manifest,records=oracle.load_baseline(target)
                self.assertEqual(manifest['seeds'],[91])
                for x,y in zip(map(float,records[0]['expected']['c']),[1,0,0,1]):self.assertLess(abs(x-y),1e-14)
                self.assertEqual(oracle.evaluate(target,Path(work)/'audit.json',audit=True),0)
                with self.assertRaises(FileExistsError):oracle.capture(target,[92],0)

    def test_estimator_probe_capture_and_immutability(self):
        fixture=oracle.ROOT/'lyapunov/testdata/sb03mx-probes.json'
        with self.assertRaises(FileExistsError):capture_sb03mx.capture(fixture)
        with tempfile.TemporaryDirectory() as work:
            target=Path(work)/'probes.json';capture_sb03mx.capture(target)
            actual=json.loads(target.read_text());expected=json.loads(fixture.read_text())
            self.assertEqual(actual['a'],expected['a'])
            self.assertEqual(len(actual['probes']),len(expected['probes']))
            for x,y in zip(actual['probes'],expected['probes']):
                for key in ('name','trans','perturbed'):self.assertEqual(x[key],y[key])
                for key in ('before','after'):
                    for a,b in zip(x[key],y[key]):self.assertLessEqual(abs(a-b),4096*oracle.EPS*max(abs(a),abs(b)))
                self.assertEqual(x['scale'],y['scale'])


if __name__=='__main__':unittest.main()
