import setuptools
from setuptools.command.build_ext import build_ext
import subprocess
import sys
import os
import platform

# --- Configuration ---
PKG_NAME = 'banbta'
GO_PACKAGES = {
    'ta': 'github.com/banbox/banta/python/ta',
    'tav': 'github.com/banbox/banta/python/tav'
}
# --- End Configuration ---

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

        for mod_name, go_pkg_path in GO_PACKAGES.items():
            mod_path = os.path.join(PKG_NAME, mod_name)
            go_lib_name = f"{mod_name}_go"
            
            # Use 'lib' prefix for Unix-like systems
            output_filename = f"lib{go_lib_name}{go_lib_ext}" if sys.platform != 'win32' else f"{go_lib_name}{go_lib_ext}"
            output_filepath = os.path.join(mod_path, output_filename)

            # --- 1. Build Go shared library ---
            go_build_cmd = ['go', 'build', '-buildmode=c-shared', '-o', output_filepath, go_pkg_path]
            print(f"Running command: {' '.join(go_build_cmd)}", flush=True)
            subprocess.check_call(go_build_cmd)

            # --- Windows Specific: Create .lib file for linker ---
            if sys.platform == 'win32':
                import_lib_a = f"{output_filepath}.a"
                import_lib_lib = os.path.join(mod_path, f"{go_lib_name}.lib")
                if os.path.exists(import_lib_a):
                    if os.path.exists(import_lib_lib):
                        os.remove(import_lib_lib)
                    os.rename(import_lib_a, import_lib_lib)
                    print(f"Prepared linker import library: {import_lib_lib}", flush=True)

            # --- 2. Generate C bindings with pybindgen ---
            py_executable = sys.executable
            pybindgen_cmd = [py_executable, os.path.join(mod_path, 'build.py')]
            print(f"Running command: {' '.join(pybindgen_cmd)}", flush=True)
            subprocess.check_call(pybindgen_cmd)

            # --- 3. Windows-specific fix for PyInit ---
            if sys.platform == 'win32':
                c_file_path = os.path.join(mod_path, f'{mod_name}.c')
                with open(c_file_path, 'r', encoding='utf-8') as f: content = f.read()
                if ' __declspec(dllexport) PyInit_' not in content:
                    content = content.replace(' PyInit_', ' __declspec(dllexport) PyInit_')
                    with open(c_file_path, 'w', encoding='utf-8') as f: f.write(content)
                    print(f"Applied Windows PyInit fix for {mod_name}.", flush=True)
        
        super().run()

ext_modules = []
for mod_name in GO_PACKAGES.keys():
    mod_path = os.path.join(PKG_NAME, mod_name)
    ext = setuptools.Extension(
        name=f'{PKG_NAME}.{mod_name}._{mod_name}',
        sources=[os.path.join(mod_path, f'{mod_name}.c')],
        include_dirs=[mod_path],
        library_dirs=[mod_path],
        libraries=[f'{mod_name}_go']
    )
    if platform.system() == "Linux":
        ext.extra_link_args = ["-Wl,-rpath,$ORIGIN"]
    elif platform.system() == "Darwin":
        ext.extra_link_args = ["-Wl,-rpath,@loader_path"]
    ext_modules.append(ext)

# Telling setuptools to include the Go shared libraries
package_data_files = []
if sys.platform == 'win32':
    package_data_files.append('*.dll')
elif sys.platform == 'darwin':
    package_data_files.append('*.dylib')
else:
    package_data_files.append('*.so')

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
    distclass=BinaryDistribution,
    ext_modules=ext_modules,
    cmdclass={'build_ext': CustomBuildExt},
    package_data={
        f'{PKG_NAME}.{mod_name}': package_data_files for mod_name in GO_PACKAGES.keys()
    },
    zip_safe=False,
) 