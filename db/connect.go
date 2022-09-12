package db

var DB *PrismaClient

func Connect() {
	DB = NewClient()
	if DB.Connect() != nil {
		panic("dbError")
	}
}
func Disconnect() {
	DB.Disconnect()
}
