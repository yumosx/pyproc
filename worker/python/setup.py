"""
Setup script for pyproc-worker Python package
"""

from setuptools import setup, find_packages
import os

# Read the README file
def read_long_description():
    here = os.path.abspath(os.path.dirname(__file__))
    readme_path = os.path.join(here, 'README.md')
    if os.path.exists(readme_path):
        with open(readme_path, 'r', encoding='utf-8') as f:
            return f.read()
    return ""

setup(
    name='pyproc-worker',
    version='0.1.0',
    author='YuminosukeSato',
    author_email='',
    description='Python worker for pyproc - Call Python from Go without CGO',
    long_description=read_long_description(),
    long_description_content_type='text/markdown',
    url='https://github.com/YuminosukeSato/pyproc',
    packages=find_packages(),
    classifiers=[
        'Development Status :: 4 - Beta',
        'Intended Audience :: Developers',
        'Topic :: Software Development :: Libraries :: Python Modules',
        'License :: OSI Approved :: Apache Software License',
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.7',
        'Programming Language :: Python :: 3.8',
        'Programming Language :: Python :: 3.9',
        'Programming Language :: Python :: 3.10',
        'Programming Language :: Python :: 3.11',
        'Programming Language :: Python :: 3.12',
        'Operating System :: POSIX',
        'Operating System :: Unix',
        'Operating System :: MacOS',
    ],
    python_requires='>=3.7',
    install_requires=[],
    extras_require={
        'dev': [
            'pytest>=7.0',
            'pytest-asyncio>=0.21',
            'ruff>=0.1.0',
        ],
        'examples': [
            'numpy>=1.21',
            'pandas>=1.3',
            'scikit-learn>=1.0',
        ],
    },
    entry_points={
        'console_scripts': [
            'pyproc-worker=pyproc_worker.cli:main',
        ],
    },
    project_urls={
        'Bug Reports': 'https://github.com/YuminosukeSato/pyproc/issues',
        'Source': 'https://github.com/YuminosukeSato/pyproc',
        'Documentation': 'https://github.com/YuminosukeSato/pyproc#readme',
    },
    keywords='pyproc worker go python ipc rpc unix-socket',
    license='Apache-2.0',
    platforms=['Linux', 'Mac OS X', 'Unix'],
)