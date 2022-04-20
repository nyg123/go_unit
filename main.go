package main

import (
	"bytes"
	"fmt"
	_type "github.com/nyg123/go_unit/type"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DiffExclude 代码变更排除文件夹
var DiffExclude = []string{
	"/protobuf/",
	".pb.go$",
	".pb.gw.go$",
	"_test.go$",
}

// UnitExclude 单元测试排除文件
var UnitExclude = []string{
	"app/editor/start.go",
	"wire_gen.go",
}

var needToTest = map[string][]string{}
var local sync.Mutex

func main() {
	path := "D:\\www\\editor_go"
	coverage, err := getCoverage(path)
	if err != nil {
		fmt.Printf("error:%v", err)
		return
	}
	var all []_type.AuthorInfo
	diffFmt := diff(path)
	blameChan := make(chan _type.AuthorInfo, 1)
	wg := &sync.WaitGroup{}
	wg.Add(len(diffFmt))
	for fileName, line := range diffFmt {
		if len(line) == 0 {
			wg.Done()
			continue
		}
		c := coverage[fileName]
		go blame(path, fileName, line, c, wg, blameChan)
	}
	// 定义等待信号
	go func() {
		wg.Wait()
		close(blameChan)
	}()
	for item := range blameChan {
		all = append(all, item)
	}
	allAuthorInfo := make(_type.AuthorInfo)
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
func getCoverage(path string) (_type.CoverageFmt, error) {
	file, err := os.Open(path + "/coverage.out")
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
	coverageFmt := make(_type.CoverageFmt)
	dataArr := strings.Split(string(data), "\n")
re2:
	for _, s := range dataArr {
		s = strings.Replace(s, "editor_go", "", 1)
		regName, _ := regexp.Compile("^/(.*?):")
		if !regName.MatchString(s) {
			continue
		}
		fileName := regName.FindStringSubmatch(s)[1]
		for _, exclude := range UnitExclude {
			reg, _ := regexp.Compile(exclude)
			if reg.MatchString(fileName) {
				continue re2
			}
		}
		coverage, ok := coverageFmt[fileName]
		if !ok {
			coverage = make(map[int]bool)
		}
		regLine, _ := regexp.Compile(":(\\d*)\\.\\d*,(\\d*)\\.([\\s\\S]*)(\\d)$")
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
func diff(path string) map[string][]int32 {
	var stdout bytes.Buffer
	cmd := exec.Command(
		"cmd", "/C",
		"cd "+path+" &  git diff 590a3ff7914300945aa59ef0ef7afd5e2758db25 -U0 -w --ignore-all-space --ignore-blank-lines",
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
		if fileName[len(fileName)-2:] != "go" {
			continue
		}
		for _, exclude := range DiffExclude {
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
	path string,
	file string,
	line []int32,
	coverage map[int]bool,
	wg *sync.WaitGroup,
	ch chan _type.AuthorInfo,
) {
	var stdout bytes.Buffer
	cmd := exec.Command("cmd", "/C", "cd "+path+" &  git blame -e -w "+file)
	cmd.Stdout = &stdout
	_ = cmd.Run()
	infoMap := make(_type.AuthorInfo)
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
