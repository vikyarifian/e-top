package dto

type UserAuth struct {
	ID       string
	Username string
	Email    string
	FullName string
	Level    string
	Token    string
	IsAuth   bool
}
