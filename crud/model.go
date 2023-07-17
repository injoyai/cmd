package crud

var ModelTemp = `package model_{Lower}

type {Upper} struct {
	ID     int64  ` + "`" + `json:"id" xorm:"ID" gorm:"column:ID"` + "`" + `
	InDate int64  ` + "`" + `json:"inDate" xorm:"created 'InDate''" gorm:"column:InDate"` + "`" + `
	Name   string ` + "`" + `json:"name" xorm:"Name" gorm:"column:Name"` + "`" + `
}

type {Upper}ListReq struct{
	Index int ` + "   `" + `json:"` + `index"` + "`" + `
	Size  int ` + "   `" + `json:"` + `size"` + "`" + `
	Name  string ` + "`" + `json:"` + `name"` + "`" + `
}

type {Upper}Req struct {
	ID     int64  ` + "`" + `json:"id" xorm:"ID" gorm:"column:ID"` + "`" + `
	InDate int64  ` + "`" + `json:"inDate" xorm:"InDate" gorm:"column:InDate"` + "`" + `
	Name   string ` + "`" + `json:"name" xorm:"Name" gorm:"column:Name"` + "`" + `
}

func (this {Upper}Req) New() (*{Upper}, string ,error) {
	return &{Upper}{
		Name : this.Name,
	}, "", nil
}
`
