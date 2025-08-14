from setuptools import setup, find_packages

setup(
    name="pyproc-worker",
    version="0.1.0",
    description="Python worker for pyproc - Call Python from Go without CGO",
    author="pyproc",
    packages=find_packages(),
    python_requires=">=3.7",
    install_requires=[],
    extras_require={
        "dev": [
            "pytest>=7.0",
            "pytest-asyncio",
            "ruff>=0.1.0",
        ]
    },
)