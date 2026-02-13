package document

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openPostgres(dsn string) gorm.Dialector {
	return postgres.Open(dsn)
}
