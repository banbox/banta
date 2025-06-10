import setuptools
from setuptools.command.build_ext import build_ext
import subprocess
import sys
import os
import platform
import json
import re

# --- Configuration ---
PKG_NAME = 'bbta'
GO_PACKAGES = {
    'ta': 'github.com/banbox/banta/python/ta',
    'tav': 'github.com/banbox/banta/python/tav'
}
# --- End Configuration ---

def normalize(name):  # https://peps.python.org/pep-0503/#normalized-names
    return re.sub(r"[-_.]+", "-", name).lower()

class BinaryDistribution(setuptools.Distribution):
    """
    This is a custom distribution class that tells setuptools that this is a
    binary distribution.
    """
    def has_ext_modules(self):
        return True

class CustomBuildExt(build_ext):
    """
    Custom build_ext command to compile the Go part of the extension using gopy.
    """
    def build_extension(self, ext: setuptools.Extension):
        # This is the directory where the package will be installed
        # e.g. build/lib.linux-x86_64-3.9
        build_lib = self.get_ext_fullpath(ext.name)
        pkg_dir = os.path.dirname(build_lib)
        pkg_name = ext.name.split('.')[0]  # 'bbta'
        output_dir = os.path.join(pkg_dir, pkg_name)

        # Ensure the target directory exists
        os.makedirs(output_dir, exist_ok=True)

        go_packages = ext.sources
        py_executable = sys.executable

        cmd = [
            "gopy",
            "build",
            "-no-make",
        ]
        if platform.system() != "Windows":
            cmd.append("-dynamic-link=True")
        
        cmd.extend([
            "-output",
            output_dir,
            "-name",
            pkg_name,
            "-vm",
            py_executable,
        ])
        cmd.extend(go_packages)

        print(f"--- Compiling Go packages using gopy: {' '.join(cmd)} ---", flush=True)

        # Setup environment for gopy
        go_path = subprocess.check_output(["go", "env", "GOPATH"]).decode("utf-8").strip()
        env = os.environ.copy()
        env["PATH"] = f'{env.get("PATH", "")}{os.pathsep}{os.path.join(go_path, "bin")}'
        
        go_env_str = subprocess.check_output(["go", "env", "-json"]).decode("utf-8").strip()
        if go_env_str:
            go_env = json.loads(go_env_str)
            env.update(go_env)

        env["CGO_ENABLED"] = "1"
        env["CGO_LDFLAGS_ALLOW"] = ".*"

        ld_flags = []
        if platform.system() == "Linux":
            ld_flags.append("-Wl,-rpath,$ORIGIN")
        elif platform.system() == "Darwin":
            ld_flags.append("-Wl,-rpath,@loader_path")
        
        if ld_flags:
            # Important: handle existing CGO_LDFLAGS
            env["CGO_LDFLAGS"] = f"{env.get('CGO_LDFLAGS', '')} {' '.join(ld_flags)}".strip()
        
        subprocess.check_call(cmd, env=env)

# --- Package Configuration ---

EXT_MODULES = []
PACKAGE_DATA = {}

# Define the Extension for setuptools.
# The sources are the Go packages to be compiled.
ext = setuptools.Extension(
    name=PKG_NAME,
    sources=list(GO_PACKAGES.values())
)

EXT_MODULES.append(ext)
    
# Specify which shared libraries to package with the wheel.
lib_patterns = []
if platform.system() == 'Windows':
    lib_patterns.append('*.dll')
elif platform.system() == 'Darwin':
    lib_patterns.append('*.dylib')
else:
    lib_patterns.append('*.so')
    
# Add the patterns to the main package's data, so the Go shared library is included.
PACKAGE_DATA[PKG_NAME] = lib_patterns

pkg_version = os.environ.get("PACKAGE_VERSION", "0.3.2")
print(f"pkg_version: {pkg_version}")

setuptools.setup(
    name=normalize(PKG_NAME),
    version=pkg_version,  # 版本号由CI根据tag自动设置
    author="banbot",
    author_email="banbot@163.com",
    description="python bindings for banta",
    long_description=open("readme.md", "r", encoding='utf-8').read(),
    long_description_content_type="text/markdown",
    url="https://github.com/banbox/banta",
    packages=setuptools.find_packages(),
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: BSD License",
        "Operating System :: OS Independent",
    ],
    distclass=BinaryDistribution,
    ext_modules=EXT_MODULES,
    cmdclass={'build_ext': CustomBuildExt},
    package_data=PACKAGE_DATA,
    zip_safe=False,
) 
