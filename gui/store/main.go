package main

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/goutil/frame/in"
	"github.com/injoyai/io"
	"net/http"
	"os"
	"path/filepath"
)

func init() {
	//cfg.Init(cfg.WithYaml("./config/config.yaml"))
}

func main() {

	dir := cfg.GetString("dir", "./upload/")
	port := cfg.GetInt("port", 8080)

	http.ListenAndServe(fmt.Sprintf(":%d", port), in.InitGo(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {
		//case "/upload":
		//	uploadFile, header, err := r.FormFile("file")
		//	in.CheckErr(err)
		//	defer uploadFile.Close()
		//
		//	localFile, err := os.Create(header.Filename + ".upload")
		//	in.CheckErr(err)
		//	defer localFile.Close()
		//
		//	_, err = io.Copy(localFile, uploadFile)
		//	in.CheckErr(err)
		//
		//	err = os.Rename(header.Filename+".upload", header.Filename)
		//	in.Err(err)

		case "/download":

			name := r.URL.Query().Get("name")

			filename := filepath.Join(dir, name)

			f, err := os.Open(filename)
			if err != nil {
				in.Return(http.StatusForbidden, nil)
			}
			defer f.Close()

			info, err := f.Stat()
			if err != nil {
				in.Return(http.StatusForbidden, nil)
			}

			w.Header().Set("Content-Disposition", "attachment; filename="+name)
			w.Header().Set("Content-Length", conv.String(info.Size))
			io.Copy(w, f)

		}

	})))

}
