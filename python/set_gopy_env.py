import os
import sys

# This script generates a batch file to set GOPY environment variables for Windows.
py_version = sys.version_info
gopy_pylib = f"python{py_version.major}{py_version.minor}"
# sys.prefix points to the Python installation directory being used by cibuildwheel
gopy_libdir = os.path.join(sys.prefix, 'libs')

with open("gopy_env.bat", "w") as f:
    f.write('@echo off\n')
    f.write('echo Setting GOPY environment variables for the current session...\n')
    f.write(f'set "GOPY_PYLIB={gopy_pylib}"\n')
    f.write(f'set "GOPY_LIBDIR={gopy_libdir}"\n')
    f.write('echo Done.\n') 