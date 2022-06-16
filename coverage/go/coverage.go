package _go

import (
    "fmt"
    "github.com/nyg123/go_unit/def"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

// GetCoverage 解析覆盖率文件
func GetCoverage(config def.Config) (def.CoverageFmt, error) {
    coverageFmt := make(def.CoverageFmt)
    file, err := os.Open(config.Path + config.CoveragePath)
    if err != nil {
        fmt.Printf("没有覆盖率文件:%v \n", err)
        return coverageFmt, nil
    }
    defer func(file *os.File) {
        _ = file.Close()
    }(file)
    data, err := ioutil.ReadAll(file)
    if err != nil {
        return nil, err
    }
    dataArr := strings.Split(string(data), "\n")
    tmpMap := map[string]bool{}
re2:
    for _, s := range dataArr { // 去重
        if _, ok := tmpMap[s]; ok {
            continue
        } else {
            tmpMap[s] = true
        }
        s = strings.Replace(s, config.CoveragePrefix, "", 1)
        regName, _ := regexp.Compile("(.*?):")
        if !regName.MatchString(s) {
            continue
        }
        fileName := regName.FindStringSubmatch(s)[1]
        for _, exclude := range config.UnitExclude {
            reg, _ := regexp.Compile(exclude)
            if reg.MatchString(fileName) {
                continue re2
            }
        }
        coverage, ok := coverageFmt[fileName]
        if !ok {
            coverage = make(map[int]bool)
        }
        regLine, _ := regexp.Compile(":(\\d*)\\.\\d*,(\\d*)\\.([\\s\\S]*?)(\\d+)$")
        Line := regLine.FindStringSubmatch(s)
        if len(Line) < 5 {
            continue
        }
        start, _ := strconv.Atoi(Line[1])
        end, _ := strconv.Atoi(Line[2])
        for ; start <= end; start++ {
            b, ok := coverage[start]
            if ok {
                coverage[start] = b || Line[4] != "0"
            } else {
                coverage[start] = Line[4] != "0"
            }
        }
        coverageFmt[fileName] = coverage
    }
    return coverageFmt, nil
}
