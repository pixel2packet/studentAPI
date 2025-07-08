package types

type Student struct {
	Id    int    `validate:"required"`
	Name  string `validate:"required"`
	Email string `validate:"required"`
	Age   int    `validate:"required"`
}
