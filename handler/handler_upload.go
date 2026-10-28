package handler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/injoyai/goutil/other/upload"
	"github.com/spf13/cobra"
)

func UploadMinio(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("未选择上传资源")
	}

	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer f.Close()

	i, err := upload.NewMinio(&upload.MinioConfig{
		Endpoint:   flags.GetString("endpoint"),
		AccessKey:  flags.GetString("access"),
		SecretKey:  flags.GetString("secret"),
		BucketName: flags.GetString("bucket"),
		Rename:     flags.GetBool("rename"),
	})
	if err != nil {
		return err
	}

	filename, err := i.Save(filepath.Base(args[0]), f)
	if err != nil {
		return err
	}
	fmt.Println("资源上传成功,地址：", filename)

	return nil
}
