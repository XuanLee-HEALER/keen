package util

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

// LoginUser 获取当前登录用户
func LoginUser() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}

	return u.Username
}

// Banner 制作title文本
func Banner(title, fill string, num uint) string {
	return fmt.Sprintf(strings.Repeat(fill, int(num)) + " " + title + " " + strings.Repeat(fill, int(num)) + "\n")
}

// MbToByte MB和B的转换，MB格式为 xx.x MB，最后结果向上取整
func MbToByte(mb string) uint64 {
	segs := strings.Split(mb, " ")
	n, _ := strconv.ParseFloat(segs[0], 64)
	n = n * 1024 * 1024
	return uint64(math.Floor(n))
}

// RemoveChildren 删除目录下所有的子项，不包括目录自身
func RemoveChildren(path string) error {
	path = filepath.Join(path, "*")
	matches, err := filepath.Glob(path)
	if err != nil {
		return err
	}

	for _, f := range matches {
		err := os.RemoveAll(f)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateScript 在工作目录创建一个脚本，返回文件名称
func CreateScript(name string, content []byte) (string, error) {
	f, err := os.Create(name)
	if err != nil {
		return "", err
	}
	defer f.Close()

	wr := bufio.NewWriter(f)

	wi, l := 0, len(content)
	for {
		lap := min(1024, l)
		n, err := wr.Write(content[wi : wi+lap])
		if err != nil {
			f.Close()
			os.Remove(name)
			return "", err
		}
		wr.Flush()
		wi += lap
		l -= lap

		if n == 0 {
			break
		}
	}

	return name, nil
}

// PathExists 判断路径是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// ToJsonFile 將Json内容打印到文件
func ToJsonFile(p any, prefix, indent, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	bs, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(f, string(bs))
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}

// ReplaceDirectoryPath 替换路径中的一部分父目录
func ReplaceDirectoryPath(oldVol, oldPath, newVol string) (string, error) {
	rel, err := filepath.Rel(oldVol, oldPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(newVol, rel), nil
}

// CreateDirIfNotExist如果路径不存在则创建目录
func CreateDirIfNotExist(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, os.ModeDir)
		}
	}

	return nil
}

// base64编码
func Base64StdEncode(str string) string {
	return base64.RawStdEncoding.EncodeToString([]byte(str))
}

// base64解码
func Base64StdDecode(str string) (string, error) {
	bs, err := base64.RawStdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	return string(bs), err
}

// IsSubDir 判断sub目录是否为dir目录的子目录
func IsSubDir(dir string, sub string) bool {
	adir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	asub, err := filepath.Abs(sub)
	if err != nil {
		return false
	}
	dir = strings.ToUpper(adir)
	sub = strings.ToUpper(asub)

	return strings.HasPrefix(sub, dir)
}
