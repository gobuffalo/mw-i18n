module github.com/gobuffalo/mw-i18n/v2

go 1.16

require (
	github.com/gobuffalo/buffalo v0.17.5
	github.com/gobuffalo/httptest v1.5.1
	github.com/nicksnyder/go-i18n v1.10.1
	github.com/stretchr/testify v1.7.0
)

replace (
	github.com/gobuffalo/buffalo v0.17.5 => github.com/fasmat/buffalo v0.16.15-0.20211121174727-77319a4a9d1a
	github.com/gobuffalo/pop/v6 v6.0.0 => github.com/fasmat/pop/v6 v6.0.0-20211121174542-8ace23c76ee8
)
