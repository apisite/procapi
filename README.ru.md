<p align="center">
  <a href="README.md">English</a> |
  <span>Pусский</span>
</p>

---

# apisite/apimap
> golang - библиотека для построения API на основе map.

[![GoDoc][gd1]][gd2]
 [![codecov][cc1]][cc2]
 [![GoCard][gc1]][gc2]
 [![GitHub Release][gr1]][gr2]
 [![GitHub code size in bytes][sz]]()
 [![GitHub license][gl1]][gl2]

[cc1]: https://codecov.io/gh/apisite/pgcall/branch/master/graph/badge.svg
[cc2]: https://codecov.io/gh/apisite/pgcall
[gd1]: https://godoc.org/github.com/apisite/pgcall?status.svg
[gd2]: https://godoc.org/github.com/apisite/pgcall
[gc1]: https://goreportcard.com/badge/github.com/apisite/tpl2x
[gc2]: https://goreportcard.com/report/github.com/apisite/pgcall
[gr1]: https://img.shields.io/github/release-pre/apisite/pgcall.svg
[gr2]: https://github.com/apisite/pgcall/releases
[sz]: https://img.shields.io/github/languages/code-size/apisite/pgcall.svg
[gl1]: https://img.shields.io/github/license/apisite/pgcall.svg
[gl2]: LICENSE

## Назначение

Проект имеет целью предоставить доступ к хранимым функциям БД (например, postgresql) следующим потребителям:
* golang-шаблонам, для формирования страниц сайта с использованием данных из БД
* внешним клиентам, для вызова функций БД из, например, javascript

Т.к. вся необходимая потребителям информация (включая список и сигнатуры функций БД) изначально размещена в БД, при ее изменении не должна требоваться перекомпиляция golang-кода.

Т.о., необходимо реализовать функцию вида:
```
func call(method string, args map[string]interface{}) ([]map[string]interface{}, error) {}
```
для случая, когда методы API представляют собой функции Postgresql.

### Дополнения

1. Выбор `map` для аргументов обусловлен необходимостью различать
  * аргумент со значением
  * аргумент без значения (NULL)
  * аргумент о значением по умолчанию
2. Использование `map` в результате, для golang-потребителей, может быть сведено к структурам с помощью [mapstructure](https://github.com/mitchellh/mapstructure)
3. На основе доступной в БД информации об аргументах функций, должна быть реализована их валидация

## Структура

Библиотека разделена на следующие части:

* [pgcall]() - реализация функции `Call`
* [pgx-pgcall]() - функционал работы с БД postgresql посредством пакета jackc/pgx
* [gin-pgcall]() - интеграция функционала в [gin](https://github.com/gin-gonic/gin)

## Особенности реализации

### Сигнатуры методов

* реестр доступных функций (и их сигнатуры) хранится в БД, там же хранятся описания функций, аргументов, результатов и примеры вызова (в перспективе - с поддержкой i18n)
* доступ к реестру производится через вызовы специальных функций БД, что позволяет скрыть от потребителей внутренние функции (в перспективе это можно сделать через права доступа в БД)
* маппинг между внешним именем функции и ее именем в БД позволяет прозрачно для клиента заменить вызываемую функцию (при совпадении сигнатур)

### pgcall

Для получения информации из реестра используются функции с зашитыми в код сигнатурами (их имена могут быть изменены в настройках):

* IndexFunc (default:"index") - список доступных функций, вызывается запросом вида `select code, nspname, proname, anno, sample, result, is_ro, is_set, is_struct from %s(namespace)`
* InDefFunc (default:"func_args") - описание аргументов функции, вызывается запросом вида `select arg, type, required, def_val, anno from %s(code)`
* OutDefFunc (default:"func_result") - описание результата функции, вызывается запросом вида `select arg, type, anno from %s(code)`

### pgx-pgcall

Реализует объект, отвечающий за взаимодействие с БД. Передается в pgcall как интерфейс [pgcall.DB](https://godoc.org/github.com/apisite/pgcall#DB)

### gin-pgcall

Добавляет в gin маршрутизацию для прямого вызова функций API и дополняет funcMap функциями доступа к API из шаблонов.
Для работы с pgcall используется интерфейс [ginpgcall.Caller](https://godoc.org/github.com/apisite/pgcall/gin-pgcall#Caller)
