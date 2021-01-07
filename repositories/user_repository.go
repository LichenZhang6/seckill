package repositories

import (
	"database/sql"
	"errors"
	"seckill/common"
	"seckill/datamodels"
	"strconv"
)

type IUserRepository interface {
	Conn() error
	Select(string) (*datamodels.User, error)
	Insert(*datamodels.User) (int64, error)
}

type UserManager struct {
	table     string
	mysqlConn *sql.DB
}

func NewUserManager(table string, db *sql.DB) IUserRepository {
	return &UserManager{table: table, mysqlConn: db}
}

func (u *UserManager) Conn() (err error) {
	if u.mysqlConn == nil {
		mysql, errMysql := common.NewMysqlConn()
		if errMysql != nil {
			return errMysql
		}
		u.mysqlConn = mysql
	}

	if u.table == "" {
		u.table = "user"
	}
	return
}

func (u *UserManager) Select(userName string) (user *datamodels.User, err error) {
	if userName == "" {
		return &datamodels.User{}, errors.New("用户名不能为空")
	}

	if err = u.Conn(); err != nil {
		return &datamodels.User{}, err
	}

	sql := "SELECT * FROM " + u.table + " WHERE userName=?"
	row, errRows := u.mysqlConn.Query(sql, userName)
	defer row.Close()
	if errRows != nil {
		return &datamodels.User{}, errRows
	}

	result := common.GetResultRow(row)
	if len(result) == 0 {
		return &datamodels.User{}, errors.New("用户不存在")
	}
	user = &datamodels.User{}
	common.DataToStructByTagSql(result, user)
	return
}

func (u *UserManager) Insert(user *datamodels.User) (userId int64, err error) {
	if err = u.Conn(); err != nil {
		return
	}

	sql := "INSERT " + u.table + " SET nickName=?, userName=?, passWord=?"
	stmt, errStmt := u.mysqlConn.Prepare(sql)
	if errStmt != nil {
		return userId, errStmt
	}

	result, errResult := stmt.Exec(user.NickName, user.UserName, user.HashPassword)
	if errResult != nil {
		return userId, errResult
	}

	return result.LastInsertId()
}

func (u *UserManager) SelectByID(userId int64) (user *datamodels.User, err error) {
	if err = u.Conn(); err != nil {
		return &datamodels.User{}, err
	}

	sql := "SELECT * FROM " + u.table + " WHERE ID=" + strconv.FormatInt(userId, 10)
	row, errRow := u.mysqlConn.Query(sql)
	if errRow != nil {
		return &datamodels.User{}, errRow
	}

	result := common.GetResultRow(row)
	if len(result) == 0 {
		return &datamodels.User{}, errors.New("用户不存在")
	}
	user = &datamodels.User{}
	common.DataToStructByTagSql(result, user)
	return
}
