<p align="center">
  <a href="README.md#apisiteprocapi">English</a> |
  <span>Pусский</span>
</p>

---

# apisite/procapi
> API для хранимых процедур БД

[![GoDoc][gd1]][gd2]
 [![codecov][cc1]][cc2]
 [![Build Status][bs1]][bs2]
 [![GoCard][gc1]][gc2]
 [![GitHub Release][gr1]][gr2]
 [![GitHub code size in bytes][sz]]()
 [![GitHub license][gl1]][gl2]

[bs1]: https://cloud.drone.io/api/badges/apisite/procapi/status.svg
[bs2]: https://cloud.drone.io/apisite/procapi
[cc1]: https://codecov.io/gh/apisite/procapi/branch/master/graph/badge.svg
[cc2]: https://codecov.io/gh/apisite/procapi
[gd1]: https://godoc.org/github.com/apisite/procapi?status.svg
[gd2]: https://godoc.org/github.com/apisite/procapi
[gc1]: https://goreportcard.com/badge/github.com/apisite/procapi
[gc2]: https://goreportcard.com/report/github.com/apisite/procapi
[gr1]: https://img.shields.io/github/release-pre/apisite/procapi.svg
[gr2]: https://github.com/apisite/procapi/releases
[sz]: https://img.shields.io/github/languages/code-size/apisite/procapi.svg
[gl1]: https://img.shields.io/github/license/apisite/procapi.svg
[gl2]: LICENSE

## Назначение

Проект имеет целью предоставить доступ к хранимым функциям БД (например, postgresql) следующим потребителям:
* golang-шаблонам, для формирования страниц сайта с использованием данных из БД
* внешним клиентам, для вызова функций БД из, например, javascript

Т.к. вся необходимая потребителям информация (включая список и сигнатуры функций БД) изначально размещена в БД, при ее изменении не должна требоваться перекомпиляция golang-кода.

Т.о., необходимо реализовать функцию вида:
```go
func Call(method string, args map[string]interface{}) ([]map[string]interface{}, error) {}
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

* [procapi]() - реализация функции `Call`
* [pgx-procapi]() - функционал работы с БД postgresql посредством пакета jackc/pgx
* [gin-procapi]() - интеграция функционала в [gin](https://github.com/gin-gonic/gin)

## Особенности реализации

### Сигнатуры методов

* реестр доступных функций (и их сигнатуры) хранится в БД, там же хранятся описания функций, аргументов, результатов и примеры вызова (в перспективе - с поддержкой i18n)
* доступ к реестру производится через вызовы специальных функций БД, что позволяет скрыть от потребителей внутренние функции (в перспективе это можно сделать через права доступа в БД)
* маппинг между внешним именем функции и ее именем в БД позволяет прозрачно для клиента заменить вызываемую функцию (при совпадении сигнатур)

### procapi

Для получения информации из реестра используются функции с зашитыми в код сигнатурами (их имена могут быть изменены в настройках):

* IndexFunc (default:"index") - список доступных функций, вызывается запросом вида `select code, nspname, proname, anno, sample, result, is_ro, is_set, is_struct from %s(namespace)`
* InDefFunc (default:"func_args") - описание аргументов функции, вызывается запросом вида `select arg, type, required, def_val, anno from %s(code)`
* OutDefFunc (default:"func_result") - описание результата функции, вызывается запросом вида `select arg, type, anno from %s(code)`

### pgx-procapi

Реализует объект, отвечающий за взаимодействие с БД. Передается в procapi как интерфейс [procapi.DB](https://godoc.org/github.com/apisite/procapi#DB)

### gin-procapi

Добавляет в gin маршрутизацию для прямого вызова функций API и дополняет funcMap функциями доступа к API из шаблонов.
Для работы с procapi используется интерфейс [ginprocapi.Caller](https://godoc.org/github.com/apisite/procapi/gin-procapi#Caller)
