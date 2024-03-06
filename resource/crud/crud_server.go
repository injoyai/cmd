package crud

var ServerTemp = `package server_{Lower}

import (
	"errors"
	model_{Lower} "{mod}/app/model/{Lower}"
	"gitee.com/injoyai/goutil/database/xorms"
)

var DB *xorms.Engine

func Init(db *xorms.Engine){
	DB=db
}

func Get{Upper}List(req *model_{Lower}.{Upper}ListReq) ([]*model_{Lower}.{Upper}, int64, error) {
	data := []*model_{Lower}.{Upper}{}
	session := DB.Limit(req.Size, req.Size*req.Index)
	if len(req.Name) > 0 {
		session.Where("Name like ?", "%"+req.Name+"%")
	}
	co, err := session.FindAndCount(&data)
	return data, co, err
}

func Get{Upper}(id int64) (*model_{Lower}.{Upper}, error) {
	data := new(model_{Lower}.{Upper})
	has, err := DB.Where("ID=?", id).Get(data)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("记录不存在")
	}
	return data, nil
}

func Post{Upper}(req *model_{Lower}.{Upper}Req) error {
	data, cols, err := req.New()
	if err != nil {
		return err
	}
	if req.ID > 0 {
		_, err = DB.Where("ID=?", req.ID).Cols(cols).Update(data)
		return err
	}
	_, err = DB.Insert(data)
	return err
}

func Del{Upper}(id int64) error {
	_, err := DB.Where("ID=?", id).Delete(new(model_{Lower}.{Upper}))
	return err
}


`
