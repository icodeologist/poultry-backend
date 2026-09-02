package models

type Admin struct {
	ID         uint   `gorm:"primaryKey"`
	AdminEmail string `json:"admin_email"`
	AdminName  string `json:"admin_name"`
	Password   string `json:"password"`
	Role       string `json:"role"`
}

type UserLoginInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
