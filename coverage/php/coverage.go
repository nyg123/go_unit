package php

import (
    "encoding/xml"
    "errors"
    "fmt"
    "github.com/nyg123/go_unit/def"
    "io/ioutil"
    "os"
    "strings"
    "time"
)

func GetCoverage(config def.Config) (def.CoverageFmt, error) {
    coverageFmt := make(def.CoverageFmt)
    file, err := os.Open(config.Path + config.CoveragePath)
    if err != nil {
        fmt.Printf("没有覆盖率文件:%v \n", err)
        return nil, err
    }
    defer func(file *os.File) {
        _ = file.Close()
    }(file)
    data, err := ioutil.ReadAll(file)
    if err != nil {
        return nil, err
    }
    coverage := def.PhpCoverage{}
    err = xml.Unmarshal(data, &coverage)
    if err != nil {
        return nil, err
    }
    timeObj := time.Unix(coverage.Generated, 0)

    date := timeObj.Format("2006-01-02 15:04:05")
    fmt.Printf("单元测试运行时间：%s \n", date)
    if len(coverage.Project) < 1 {
        return nil, errors.New("no project")
    }
    project := coverage.Project[0]

    for _, file := range project.File {
        fileName := strings.Replace(file.Name, config.CoveragePrefix, "", 1)
        coverageFmt[fileName] = map[int]bool{}
        for _, line := range file.Line {
            coverageFmt[fileName][int(line.Num)] = line.Count > 0
        }
    }
    return coverageFmt, nil
}
