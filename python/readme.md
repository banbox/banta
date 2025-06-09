# 安装
```shell
# 安装Python依赖  
python3 -m pip install pybindgen setuptools wheel
  
# 安装Go工具  
go install golang.org/x/tools/cmd/goimports@latest  
go install github.com/go-python/gopy@latest
```

# 编译
```shell
gopy pkg -output=_out -vm=python3 \
  -name=bbta \
  -version=0.3.0 \
  -author="banbot" \
  -email="banbot@163.com" \
  -desc="python bindings for banta" \
  -url="https://github.com/banbox/banta" \
  github.com/banbox/banta/python/ta \
  github.com/banbox/banta/python/tav
```

# 构建
```shell
cd _out
make install
```
# 发布
```shell
# 构建分发包  
python3 setup.py sdist bdist_wheel
  
# 上传到PyPI  
python3 -m pip install twine
twine upload dist/*
```

