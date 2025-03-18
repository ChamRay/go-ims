package main

import (
	"runtime"
	"strings"
)

func TrimNewLineWithoutSpace(b string) string {
	osType := runtime.GOOS
	if osType == "windows" {
		return strings.TrimSpace(strings.TrimSuffix(b, "\r\n"))
	} else if osType == "linux" || osType == "darwin" {
		return strings.TrimSpace(strings.TrimSuffix(b, "\n"))
	}
	return strings.TrimSpace(strings.TrimSuffix(b, "\n"))
}

func getNewLine() string {
	osType := runtime.GOOS
	if osType == "windows" {
		return "\r\n"
	}
	return "\n"
}
