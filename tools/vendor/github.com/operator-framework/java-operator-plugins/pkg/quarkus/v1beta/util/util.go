// Copyright 2021 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"strings"
)

var (
	wordMapping = map[string]string{
		"http": "HTTP",
		"url":  "URL",
		"ip":   "IP",
	}

	javaKeywords = map[string]int8{
		"abstract":     0,
		"assert":       0,
		"boolean":      0,
		"break":        0,
		"byte":         0,
		"case":         0,
		"catch":        0,
		"char":         0,
		"class":        0,
		"const":        0,
		"continue":     0,
		"default":      0,
		"do":           0,
		"double":       0,
		"else":         0,
		"enum":         0,
		"extends":      0,
		"final":        0,
		"finally":      0,
		"float":        0,
		"for":          0,
		"goto":         0,
		"if":           0,
		"implements":   0,
		"import":       0,
		"instanceof":   0,
		"int":          0,
		"interface":    0,
		"long":         0,
		"native":       0,
		"new":          0,
		"package":      0,
		"private":      0,
		"protected":    0,
		"public":       0,
		"return":       0,
		"short":        0,
		"static":       0,
		"strictfp":     0,
		"super":        0,
		"switch":       0,
		"synchronized": 0,
		"this":         0,
		"throw":        0,
		"throws":       0,
		"transient":    0,
		"try":          0,
		"void":         0,
		"volatile":     0,
		"while":        0,
	}
)

func translateWord(word string, initCase bool) string {
	if val, ok := wordMapping[word]; ok {
		return val
	}
	if initCase {
		return strings.Title(word)
	}
	return word
}

func ReverseDomain(domain string) string {
	s := strings.Split(domain, ".")

	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return strings.Join(s, ".")
}

// Converts a string to CamelCase
func ToCamel(s string) string {
	// s = addWordBoundariesToNumbers(s)
	s = strings.Trim(s, " ")
	n := ""
	bits := []string{}
	for _, v := range s {
		if v == '_' || v == ' ' || v == '-' {
			bits = append(bits, n)
			n = ""
		} else {
			n += string(v)
		}
	}
	bits = append(bits, n)

	ret := ""
	for i, substr := range bits {
		ret += translateWord(substr, i != 0)
	}
	return ret
}

func ToClassname(s string) string {
	return translateWord(ToCamel(s), true)
}

func SanitizeDomain(domain string) string {
	// Split into the domain portions via "."
	domainSplit := strings.Split(domain, ".")

	// Loop through each portion of the domain
	for i, domainPart := range domainSplit {
		// If there are hyphens, replace with underscore
		if strings.Contains(domainPart, "-") {
			fmt.Printf("\ndomain portion (%s) contains hyphens ('-') and needs to be sanitized to create a legal Java package name. Replacing all hyphens with underscores ('_')\n", domainPart)
			domainSplit[i] = strings.ReplaceAll(domainPart, "-", "_")
		}

		// If any portion includes a keyword, replace with keyword_
		if _, ok := javaKeywords[domainPart]; ok {
			fmt.Printf("\ndomain portion (%s) is a Java keyword and needs to be sanitized to create a legal Java package name. Adding an underscore ('_') to the end of the domain portion\n", domainPart)
			domainSplit[i] = domainPart + "_"
		}

		// If any portion starts with number, make it start with underscore
		if domainPart != "" && domainPart[0] >= '0' && domainPart[0] <= '9' {
			fmt.Printf("\ndomain portion(%s) begins with a digit and needs to be sanitized to create a legal Java package name. Adding an underscore('_') to the beginning of the domain portion\n", domainPart)
			domainSplit[i] = "_" + domainPart
		}

	}

	return strings.Join(domainSplit, ".")
}
