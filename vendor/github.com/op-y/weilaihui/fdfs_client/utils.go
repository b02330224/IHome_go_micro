package fdfs_client

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/op-y/weilaihui/goconfig/config"
)

type Errno struct {
	status int
}

func (e Errno) Error() string {
	errmsg := fmt.Sprintf("errno [%d] ", e.status)
	switch e.status {
	case 17:
		errmsg += "File Exist"
	case 22:
		errmsg += "Argument Invlid"
	}
	return errmsg
}

type FdfsConfigParser struct{}

var (
	ConfigFile *config.Config
)

func (this *FdfsConfigParser) Read(filename string) (*config.Config, error) {
	return config.ReadDefault(filename)
}

func fdfsCheckFile(filename string) error {
	if _, err := os.Stat(filename); err != nil {
		return err
	}
	return nil
}

func readCstr(buff io.Reader, length int) (string, error) {
	str := make([]byte, length)
	n, err := buff.Read(str)
	if err != nil || n != len(str) {
		return "", Errno{255}
	}

	for i, v := range str {
		if v == 0 {
			str = str[0:i]
			break
		}
	}
	return string(str), nil
}

func getFileExt(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return ""
}

func splitRemoteFileId(remoteFileId string) ([]string, error) {
	parts := strings.SplitN(remoteFileId, "/", 2)
	if len(parts) < 2 {
		return nil, errors.New("error remoteFileId")
	}
	return parts, nil
}
