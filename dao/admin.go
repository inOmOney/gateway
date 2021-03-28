package dao

import (
	"gateway/dto"
	"gateway/lib"
	"gateway/public"
	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

type Admin struct {
	Id        int       `json:"id"  description:"自增主键"`
	UserName  string    `json:"user_name" " description:"管理员用户名"`
	Salt      string    `json:"salt"  description:"盐"`
	Password  string    `json:"password"  description:"密码"`
	UpdatedAt time.Time `json:"update_at"  description:"更新时间"`
	CreatedAt time.Time `json:"create_at" description:"创建时间"`
	IsDelete  int       `json:"is_delete"  description:"是否删除"`
}

func (admin *Admin) LoginCheck(c *gin.Context, db lib.TDManager, param *dto.AdminLoginInput) (*Admin, error) {


	adminInfo, err := admin.Find(c, db, &Admin{UserName: param.UserName})
	if err != nil {
		return nil, errors.New("用户信息不存在")
	}
	salt := public.PasswordAddSalt(param.PassWord, adminInfo.Salt)
	if salt != adminInfo.Password {
		return nil, errors.New("密码错误，请重新输入")
	}
	return adminInfo, nil
}

// 用过用户名找整个用户信息
func (admin *Admin) Find(c *gin.Context, db lib.TDManager, param *Admin) (*Admin, error) {
	where := map[string]interface{}{
		"user_name": param.UserName,
		"is_delete": 0,
	}
	table := "gateway_admin"
	selectFields := []string{"*"}

	cond, vals, err := builder.BuildSelect(table, where, selectFields)
	if err != nil {
		return nil, errors.New("系统繁忙")
	}

	row, err := lib.DBQuery(public.GetGinTraceContext(c), db, cond, vals... )
	if err != nil {
		return nil, errors.New("系统繁忙")
	}
	adminInfo := &Admin{}
	scanner.Scan(row, &adminInfo)
	return adminInfo, nil
}

func (admin *Admin) Save(db lib.TDManager, param *Admin) error{
	where := map[string]interface{}{
		"user_name": param.UserName,
	}
	update := map[string]interface{}{
		"password": param.Password,
	}
	table := "gateway_admin"
	cond, vals, err := builder.BuildUpdate(table, where, update)
	if err != nil {
		return errors.New("系统繁忙")
	}
	_, err = db.Exec(cond, vals...)
	if err != nil {
		return errors.New("系统繁忙")
	}
	//if affected, _ := result.RowsAffected();affected!= int64(1){
	//	return errors.New("修改成功")
	//}
	return nil
}

