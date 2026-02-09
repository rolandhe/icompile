package common

import (
	"fmt"
	"go/format"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	strSep = "_"
)

const (
	ServiceFileSuffix = ".idl"
)

// Capitalize 首字母大写
func Capitalize(s string) string {
	return FormatVariable(s, true)
}

func FormatVariable(v string, isTitleCase bool) string {
	i := 0
	if !isTitleCase {
		i = 1
		if v != "" {
			v = string(unicode.ToLower(rune(v[0]))) + v[1:]
		}
	}
	words := strings.Split(v, strSep)
	for ; i < len(words); i++ {
		words[i] = strings.Title(words[i])
	}
	return strings.Join(words, "")
}

const (
	defaultDirPermission  = 0755
	defaultFilePermission = 0644
)

func IsExist(v string) bool {
	_, err := os.Stat(v)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func CreateDir(dirName string) error {
	if IsExist(dirName) {
		return nil
	}
	err := os.MkdirAll(dirName, defaultDirPermission)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirName, err)
	}
	return nil
}

func FormatDir(dir string) string {
	if dir == "" {
		return ""
	}

	dir = GetUserHomeDir(dir)

	if string(dir[len(dir)-1]) != "/" {
		return dir + "/"
	}
	return dir
}

func GetFileNameFromPath(path string) string {
	// 获取文件名（不带扩展名）
	fileName := filepath.Base(path)
	fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
	return fileName
}

func GetUserHomeDir(path string) string {
	// 替换路径中的波浪线（~）为家目录路径
	if strings.HasPrefix(path, "~/") {
		usr, err := user.Current()
		if err != nil {
			return path
		}
		homeDir := usr.HomeDir
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

func FormatGoFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// 格式化go代码
	formatted, err := format.Source(content)
	if err != nil {
		return fmt.Errorf("failed to format Go code in %s: %w", filename, err)
	}

	// 将格式化后的代码写回原文件
	err = os.WriteFile(filename, formatted, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}
	return nil
}
