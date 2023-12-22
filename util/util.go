package util

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"gitea.fcdm.top/lixuan/keen/fp"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type CleanFunc func()

var TrimSpace = func(r rune) bool { return r == ' ' || r == '\t' }

// ExitWith 以状态码{code}退出程序，执行{clean}函数
func ExitWith(code int, clean CleanFunc) {
	clean()
	os.Exit(code)
}

// ReadListOutput 读取Format-list输出内容，返回数组的第一个元素为header
func ReadListOutput(bs []byte) []map[string]string {
	res := make([]map[string]string, 0)

	scanner := bufio.NewScanner(bytes.NewReader(bs))
	valid := false

	tmap := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			colonIdx := strings.Index(line, ":")
			if colonIdx != -1 {
				k, v := line[:colonIdx], line[colonIdx+1:]
				k, v = strings.TrimFunc(k, TrimSpace), strings.TrimFunc(v, TrimSpace)
				tmap[k] = v
				valid = true
			}
		} else {
			if valid {
				valid = !valid
				res = append(res, tmap)
				tmap = make(map[string]string)
			}
		}
	}

	if len(tmap) > 0 {
		res = append(res, tmap)
	}

	return res
}

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

func xmin(a, b int) int {
	if a < b {
		return a
	}
	return b
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
		lap := xmin(1024, l)
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

// CreateDirAs 指定uid、gid创建目录，此处uid和gid来源必须合法
func CreateDirAs(dir string, uid, gid string) error {
	err := os.MkdirAll(dir, os.ModeDir|0700)
	if err != nil {
		return err
	}

	iuid, _ := strconv.Atoi(uid)
	igid, _ := strconv.Atoi(gid)
	return os.Chown(dir, iuid, igid)
}

// DirSize 计算目录所有文件总大小
func DirSize(root string) (uint64, error) {
	rs, err := IterDirWithVal[uint64](root, func(s string, de fs.DirEntry) bool { return de.IsDir() }, func(s string, de fs.DirEntry) (uint64, error) {
		info, err := de.Info()
		if err != nil {
			return 0, err
		} else {
			return uint64(info.Size()), nil
		}
	})
	if err != nil {
		return 0, err
	}

	r, _ := fp.Reduce[uint64](rs, func(e1, e2 uint64) uint64 { return e1 + e2 })
	return r, nil
}

// UidGid 获取指定用户名对应的UID和GID
func UidGid(username string) (string, string, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return "", "", err
	}

	if u == nil {
		return "", "", fmt.Errorf("user %s is nil", username)
	}

	return u.Uid, u.Gid, nil
}

// IterDir 遍历目录，执行函数
func IterDir(path string, filter func(string, fs.DirEntry) bool, f func(string, fs.DirEntry) error) error {
	return filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filter(path, d) {
			return nil
		}

		err = f(path, d)
		if err != nil {
			return err
		}

		return nil
	})
}

// IterDirWithVal 遍历目录，执行函数，收集函数执行结果并返回
func IterDirWithVal[T any](path string, filter func(string, fs.DirEntry) bool, f func(string, fs.DirEntry) (T, error)) ([]T, error) {
	res := make([]T, 0)

	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filter(path, d) {
			return nil
		}

		v, err := f(path, d)
		if err != nil {
			return err
		}
		res = append(res, v)

		return nil
	})

	return res, err
}

func ConvertGBKToUtf8(str string) (string, error) {
	rd := transform.NewReader(bytes.NewReader([]byte(str)), simplifiedchinese.GBK.NewDecoder())
	d, err := io.ReadAll(rd)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

func ConvertUtf8ToGBK(str string) (string, error) {
	rd := transform.NewReader(bytes.NewReader([]byte(str)), simplifiedchinese.GBK.NewEncoder())
	d, err := io.ReadAll(rd)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

type Less func(i, j any) bool

func Min[T any](eles []T, cmp Less) (min T, find bool) {
	if len(eles) == 0 {
		find = false
	} else if len(eles) == 1 {
		min = eles[0]
		find = true
	} else {
		b := eles[0]
		for _, e := range eles[1:] {
			if !cmp(b, e) {
				b = e
			}
		}
		min = b
		find = true
	}

	return
}

func MinInt(arr []int) (min int, find bool) {
	return Min[int](arr, func(i, j any) bool { return i.(int) < j.(int) })
}

func Max[T any](eles []T, cmp Less) (max T, find bool) {
	if len(eles) == 0 {
		find = false
	} else if len(eles) == 1 {
		max = eles[0]
		find = true
	} else {
		b := eles[0]
		for _, e := range eles[1:] {
			if cmp(b, e) {
				b = e
			}
		}
		max = b
		find = true
	}

	return
}

func MaxInt(arr []int) (min int, find bool) {
	return Max[int](arr, func(i, j any) bool { return i.(int) < j.(int) })
}

func MinMax[T any](eles []T, cmp Less) (min, max T, find bool) {
	if len(eles) == 0 {
		find = false
	} else if len(eles) == 1 {
		min, max = eles[0], eles[0]
		find = true
	} else {
		bmin := eles[0]
		bmax := eles[0]
		for _, e := range eles[1:] {
			if !cmp(e, bmax) {
				bmax = e
			}
			if cmp(e, bmin) {
				bmin = e
			}
		}
		min = bmin
		max = bmax
		find = true
	}

	return
}

func MinMaxInt(arr []int) (min, max int, find bool) {
	return MinMax[int](arr, func(i, j any) bool { return i.(int) < j.(int) })
}

type ErrGroup struct {
	errors []error
}

func NewErrGroup() *ErrGroup {
	return &ErrGroup{
		errors: make([]error, 0),
	}
}

func (eg *ErrGroup) AddErrs(errs ...error) {
	eg.errors = append(eg.errors, errs...)
}

func (eg *ErrGroup) IsNil() bool {
	return len(eg.errors) == 0
}

func (eg *ErrGroup) Error() string {
	sb := strings.Builder{}
	for i, e := range eg.errors {
		sb.WriteString(fmt.Sprintf("err_idx(%d) - err(%v)\n", i, e))
	}
	return sb.String()
}
