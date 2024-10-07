### visual code 开发
增加如下调试代码
```
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [        {
        "name": "Launch Package",
        "type": "go",
        "request": "launch",
        "mode": "auto",
        "cwd": "${workspaceFolder}",
        "program": "${workspaceFolder}/main.go",
        "env": {
        },
        "args":[]
    }]
}
```
### 根目录下
conf/tank.json、html/下前端构建的代码


### 正式构建
```
### 本地构建
go build -o tank.exe main.go
### 前端构建后
放置build下html下
在build/pack下，运行build.bat即可

### 镜像构建
docker build -t tank:v1 .
```

### Dockfile_multi 多阶段镜像构建需要docker版本高于17.05才行

admin/123456