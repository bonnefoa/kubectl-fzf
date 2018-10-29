#!/usr/bin/env python2

from distutils.core import setup
from setuptools import find_packages


setup(
    name='kube_watcher',
    description='fzf',
    version='0.1.0',
    url='https://github.com/bonnefoa/kubectl-fzf',
    packages=find_packages(),
    install_requires=[
        'kubernetes',
    ],
    entry_points={
        'console_scripts': [
            'kube_watcher=kube_watcher.kube_watcher:main'
        ]
    },
)
