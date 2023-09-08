// go:build windows

package pvd

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)

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

// Listening 根据PID查询监听的TCP/IP端口号
// func Listening(pid uint) (uint, error) {
// 	const (
// 		LISTEN_STATE = 2
// 		ALL_IP       = "0.0.0.0"
// 	)

// 	type TcpInfo struct {
// 		LocalAddress string `json:"LocalAddress"`
// 		LocalPort    int    `json:"LocalPort"`
// 		State        int    `json:"State"`
// 	}

// 	var tcpInfos []TcpInfo
// 	err := PSRetrieve(fmt.Sprintf(GET_LISTEN_TCP_PORT_CMD, pid), &tcpInfos)
// 	if err != nil {
// 		return 0, err
// 	}

// 	for _, tcpInfo := range tcpInfos {
// 		if (tcpInfo.LocalAddress == ALL_IP) && tcpInfo.State == LISTEN_STATE {
// 			return uint(tcpInfo.LocalPort), nil
// 		}
// 	}

// 	return 0, errors.New("failed to find any of valid listening port")
// }

type CpErr struct {
	SrcPath string
	DstPath string
	error
}

func (e CpErr) Error() string {
	return fmt.Sprintf("copy error, from %s to %s failed: %v", e.SrcPath, e.DstPath, e.error)
}

func (e CpErr) IsSpaceErr() bool {
	if v, ok := e.error.(*fs.PathError); ok {
		if v.Err.Error() == "There is not enough space on the disk." {
			return true
		}
	}
	return false
}

// ParaCpFiles 并发拷贝文件，将多个文件拷贝到一个路径下
// func ParaCpFiles(dsts []string, srcs []string) chan CpErr {
// 	ch := make(chan CpErr, 1)
// 	defer close(ch)

// 	dstFiles, srcFiles := make([]*os.File, 0), make([]*os.File, 0)
// 	for i := range srcs {
// 		srcP, dstP := srcs[i], dsts[i]
// 		rd, err := os.OpenFile(srcP, os.O_RDONLY, 0777)
// 		if err != nil {
// 			ch <- CpErr{srcP, dstP, err}
// 			return ch
// 		}

// 		wt, err := os.Create(dstP)
// 		if err != nil {
// 			ch <- CpErr{srcP, dstP, err}
// 			return ch
// 		}

// 		dstFiles = append(dstFiles, wt)
// 		srcFiles = append(srcFiles, rd)
// 	}

// 	err := yfile.RCopy(dstFiles, srcFiles)
// 	if err.Err != nil {
// 		switch err.Op {
// 		case "read":
// 			ch <- CpErr{SrcPath: err.FName, error: err.Err}
// 		case "write":
// 			ch <- CpErr{DstPath: err.FName, error: err.Err}
// 		}
// 	}

// 	return ch
// }

// RemoveChildren 删除目录下所有的子项
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

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func StringRegVal(regK registry.Key, path string, access uint32, ks ...string) ([]string, error) {
	res := make([]string, 0)

	key, err := registry.OpenKey(regK, path, access)
	if err != nil {
		return nil, fmt.Errorf("failed to open registry key [%s]: %v", path, err)
	}
	defer key.Close()

	for _, k := range ks {
		v, _, err := key.GetStringValue(k)
		if err != nil {
			return nil, fmt.Errorf("failed to find specific key %s: %v", k, err)
		}
		res = append(res, v)
	}

	return res, nil
}

// MakeUp log helper
// func MakeUp(l, f string, v ...any) string {
// 	return fmt.Sprintf("%s %s %s\r\n", l, ylog.CurMod(), fmt.Sprintf(f, v...))
// }
