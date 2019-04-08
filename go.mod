module github.com/apisite/procapi

go 1.12

require (
	github.com/apisite/procapi/ginproc v0.0.0-00010101000000-000000000000 // indirect
	github.com/apisite/procapi/pgtype v0.0.0-00010101000000-000000000000
	github.com/birkirb/loggers-mapper-logrus v0.0.0-20180326232643-461f2d8e6f72
	github.com/golang/mock v1.2.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/lib/pq v1.0.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.1
	github.com/stretchr/testify v1.3.0
	gopkg.in/birkirb/loggers.v1 v1.1.0
)

replace (
	github.com/apisite/procapi/ginproc => ./ginproc
	github.com/apisite/procapi/pgtype => ./pgtype
)
