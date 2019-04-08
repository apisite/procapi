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

* [procapi](https://github.com/apisite/procapi) - реализация функции `Call`
* [pgtype](https://github.com/apisite/procapi/tree/master/pgtype) - функционал работы с БД postgresql посредством пакета jackc/pgx
* [ginproc](https://github.com/apisite/procapi/tree/master/ginproc) - интеграция функционала в [gin](https://github.com/gin-gonic/gin)

## Особенности реализации

### Тесты

#### Mock DB

Функционал доступа к БД подменяется с помощью [gomock](https://github.com/golang/mock/)

Варианты запуска:

1. `go test ./...` - если пакет находится в дереве с корнем `$GOPATH`
2. `GO111MODULE=on go test ./... ./pgtype/... ./ginproc/...` - при любом пути к пакету
3. `make cov` - тесты с генерацией отчета о покрытии

#### Postgresql

Т.к. для работы с БД могут потребоваться параметры соединения, они вынесены в файл настроек `.env`, для создания которого необходимо выполнить `make config`.

Варианты запуска Postgresql

1. Внешний, в настройках задаются параметры соединения, пользователь и БД должны существовать
2. Внутренний (с помощью docker). Для запуска необходимо в отдельной консоли выполнить `make test-docker-run`

В любом из этих случаев, при выполнении тестов будут созданы и наполнены данными 3 схемы БД, которые после выполнения тестов будут удалены (по завершении тестов выполняется `ROLLBACK`).

Для того, чтобы имена схем не пересеклись с уже существующими, к ним будет добавлен суффикс - случайная последовательность символов.

Необходимый для тестов SQL-код подключается в каталог тестовых данных (`testdata`) посредством `git submodule`. Т.о., если проект не был склонирован с ключем `--recursive`, для подгрузки SQL необходимо выполнить
```
git submodule init
git submodule update
```

Используемый для тестов SQL-код разработан в рамках проекта [pomasql](https://github.com/pomasql) и включает пакеты:

* [poma](https://github.com/pomasql/poma) - сервисные функции
* [rpc](https://github.com/pomasql/rpc) - поддержка RPC
* [rpc_testing](https://github.com/pomasql/rpc_testing) - тестовые функции для [procapi](https://github.com/apisite/procapi)

См. также: [.drone.yml](https://github.com/apisite/procapi/blob/master/.drone.yml) - пример запуска тестов

### Сигнатуры методов

* реестр доступных функций (и их сигнатуры) хранится в БД, там же хранятся описания функций, аргументов, результатов и примеры вызова (в перспективе - с поддержкой i18n)
* доступ к реестру производится через вызовы специальных функций БД, что позволяет скрыть от потребителей внутренние функции (в перспективе это можно сделать через права доступа в БД)
* маппинг между внешним именем функции и ее именем в БД позволяет прозрачно для клиента заменить вызываемую функцию (при совпадении сигнатур)

### procapi

Для получения информации из реестра используются функции с зашитыми в код сигнатурами (их имена могут быть изменены в настройках):

* `--db.index name` (default:"index") - список доступных функций, вызывается запросом вида `select code, nspname, proname, anno, sample, result, is_ro, is_set, is_struct from %s(namespace)`
* `--db.indef name` (default:"func_args") - описание аргументов функции, вызывается запросом вида `select arg, type, required, def_val, anno from %s(code)`
* `--db.outdef name` (default:"func_result") - описание результата функции, вызывается запросом вида `select arg, type, anno from %s(code)`

### pgtype

Модуль предназначен для обработки случаев, когда [sqlx](https://github.com/jmoiron/sqlx) не конвертирует полученные из БД значения необходимым для API образом. Код реализовывает интерфейс [Marshaller](https://godoc.org/github.com/apisite/procapi#Marshaller) и может быть заменен другим с помощью вызова [SetMarshaller](https://godoc.org/github.com/apisite/procapi#Service.SetMarshaller).

### ginproc

Модуль добавляет в gin маршрутизацию для прямого вызова функций API и дополняет funcMap функциями доступа к API из шаблонов.
Для работы с procapi используется интерфейс [ginproc.Caller](https://godoc.org/github.com/apisite/procapi/ginproc#Caller)
