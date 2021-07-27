package testutil

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Diff compares two gotestfmt objects by creating a JSON representation and
// then diffing the output. It is not intended to be used outside of gotestfmt
// as it does not support objects of different sizes.
func Diff(o1 interface{}, o2 interface{}) string {
	w1, err := json.MarshalIndent(o1, "", "  ")
	if err != nil {
		panic(err)
	}
	w2, err := json.MarshalIndent(o2, "", "  ")
	if err != nil {
		panic(err)
	}
	if string(w1) == string(w2) {
		return ""
	}
	l1 := strings.Split(string(w1), "\n")
	l2 := strings.Split(string(w2), "\n")
	result := strings.Builder{}
	result.WriteString(fmt.Sprintf("--- expected\n+++ actual\n@@ -1,%d +1,%d @@\n", len(l1), len(l2)))
	for i := 0; i < len(l1); i++ {
		if len(l2) < i + 1 {
			result.WriteString(fmt.Sprintf("-%s\n", l1[i]))
			continue
		}
		if l1[i] != l2[i] {
			result.WriteString(fmt.Sprintf("-%s\n", l1[i]))
			result.WriteString(fmt.Sprintf("+%s\n", l2[i]))
		} else {
			result.WriteString(fmt.Sprintf(" %s\n", l1[i]))
		}
	}
	for i := len(l1); i < len(l2); i++ {
		result.WriteString(fmt.Sprintf("+%s\n", l2[i]))
	}
	return result.String()
}
