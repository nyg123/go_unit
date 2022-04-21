package main

import (
    "bytes"
    "flag"
    "fmt"
    "github.com/jinzhu/configor"
    "github.com/nyg123/go_unit/def"
    "io/ioutil"
    "os"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
    "sync"
    "time"
)

var configPath = flag.String("c", "unitConf.json", "配置文件")

var Config = def.Config{}

var needToTest = map[string][]string{}
var local sync.Mutex

func main() {
    err := configor.Load(&Config, *configPath)
    if err != nil {
        panic(err)
    }
    coverage, err := getCoverage()
    if err != nil {
        fmt.Printf("error:%v", err)
        return
    }
    var all []def.AuthorInfo
    diffFmt := diff()
    blameChan := make(chan def.AuthorInfo, 1)
    wg := &sync.WaitGroup{}
    wg.Add(len(diffFmt))
    for fileName, line := range diffFmt {
        if len(line) == 0 {
            wg.Done()
            continue
        }
        c := coverage[fileName]
        go blame(fileName, line, c, wg, blameChan)
    }
    // 定义等待信号
    go func() {
        wg.Wait()
        close(blameChan)
    }()
    for item := range blameChan {
        all = append(all, item)
    }
    allAuthorInfo := make(def.AuthorInfo)
    for _, authorInfo := range all {
        for email, info := range authorInfo {
            allInfo := allAuthorInfo[email]
            allInfo.LineNum += info.LineNum
            allInfo.NeedTest += info.NeedTest
            allInfo.TestNum += info.TestNum
            allAuthorInfo[email] = allInfo
        }
    }
    var lineNum = 0
    var needTest = 0
    var testNum = 0
    for email, info := range allAuthorInfo {
        c := 0.0
        if info.NeedTest > 0 {
            c = float64(info.TestNum) * 100 / float64(info.NeedTest)
        }
        lineNum += info.LineNum
        needTest += info.NeedTest
        testNum += info.TestNum
        fmt.Printf(
            "提交人:%s\t变更行数%d\t可测试代码行数%d\t单元测试覆盖行数%d\t覆盖率%.2f%% \n", email, info.LineNum, info.NeedTest, info.TestNum,
            c,
        )
    }
    c := 0.0
    if needTest > 0 {
        c = float64(testNum) * 100 / float64(needTest)
    }
    fmt.Printf(
        "                    合计：\t变更行数%d\t可测试代码行数%d\t单元测试覆盖行数%d\t覆盖率%.2f%% \n", lineNum, needTest, testNum, c,
    )
    if !Config.ShowDetail {
        return
    }
    t := time.Now().Format("0102150405")
    for email, content := range needToTest {
        f, _ := os.OpenFile(t+"_"+email+".log", os.O_WRONLY|os.O_CREATE, 0)
        for _, s := range content {
            _, _ = f.WriteString(s)
        }
        _ = f.Close()
    }
}

// 获取解析覆盖率
func getCoverage() (def.CoverageFmt, error) {
    file, err := os.Open(Config.CoveragePath)
    if err != nil {
        return nil, err
    }
    defer func(file *os.File) {
        _ = file.Close()
    }(file)
    data, err := ioutil.ReadAll(file)
    if err != nil {
        return nil, err
    }
    coverageFmt := make(def.CoverageFmt)
    dataArr := strings.Split(string(data), "\n")
    tmp_map := map[string]bool{}
re2:
    for _, s := range dataArr { // 去重
        if _, ok := tmp_map[s]; ok {
            continue
        } else {
            tmp_map[s] = true
        }
        s = strings.Replace(s, "editor_go", "", 1)
        regName, _ := regexp.Compile("^/(.*?):")
        if !regName.MatchString(s) {
            continue
        }
        fileName := regName.FindStringSubmatch(s)[1]
        for _, exclude := range Config.UnitExclude {
            reg, _ := regexp.Compile(exclude)
            if reg.MatchString(fileName) {
                continue re2
            }
        }
        coverage, ok := coverageFmt[fileName]
        if !ok {
            coverage = make(map[int]bool)
        }
        regLine, _ := regexp.Compile(":(\\d*)\\.\\d*,(\\d*)\\.([\\s\\S]*)(\\d*?)$")
        Line := regLine.FindStringSubmatch(s)
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

// 获取git变更记录
func diff() map[string][]int32 {
    var stdout bytes.Buffer
    cmd := exec.Command(
        "bash", "-c",
        "cd "+Config.Path+" &  git diff "+Config.DiffCommit+" -U0 -w --ignore-all-space --ignore-blank-lines",
    )
    cmd.Stdout = &stdout
    _ = cmd.Run()
    outList := strings.Split(stdout.String(), "\n")
    regName, _ := regexp.Compile("diff --git([\\s\\S]*)\\sb/(.*?)$")
    regLine, _ := regexp.Compile("@@([\\s\\S]*)\\+(\\d*),?(\\d*) @@")
    fileName := ""
    diffFmt := make(map[string][]int32)
re:
    for _, str := range outList {
        if regName.MatchString(str) {
            fileName = regName.FindStringSubmatch(str)[2]
            diffFmt[fileName] = []int32{}
        }
        if len(fileName) < 2 {
            continue
        }
        if fileName[len(fileName)-2:] != "go" {
            continue
        }
        for _, exclude := range Config.DiffExclude {
            reg, _ := regexp.Compile(exclude)
            if reg.MatchString(fileName) {
                continue re
            }
        }
        if regLine.MatchString(str) {
            match := regLine.FindStringSubmatch(str)
            star, _ := strconv.Atoi(match[2])
            step := 1
            if match[3] != "" {
                step, _ = strconv.Atoi(match[3])
            }
            for step > 0 {
                step--
                diffFmt[fileName] = append(diffFmt[fileName], int32(star))
                star++
            }
        }
    }
    return diffFmt
}

// 获取git变更责任人，并组装覆盖率
func blame(
    file string,
    line []int32,
    coverage map[int]bool,
    wg *sync.WaitGroup,
    ch chan def.AuthorInfo,
) {
    var stdout bytes.Buffer
    cmd := exec.Command("bash", "-c", "cd "+Config.Path+" &  git blame -e -w "+file)
    cmd.Stdout = &stdout
    _ = cmd.Run()
    infoMap := make(def.AuthorInfo)
    outList := strings.Split(stdout.String(), "\n")
    reg, _ := regexp.Compile("\\(<(.*?)>([\\s\\S]*)\\+0800\\s*(\\d*?)\\)\\s")
    for _, out := range outList {
        if out == "" {
            continue
        }
        match := reg.FindStringSubmatch(out)
        if len(match) < 3 || match[3] == "" {
            fmt.Println(file)
            continue
        }
        curLine, _ := strconv.Atoi(match[3])
        for _, l := range line {
            if l == int32(curLine) {
                info := infoMap[match[1]]
                info.LineNum++
                c, ok := coverage[curLine]
                if ok {
                    info.NeedTest++
                    if c {
                        info.TestNum++
                    } else {
                        local.Lock()
                        needToTest[match[1]] = append(
                            needToTest[match[1]], fmt.Sprintf("%s:%d \n", file, curLine),
                        )
                        local.Unlock()
                    }
                }
                infoMap[match[1]] = info
                break
            }
        }
    }
    ch <- infoMap
    wg.Done()
}
