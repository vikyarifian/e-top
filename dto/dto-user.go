package dto

type UserAuth struct {
	Username string
	Email    string
	FullName string
	Level    string
	Token    string
	IsAuth   bool
}
