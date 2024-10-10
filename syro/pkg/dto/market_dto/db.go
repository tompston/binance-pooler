package market_dto

// Mongo struct is used to group up the db specific methods for the dtos
type Mongo struct{}

func NewMongoInterface() *Mongo { return &Mongo{} }
