package subjectidentifier

import "fmt"

type Identifier interface{}

type Email struct {
	Format string `json:"format"`
	Email  string `json:"email"`
}

func NewEmail(email string) Identifier {
	return &Email{
		Format: "email",
		Email:  email,
	}
}

func NewIdentifier(field, value string) (Identifier, error) {
	switch field {
	case "email":
		return NewEmail(value), nil
	default:
		return nil, fmt.Errorf("unsupported identifier type: %s", field)
	}
}
