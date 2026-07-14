package models

type Admin struct {
	ID         int    `gorm:"primaryKey"`
	AdminEmail string `json:"admin_email"`
	AdminName  string `json:"admin_name"`
	Password   string `json:"password"`
}

type AdminLoginInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
