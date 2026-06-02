// controllers/main_test.go
package controllers

import (
	"os"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
)

// TestMain runs before all tests in the controllers package.
// It sets required config values in memory so no app.conf file is needed.
func TestMain(m *testing.M) {
	beego.AppConfig.Set("users_csv_path", "data/users.csv")
	beego.AppConfig.Set("expenses_csv_path", "data/expenses.csv")
	os.Exit(m.Run())
}
