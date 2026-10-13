#!/usr/bin/env python3
"""Run both verification workflows without a reference checkout or native libraries."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

ROOT=Path(__file__).resolve().parents[1]


def main():
    with tempfile.TemporaryDirectory(prefix='oracle-offline-') as directory:
        root=Path(directory)
        for name in ('go.mod','go.sum','Makefile'):
            shutil.copyfile(ROOT/name,root/name)
        for name in ('cmd','hessenberg','lyapunov','oracle'):
            shutil.copytree(ROOT/name,root/name,ignore=shutil.ignore_patterns('reference','__pycache__','*.pyc'))
        if (root/'reference').exists() or (root/'oracle/reference').exists():
            raise RuntimeError('offline copy contains reference sources')
        env=dict(os.environ,CGO_ENABLED='0',GOPROXY='off',GOSUMDB='off',GOWORK='off',
            FC='/missing/fortran',ORACLE_BLAS_LIBS='-lmissing')
        subprocess.run(['make','verify'],cwd=root,env=env,check=True,timeout=300)
        print('Clean-copy offline verification passed for both migrations.')


if __name__=='__main__':main()
