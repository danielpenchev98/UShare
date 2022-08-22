package dbconn

import (
	"fmt"
	"os"

	myerr "github.com/danielpenchev98/UShare/web-server/internal/error"
	"gorm.io/gorm"
)

const (
	//DBname - name of env variable, containing the name of the database
	dbName = "DB_NAME"

	//DBuser - name of env variable, containing the db username
	dbUser = "DB_USER"

	//DBpass - name of env variable, containing the db password
	dbPass = "DB_PASS"

	//DBport - name of env variable, containing the port on which the db server is running on
	dbPort = "DB_PORT"

	//DBdomain - name of env variable, containing the domain of the db server
	dbHost = "DB_HOST"
)

var dbConn *gorm.DB

//GetDBConn - creates a database connection or returns an already existing one
func GetDBConn(creator func(string) gorm.Dialector) (*gorm.DB, error) {
	if dbConn != nil {
		return dbConn, nil
	}

	dbConn, err := gorm.Open(creator(getDBDns()), &gorm.Config{})
	if err != nil {
		return nil, myerr.NewServerErrorWrap(err, "Cannot create a connection to the database.")
	}

	return dbConn, nil

}

func getDBDns() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		os.Getenv(dbHost),
		os.Getenv(dbUser),
		os.Getenv(dbPass),
		os.Getenv(dbName),
		os.Getenv(dbPort),
	)
}
