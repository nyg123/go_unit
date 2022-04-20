package def

import "encoding/xml"

type Coverage struct {
	XMLName   xml.Name  `xml:"coverage"`
	Generated int64     `xml:"generated,attr"`
	Project   []Project `xml:"project"`
}

type Project struct {
	XMLName xml.Name `xml:"project"`
	File    []File   `xml:"file"`
}

type File struct {
	XMLName xml.Name `xml:"file"`
	Name    string   `xml:"name,attr"`
	Line    []Line   `xml:"line"`
}

type Line struct {
	XMLName xml.Name `xml:"line"`
	Num     int32    `xml:"num,attr"`
	Count   int8     `xml:"count,attr"`
}
