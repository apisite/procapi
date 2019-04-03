module github.com/apisite/pgcall

go 1.12

require (
	github.com/apisite/pgcall/gin-pgcall v0.0.0-20190331220424-b9c86761817f // indirect
	github.com/apisite/pgcall/pgx-pgcall v0.0.0-20190403012219-1673d0ba2f0b
	github.com/birkirb/loggers-mapper-logrus v0.0.0-20180326232643-461f2d8e6f72
	github.com/golang/mock v1.2.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.0
	github.com/stretchr/testify v1.3.0
	gopkg.in/birkirb/loggers.v1 v1.1.0
)

replace (
	github.com/apisite/pgcall/gin-pgcall => ./gin-pgcall
	github.com/apisite/pgcall/pgx-pgcall => ./pgx-pgcall
)
