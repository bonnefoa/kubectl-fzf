#!/usr/bin/env python2

from distutils.core import setup
from setuptools import find_packages


setup(
    name='kubectl_fzf_cache_builder',
    description='fzf',
    version='0.1.0',
    url='https://github.com/bonnefoa/kubectl-fzf',
    packages=find_packages(),
    install_requires=[
        'kubernetes',
    ],
    test_suite="tests",
    entry_points={
        'console_scripts': [
            'kubectl_fzf_cache_builder=kubectl_fzf_cache_builder.cli:main'
        ]
    },
)
