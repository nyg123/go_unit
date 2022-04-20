package def

type CoverageFmt map[string]map[int]bool

type AuthorInfo map[string]Info

type Info struct {
	LineNum  int
	NeedTest int
	TestNum  int
}
