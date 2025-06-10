注意，下面是手动编译构建pip包的流程，只支持当前操作系统，无法发布到pypi，需通过github actions自动编译发布。

# 安装
```shell
# 安装Python依赖  
python3 -m pip install pybindgen setuptools wheel
  
# 安装Go工具  
go install golang.org/x/tools/cmd/goimports@latest  
go install github.com/go-python/gopy@latest
```

# 编译
```shell linux
gopy build -output=_out -vm=python3 \
  -name=bbta \
  -dynamic-link=True \
  github.com/banbox/banta/python/ta \
  github.com/banbox/banta/python/tav
```
```shell windows
gopy build -output=_out -vm=python3 ^
  -name=bbta ^
  github.com/banbox/banta/python/ta ^
  github.com/banbox/banta/python/tav
```
注意-dynamic-link=True在linux下编译必须添加，windows下必须移除

然后在根目录下运行python，执行`from _out import ta,tav`即可

## 打包pkg
下面打包是为了上传到pypi准备，本地测试无需打包
```shell linux
gopy pkg -output=_out -vm=python3 \
  -name=bbta \
  -dynamic-link=True \
  --version=0.3.0
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

