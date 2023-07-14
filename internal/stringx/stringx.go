/*
Copyright © 2023 Xu Wu <ixw1991@126.com>
Use of this source code is governed by a MIT style
license that can be found in the LICENSE file.
*/
package stringx

import (
	"strings"
	"unicode"
)

// StringInSlice 函数检查字符串 a 是否在列表 list 中
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// ToPascalCase 函数将一个由'-'连接的字符串转换为 PascalCase
// 例如，"hello-world" 转换为 "HelloWorld"
func ToPascalCase(str string) string {
	splitted := strings.FieldsFunc(str, func(r rune) bool { return r == '-' })
	for i := 0; i < len(splitted); i++ {
		runes := []rune(splitted[i])
		runes[0] = unicode.ToUpper(runes[0])
		splitted[i] = string(runes)
	}
	return strings.Join(splitted, "")
}

// SkipFirstPart 函数用于跳过以 "-" 分隔的字符串的第一部分
// 如果输入的字符串s包含 "-"，例如："hello-world-android"，函数会跳过第一个 "-" 之前的部分（即"hello"），并返回剩余部分转换为驼峰命名的字符串，如："worldAndroid"
// 如果输入的字符串s不包含 "-"，函数会原样返回输入的字符串s
func SkipFirstPart(s string) string {
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		parts = parts[1:]
		return ToCamelCase(strings.Join(parts, "-"))
	}
	return s
}

// SkipLastPart 函数跳过以 '-' 分隔的字符串的最后一部分
// 例如，"hello-world-android" 转换为 "hello-world"
// 如果字符串没有 '-'，例如 "tope"，则返回原字符串
func SkipLastPart(s string) string {
	parts := strings.Split(s, "-")
	if len(parts) < 2 {
		// 如果没有 '-'，返回原字符串
		return s
	}

	// 返回除最后一部分之外的所有部分
	return strings.Join(parts[:len(parts)-1], "-")
}

// SkipFirstAndLastParts 去掉输入字符串的第一部分和最后一部分，
// 如果只有一个 '-'，则只去掉第一部分。
// 例如，"hello-world-golang" 转换为 "world"，
// "hello-world" 转换为 "world"，
// "hello" 保持不变。
func SkipFirstAndLastParts(s string) string {
	parts := strings.Split(s, "-")
	if len(parts) < 3 {
		if len(parts) == 2 {
			return parts[1]
		}
		return s
	}
	return strings.Join(parts[1:len(parts)-1], "-")
}

// ToCamelCase 函数将一个由'-'连接的字符串转换为 camelCase
// 例如，"hello-world" 转换为 "helloWorld"
func ToCamelCase(s string) string {
	s = ToPascalCase(s)
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
