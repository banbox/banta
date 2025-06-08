import setuptools
from setuptools.command.build_ext import build_ext
import subprocess
import sys
import os
import platform

class BinaryDistribution(setuptools.Distribution):
    """
    This is a custom distribution class that tells setuptools that this is a
    binary distribution.
    """
    def has_ext_modules(self):
        return True

class CustomBuildExt(build_ext):
    """
    Custom build_ext command to compile the Go and C parts of the extension.
    """
    def run(self):
        # Determine the shared library extension based on the platform
        if sys.platform == 'win32':
            go_lib_ext = '.dll'
        elif sys.platform == 'darwin':
            go_lib_ext = '.dylib'
        else:
            go_lib_ext = '.so'

        # --- 1. Build the Go shared library ---
        go_build_cmd = [
            'go', 'build',
            '-buildmode=c-shared',
            '-o', f'ta/ta_go{go_lib_ext}',
            './ta/ta.go'
        ]
        print(f"Running command: {' '.join(go_build_cmd)}", flush=True)
        subprocess.check_call(go_build_cmd)

        # --- 2. Generate C bindings with pybindgen ---
        # This script generates ta/ta.c, which is the CPython wrapper.
        py_executable = sys.executable
        pybindgen_cmd = [py_executable, 'ta/build.py']
        print(f"Running command: {' '.join(pybindgen_cmd)}", flush=True)
        subprocess.check_call(pybindgen_cmd)

        # --- 3. Apply a Windows-specific fix for PyInit ---
        # This replicates the 'sed' command from the original Makefile.
        if sys.platform == 'win32':
            c_file_path = os.path.join('ta', 'ta.c')
            with open(c_file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            content = content.replace(' PyInit_', ' __declspec(dllexport) PyInit_')
            with open(c_file_path, 'w', encoding='utf-8') as f:
                f.write(content)
            print("Applied Windows-specific PyInit fix.", flush=True)

        # --- 4. Proceed with the standard C extension compilation ---
        # Setuptools will now compile ta/ta.c and link it against the Go library.
        super().run()

# Define the C extension module for setuptools
ext_modules = [
    setuptools.Extension(
        name='ta._ta',
        sources=['ta/ta.c'],
        include_dirs=['ta'],
        library_dirs=['ta'],
        # The name of the go library is 'ta_go' (without 'lib' prefix on Linux/macOS)
        libraries=['ta_go']
    )
]

# Add platform-specific linker arguments for rpath.
# This ensures the Python extension can find the Go shared library at runtime.
if platform.system() == "Linux":
    ext_modules[0].extra_link_args = ["-Wl,-rpath,$ORIGIN"]
elif platform.system() == "Darwin":
    ext_modules[0].extra_link_args = ["-Wl,-rpath,@loader_path"]


setuptools.setup(
    name="banbta",
    version="0.3.0",
    author="banbot",
    author_email="banbot@163.com",
    description="python bindings for banta",
    long_description=open("README.md", "r", encoding='utf-8').read(),
    long_description_content_type="text/markdown",
    url="https://github.com/banbox/banta",
    packages=setuptools.find_packages(),
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: BSD License",
        "Operating System :: OS Independent",
    ],
    include_package_data=True,
    distclass=BinaryDistribution,
    ext_modules=ext_modules,
    cmdclass={
        'build_ext': CustomBuildExt,
    },
) 