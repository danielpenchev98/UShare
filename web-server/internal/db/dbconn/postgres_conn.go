package dbconn

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//PostgresDialectorCreator - creates postgre database specific dialector
func PostgresDialectorCreator(dbDNS string) gorm.Dialector {
	return postgres.Open(dbDNS)
}
