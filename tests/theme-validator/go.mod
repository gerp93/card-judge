module github.com/grantfbarnes/card-judge/tests/theme-validator

go 1.24.0

require (
	github.com/chromedp/chromedp v0.11.2
	github.com/grantfbarnes/card-judge/tests/setup v0.0.0
	github.com/grantfbarnes/card-judge/tests/util v0.0.0
	github.com/jung-kurt/gofpdf v1.16.2
)

replace github.com/grantfbarnes/card-judge/tests/setup => ../setup

replace github.com/grantfbarnes/card-judge/tests/util => ../util

replace github.com/grantfbarnes/card-judge => ../../src

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/chromedp/cdproto v0.0.0-20241022234722-4d5d5faf59fb // indirect
	github.com/chromedp/sysutil v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	github.com/grantfbarnes/card-judge v0.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	golang.org/x/sys v0.26.0 // indirect
)
