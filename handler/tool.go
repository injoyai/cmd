package handler

import (
	"fmt"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"os"
	"path/filepath"
	"strings"
)

func _merge_ts(dir, output string) error {

	//判断ffmpeg是否下载
	resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:    "ffmpeg",
		Dir:         oss.ExecDir(),
		ProxyEnable: true,
		//ProxyAddress: flags.GetString("proxy"),
	})

	lsFilename := filepath.Join(dir, "ts_list.txt")
	lsFilename = strings.ReplaceAll(lsFilename, "\\", "/")
	fmt.Println("生成TS列表文件:", lsFilename)
	file, err := os.Create(lsFilename)
	if err != nil {
		return err
	}
	defer os.Remove(lsFilename)
	defer file.Close()

	err = oss.RangeFileInfo(dir, func(info *oss.FileInfo) (bool, error) {
		if strings.HasSuffix(info.Name(), ".ts") {
			_, err = file.WriteString("file '" + info.Name() + "'\r\n")
		}
		return true, err
	})
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("ffmpeg -f concat -i %s -c copy %s", lsFilename, output)
	fmt.Println("执行ffmpeg命令:", cmd)
	return shell.Run(cmd)
}
