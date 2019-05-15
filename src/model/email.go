package model

type EmailConfig struct {
	Provider string `bson:"provider"`
	From     string `bson:"from"`

	Host     string `bson:"host"`
	User     string `bson:"user"`
	Password string `bson:"pass"`
	Secure   bool   `bson:"secure"`
	Port     int    `bson:"port"`
}
