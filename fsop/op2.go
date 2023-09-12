// go:build go1.20
package fsop

import (
	"math/rand"
	"os"
)

// CreateRandomFile 创建随机内容的文件
func CreateRandomFile(path string, size FileSize) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	defer f.Sync()

	var ruler int
	unit := make([]byte, WRITE_UNIT)
	for i := 0; i < int(size); i++ {
		ruler = i % int(WRITE_UNIT)
		unit[ruler] = byte(rand.Intn(0xff))
		if ruler == int(WRITE_UNIT)-1 {
			_, err := f.Write(unit)
			if err != nil {
				return err
			}
			clear(unit)
		}
	}

	if ruler != int(WRITE_UNIT)-1 {
		_, err := f.Write(unit[:ruler+1])
		if err != nil {
			return err
		}
	}

	return nil
}
