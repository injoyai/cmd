package handler

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/other/excel"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

//go:embed chart.html
var html string

func Chart(cmd *cobra.Command, args []string, flags *Flags) {
	logs.SetFormatterWithDefault()
	if len(args) == 0 {
		logs.Err("无效路径")
		return
	}

	f, err := os.Open(args[0])
	logs.PanicErr(err)
	defer f.Close()

	var result = map[string][][]string{}

	switch filepath.Ext(args[0]) {
	case ".csv":
		r := csv.NewReader(f)
		r.FieldsPerRecord = -1
		result["Sheet1"], _ = r.ReadAll()

	default:
		result, err = excel.FromExcel(f)

	}
	if err != nil {
		logs.Err(err)
		return
	}

	width := flags.GetInt("width", 700)
	height := flags.GetInt("height", 400)
	x, y := excel.ToInt(flags.GetString("label", "A1"))

	lorca.Run(&lorca.Config{
		Width:  width,
		Height: height,
		Html:   html,
	}, func(app lorca.APP) error {

		labels := []string(nil)
		m := make(map[int][]interface{})
		names := make(map[int]string)

		for _, page := range result {
			for line, rows := range page {
				if line == y-1 {
					//这行是标题

					for i, label := range rows {
						if i != x-1 {
							m[i] = []interface{}(nil)
							names[i] = label
						}
					}
					continue
				}
				for i, v := range rows {
					if i < len(rows) && i != x-1 {
						v = strings.Trim(v, "\t")
						v = strings.Trim(v, " ")
						m[i] = append(m[i], v)
					} else if i == x-1 {
						labels = append(labels, v)
					}
				}
			}
		}

		colors := []string{"rgba(75, 192, 192)", "rgba(192, 75, 75)", "rgb(255, 99, 132)", "rgb(54, 162, 235)", "rgb(255, 206, 86)", "rgb(75, 192, 192)", "rgb(153, 102, 255)", "rgb(255, 159, 64)"}

		datasets := []g.Map(nil)
		for i, data := range m {
			if names[i] == "" {
				continue
			}
			datasets = append(datasets, g.Map{
				"label":       names[i],
				"data":        data,
				"color":       colors[i%len(colors)],
				"borderWidth": 2,
				"tension":     0.4,
			})
		}

		data := g.Map{
			"labels":   labels,
			"datasets": datasets,
		}

		app.Eval(fmt.Sprintf("loading(%s)", conv.String(data)))

		return nil
	})
}
