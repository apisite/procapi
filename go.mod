module github.com/apisite/pgcall

go 1.12

require (
	github.com/apisite/pgcall/pgiface v0.0.0
	github.com/birkirb/loggers-mapper-logrus v0.0.0-20180326232643-461f2d8e6f72
	github.com/blang/vfs v1.0.0 // indirect
	github.com/daaku/go.zipexe v1.0.0 // indirect
	github.com/golang/mock v1.2.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/phogolabs/parcello v0.8.1
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.0
	github.com/stretchr/testify v1.3.0
	gopkg.in/birkirb/loggers.v1 v1.1.0
)

replace github.com/apisite/pgcall/pgiface v0.0.0 => ./pgiface
