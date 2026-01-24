package domain

type User struct {
	ID       string
	Name     string
	Email    string
	Password string
	Role     UserRole
}

type UserRole string

const (
	RoleAdmin    UserRole = "ADMIN"
	RoleOperator UserRole = "OPERATOR"
)
