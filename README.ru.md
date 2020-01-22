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
 [![LoC][loc1]][loc2]
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
[loc1]: https://raw.githubusercontent.com/apisite/procapi/master/.loc.svg?sanitize=true "Lines of Code"
[loc2]: https://github.com/apisite/procapi/blob/master/LOC.md
[gl1]: https://img.shields.io/github/license/apisite/procapi.svg
[gl2]: https://github.com/apisite/procapi/blob/master/LICENSE


## Назначение

Пакет [apisite/procapi](https://github.com/apisite/procapi) является частью проекта [apisite](https://github.com/apisite/apisite) и предназначен для предоставления доступа к хранимым функциям БД (например, postgresql) следующим потребителям:
* golang-шаблонам, для формирования страниц сайта с использованием данных из БД
* внешним клиентам, для вызова функций БД из, например, javascript

Т.к. вся необходимая потребителям информация (включая список и сигнатуры функций БД) изначально размещена в БД, при ее изменении не требуется перекомпиляция golang-кода.

Задача пакета решается функцией вида:
```go
func Call(method string, args map[string]interface{}) ([]map[string]interface{}, error) {}
```
для случая, когда методы API представляют собой функции Postgresql.

### Дополнения

1. Выбор `map` для аргументов обусловлен необходимостью различать
  * аргумент со значением (/name?arg=XX)
  * аргумент без значения (NULL) (/name?arg)
  * аргумент со значением по умолчанию (/name)
2. Использование `map` в результате, для golang-потребителей, может быть сведено к структурам с помощью [mapstructure](https://github.com/mitchellh/mapstructure)
3. Доступная в БД информация об аргументах функций используется для их валидации

## Структура

Библиотека разделена на следующие части:

* [procapi](https://github.com/apisite/procapi) - реализация функции `Call`
* [ginproc](https://github.com/apisite/procapi/tree/master/ginproc) - интеграция функционала в [gin](https://github.com/gin-gonic/gin)

## Особенности реализации

### Тесты

Варианты запуска:

1. `TEST_DATABASE="{PG_DSN}" TZ="Europe/Berlin" go test ./...`
3. `make cov` - тесты с генерацией отчета о покрытии (см `make help`)

#### Postgresql

Т.к. для работы с БД могут потребоваться параметры соединения, они вынесены в файл настроек `.env`, для создания которого необходимо выполнить `make config`.

Варианты запуска Postgresql:

1. Внешний, в настройках задаются параметры соединения, пользователь и БД должны существовать
2. Внутренний (с помощью docker). Для запуска необходимо в отдельной консоли выполнить `make test-docker-run`

В любом из этих случаев, при выполнении тестов будут созданы и наполнены данными 3 схемы БД, которые после выполнения тестов будут удалены (по завершении тестов выполняется `ROLLBACK`).

Необходимый для тестов SQL-код подключается в каталог тестовых данных (`testdata`) посредством `git submodule`. Т.о., если проект не был склонирован с ключем `--recursive`, для подгрузки SQL необходимо выполнить
```
git submodule init
git submodule update
```

Используемый для тестов SQL-код разработан в рамках проекта [pgmig](https://github.com/pgmig) и включает пакеты:

* [pgmig](https://github.com/pgmig-sql/pgmig) - сервисные функции
* [rpc](https://github.com/pgmig-sql/rpc) - поддержка RPC
* [rpc_testing](hhttps://github.com/pgmig-sql/rpc_testing) - тестовые функции для [procapi](https://github.com/apisite/procapi)

См. также: [.drone.yml](https://github.com/apisite/procapi/blob/master/.drone.yml) - пример запуска тестов

### Сигнатуры методов

* реестр доступных функций (и их сигнатуры) хранится в БД, там же хранятся описания функций, аргументов, результатов и примеры вызова (в перспективе - с поддержкой i18n)
* в реестре также задается соответствие между именем функции в API и в БД
* доступ к реестру производится через вызовы специальных функций БД (их схема и имена задаются в настройках)

### procapi

Для получения информации из реестра используются функции с зашитыми в код сигнатурами (их имена могут быть изменены в настройках), вызов имеет вид `SELECT * FROM %c(code)`:

* `--db.index $name` (default:"index") - список доступных функций, структура ответа - [Method](https://godoc.org/github.com/apisite/procapi#Method)
* `--db.indef $name` (default:"func_args") - описание аргументов функции, структура ответа - [InDef](https://godoc.org/github.com/apisite/procapi#InDef)
* `--db.outdef $name` (default:"func_result") - описание результата функции, структура ответа - [OutDef](https://godoc.org/github.com/apisite/procapi#OutDef)

### pgtype

Структура `PGType` дополняет возможности [pgx v4](https://github.com/jackc/pgx/tree/v4). Код реализовывает интерфейс [Marshaller](https://godoc.org/github.com/apisite/procapi#Marshaller) и может быть заменен другим с помощью вызова [SetMarshaller](https://godoc.org/github.com/apisite/procapi#Service.SetMarshaller).

### ginproc

Пакет добавляет в gin маршрутизацию для прямого вызова функций API и дополняет funcMap функциями доступа к API из шаблонов.
Для работы с procapi используется интерфейс [ginproc.Caller](https://godoc.org/github.com/apisite/procapi/ginproc#Caller)

## TODO

* [ ] JWT + OAuth
* [ ] Проверка доступа
* [ ] [Кэш](https://github.com/golang/groupcache)
* [ ] тесты ошибок
* [ ] поддержка нативных методов API (golang)
* [ ] LISTEN для фоновой обработки задач
* [ ] Read-Only транзакции для Method.IsRO
* [ ] Rollback транзакции для тестов из шаблонов

## История

* 2010 - PGWS, 1я реализация API для функций Postgresql (perl)
* 2012 - [PGWS](https://github.com/LeKovr/pgws) - часть проекта опубликована на github.com
* 2016 - [dbrpc](https://github.com/LeKovr/dbrpc) - реализация API на go (+pgx)
* 2018 - [pgfc](https://github.com/apisite/pgfc) - API с поддержкой вызова из шаблонов
* 2019 - [procapi] - API с тестами и поддержкой транзакций
