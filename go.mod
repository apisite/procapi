module github.com/apisite/procapi

go 1.12

replace (
	github.com/apisite/procapi/ginproc => ./ginproc
	github.com/apisite/procapi/pgtype => ./pgtype
	github.com/pgmig/pgx-scanmore => ./../../pgmig/pgx-scanmore

)

require (
	//	github.com/apisite/procapi/pgtype v0.0.0-00010101000000-000000000000
	github.com/birkirb/loggers-mapper-logrus v0.0.0-20180326232643-461f2d8e6f72
	github.com/gin-gonic/gin v1.5.0
	github.com/golang/mock v1.3.1
	github.com/jackc/pgtype v1.0.2
	github.com/jackc/pgx/v4 v4.1.2
	github.com/jessevdk/go-flags v1.4.0
	github.com/lib/pq v1.2.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pgmig/pgmig v0.35.0
	github.com/pgmig/pgx-scanmore v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	gopkg.in/birkirb/loggers.v1 v1.1.0
)
