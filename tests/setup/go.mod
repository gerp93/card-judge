module github.com/grantfbarnes/card-judge/tests/setup

go 1.24.0

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/grantfbarnes/card-judge v0.0.0
	github.com/grantfbarnes/card-judge/tests/util v0.0.0
)

require filippo.io/edwards25519 v1.1.0 // indirect

replace github.com/grantfbarnes/card-judge => ../../src

replace github.com/grantfbarnes/card-judge/tests/util => ../util
