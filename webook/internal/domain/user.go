package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Password string

	// UTC 0 的时区
	Ctime time.Time

	Phone    string
	Nickname string
	Birthday string
	AboutMe  string

	//Addr Address
}

//type Address struct {
//	Province string
//	Region   string
//}

//func (u User) ValidateEmail() bool {
// 在这里用正则表达式校验
//return u.Email
//}
