## 统计一个分支每人开发的代码行数,及单测覆盖率 目前只支持golang php

- 统计某个分支基于基础分支(master)的提交
- 不统计空行，空格的变动
- 可指定需统计的文件扩展名，如php,go
- 可排除不需要统计的文件夹，如vendor
- 代码提交后又删除的，不计入统计

## 使用方法
> go_unit -c unitConf.json

```-c``` 指定配置文件 默认是unitConf.json

执行的结果如下
```txt
提交人:aaa@email.com  变更行数1681    可测试代码行数1051      单元测试覆盖行数933     覆盖率88.77%
提交人:aaa@email.com  变更行数52      可测试代码行数27        单元测试覆盖行数26      覆盖率96.30%
提交人:aaa@email.com  变更行数129     可测试代码行数104       单元测试覆盖行数32      覆盖率30.77%
提交人:aaa@email.com  变更行数1       可测试代码行数1         单元测试覆盖行数1       覆盖率100.00%
            合计：    变更行数1863    可测试代码行数1183      单元测试覆盖行数992     覆盖率83.85%

```

## 配置说明,以go项目为例
```json
{
    "lang" : "go",
    "path" : "./",
    "coverage_path" : "./coverage.out",
    "coverage_prefix" : "go_unit/",
    "diff_commit" : "master",
    "diff_exclude" : [
        "/protobuf/",
        ".pb.go$",
        ".pb.gw.go$",
        "_test.go$"
    ],
    "unit_exclude" : [
        "main.go"
    ],
    "ext" : [
        "go"
    ],
    "show_detail" : true
}
```
- lang: 指定语言，目前支持go,php
- path: 指定项目目录，默认是当前目录
- coverage_path: 指定覆盖率输出文件路径，默认是coverage.out
- coverage_prefix: 指定覆盖率输出文件前缀，覆盖率文件中的文件路径会加的有项目名称，
  例如 `go_unit/main.go:32.9,35.25 1 0` 需要指定前缀为 `go_unit/`
- diff_commit 用当前分支与哪个分支进行比较，一般是与`master`对比
- diff_exclude: 指定不需要统计代码变更的文件，支持正则 例如 `/protobuf/` 一些自动生成的文件、外部包、或者是测试文件我们不需要统计为代码变更
- unit_exclude: 指定不需要统计覆盖率的文件，支持正则 例如 `main.go` 
- ext: 指定统计的文件扩展名 例如 `go`
- show_detail: 是否输出每个人的单测未覆盖的代码详情，默认是true