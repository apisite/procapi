# gin-pgcall
> клей pgcall для gin-gonic

[![GoCard][gc1]][gc2]
 [![GitHub Release][gr1]][gr2]
 [![GitHub code size in bytes][sz]]()
 [![GitHub license][gl1]][gl2]

[gc1]: https://goreportcard.com/badge/apisite/pgcall/gin-pgcall
[gc2]: https://goreportcard.com/report/github.com/apisite/pgcall/gin-pgcall
[gr1]: https://img.shields.io/github/release/apisite/pgcall/gin-pgcall.svg
[gr2]: https://github.com/apisite/pgcall/releases
[sz]: https://img.shields.io/github/languages/code-size/apisite/gin-pgcall.svg
[gl1]: https://img.shields.io/github/license/apisite/gin-pgcall.svg
[gl2]: LICENSE

<p align="center">
  <a href="../../../../gin-pgcall/README.md">English</a> |
  <span>Русский</span>
</p>

* Статус проекта: Реализован концепт

[pgcall/gin-pgcall](https://github.com/apisite/pgcall/gin-pgcall) - golang библиотека для использования [pgcall](https://github.com/apisite/pgcall) в проектах на [gin-gonic](https://github.com/gin-gonic/gin).

## Использование

```
	allFuncs := template.FuncMap{}
	appendFuncs(allFuncs)

	s, err := pgcall.NewServer(cfg.pgcall, log, cfg.DBConnect, nil)
	if err != nil {
		log.Fatal(err)
	}
	s.SetFuncBlank(allFuncs)
	err = templates.LoadTemplates(allFuncs)
	if err != nil {
		log.Fatal(err)
	}

	s.Route("/rpc", r)

	templates.FuncHandler = func(ctx *gin.Context, funcs template.FuncMap) {
		s.SetFuncRequest(funcs, ctx)
	}

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}

```
## См. также

* [pgcall](https://github.com/apisite/pgcall) - golang библиотека для вызова хранимых функций postgresql
* enfist - пример готового приложения

## Похожие проекты

Есть много возможностей предоставить доступ к БД внешнему клиенту (в т.ч. клиенту на javascript). Ниже перечислены самые популярные, если знаете других, напишите нам и мы добавим:

* [postgrest](https://github.com/PostgREST/postgrest)
* [pgweb](https://sosedoff.github.io/pgweb/)

## Лицензия

Лицензия MIT (MIT), см. [LICENSE](LICENSE) (неофициальный перевод,
 [источник перевода](https://ru.wikipedia.org/wiki/%D0%9B%D0%B8%D1%86%D0%B5%D0%BD%D0%B7%D0%B8%D1%8F_MIT), [оригинал лицензии](../../LICENSE)).

Copyright (c) 2018 Алексей Коврижкин <lekovr+apisite@gmail.com>
