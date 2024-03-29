package main

import (
	"fmt"
	"github.com/injoyai/goutil/other/upload"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func handlerUploadMinio(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		logs.Err("未选择上传资源")
		return
	}

	f, err := os.Open(args[0])
	if err != nil {
		logs.Err(err)
		return
	}

	i, err := upload.NewMinio(&upload.MinioConfig{
		Endpoint:   flags.GetString("endpoint"),
		AccessKey:  flags.GetString("access"),
		SecretKey:  flags.GetString("secret"),
		BucketName: flags.GetString("bucket"),
		Rename:     flags.GetBool("rename"),
	})
	if err != nil {
		logs.Err(err)
		return
	}

	filename, err := i.Save(filepath.Base(args[0]), f)
	if err != nil {
		logs.Err(err)
		return
	}
	fmt.Println("资源上传成功,地址：", filename)
}
