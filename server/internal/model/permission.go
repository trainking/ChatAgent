package model

type Permission struct {
	Code        string `db:"code" json:"code"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
}

type RolePermissions struct {
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}
