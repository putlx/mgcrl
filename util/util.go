package util

import (
	"bytes"
	"io"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/robertkrimen/otto"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func EvalJavaScript(script string) string {
	val, err := otto.New().Run(script)
	if err != nil {
		panic(err)
	}
	s, err := val.ToString()
	if err != nil {
		panic(err)
	}
	return s
}

func GbkToUtf8(s string) string {
	reader := transform.NewReader(bytes.NewReader([]byte(s)), simplifiedchinese.GBK.NewDecoder())
	data, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func Reverse(a interface{}) {
	n := reflect.ValueOf(a).Len()
	swap := reflect.Swapper(a)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func AsFileBasename(name string) string {
	return strings.Map(func(c rune) rune {
		if strings.ContainsRune(`"\/:*?<>|`, c) {
			return ' '
		}
		return c
	}, name)
}

var imageTypes = map[string]struct{}{
	".tiff": {},
	".bmp":  {},
	".gif":  {},
	".svg":  {},
	".png":  {},
	".jpeg": {},
	".jpg":  {},
	".webp": {},
	".ico":  {},
}

func IsImageFile(name string) bool {
	_, exists := imageTypes[strings.ToLower(filepath.Ext(name))]
	return exists
}
