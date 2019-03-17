module github.com/apisite/pgcall/pgx-pgcall

require (
	github.com/apisite/pgcall/pgiface v0.0.0
	github.com/birkirb/loggers-mapper-logrus v0.0.0-20180326232643-461f2d8e6f72
	github.com/cockroachdb/apd v1.1.0 // indirect
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgx v3.3.0+incompatible
	github.com/kr/pretty v0.1.0 // indirect
	github.com/lib/pq v1.0.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24 // indirect
	github.com/sirupsen/logrus v1.3.0
	github.com/stretchr/testify v1.3.0
	gopkg.in/birkirb/loggers.v1 v1.1.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/apisite/pgcall/pgiface v0.0.0 => ../pgiface
