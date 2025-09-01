package crud

var ServerTemp = `package {Lower}

import (
	"errors"
)

func getList(req *{Upper}ListReq) ([]*{Upper}, int64, error) {
	data := []*{Upper}{}
	session := db.Limit(req.Size, req.Size*req.Index)
	if len(req.Name) > 0 {
		session.Where("Name like ?", "%"+req.Name+"%")
	}
	co, err := session.FindAndCount(&data)
	return data, co, err
}

func get(id int64) (*{Upper}, error) {
	data := new({Upper})
	has, err := db.Where("ID=?", id).Get(data)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("记录不存在")
	}
	return data, nil
}

func post(req *{Upper}Req) error {
	data, cols, err := req.New()
	if err != nil {
		return err
	}
	if req.ID > 0 {
		_, err = db.Where("ID=?", req.ID).Cols(cols).Update(data)
		return err
	}
	_, err = db.Insert(data)
	return err
}

func del(id int64) error {
	_, err := db.Where("ID=?", id).Delete(new({Upper}))
	return err
}


`
