package dto

type UserAuth struct {
	ID       string `json:"id,omitempty" form:"-"`
	Username string `json:"username,omitempty" form:"-"`
	Email    string `json:"email,omitempty" form:"-"`
	FullName string `json:"full_name,omitempty" form:"-"`
	Level    string `json:"level,omitempty" form:"-"`
	Token    string `json:"token,omitempty" form:"-"`
	IsAuth   bool   `json:"is_auth,omitempty" form:"-"`
}
