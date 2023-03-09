# ArangoDB Adapter

[![GoDoc](https://godoc.org/github.com/adamwasila/arangodb-adapter?status.svg)](https://godoc.org/github.com/adamwasila/arangodb-adapter) [![Go Report Card](https://goreportcard.com/badge/adamwasila/arangodb-adapter)](https://goreportcard.com/report/adamwasila/arangodb-adapter) [![Build Status](https://github.com/adamwasila/arangodb-adapter/actions/workflows/main.yml/badge.svg)](https://github.com/adamwasila/arangodb-adapter/actions/workflows/main.yml)

ArangoDB Adapter is the [Arango DB](https://www.arangodb.com/) adapter for [Casbin](https://github.com/casbin/casbin).

## TODO

- Adapter cleanup (closing connections). See [this issue](https://github.com/arangodb/go-driver/issues/43).
- ~~Remove hardcoded db & collection names.~~
- ~~Indexes.~~
- Filtered policies.
- ~~Policy removal.~~
- ~~Add partial policy removal.~~
- ~~Unit tests.~~
- Better README (examples of use).

## Example

Following snippet of code shows how to initialize adapter and use with casbin enforcer. See [documentation](https://godoc.org/github.com/adamwasila/arangodb-adapter) for list of all available options.

```golang
a, err := arango.NewAdapter(
    arango.OpCollectionName("casbinrules"),
    arango.OpFieldMapping("p", "sub", "obj", "act"))
if err != nil {
    ...
}

e, err := casbin.NewEnforcer("model.conf", a)

...

```

## Contributing

### Documentation

Currently this README and examples folder are best source of documentation for this project and of course - source code itself.

### Reporting issues

Raise an issue for bugs, enhancements and general discussions/questions about adapter.

### Running tests

It would make very little sense to perform an isolated unit tests for code like that. Therefore tests connects to real database instance. To test fully all options two instances must be run: with and without authorization. [CI setup](.travis.yml) may be helpful to establish working testing rig:

```console
docker run -e ARANGO_NO_AUTH=1 -p 127.0.0.1:8529:8529 -d --name arangodb-instance-no-auth arangodb:3.7.2
docker run -e ARANGO_ROOT_PASSWORD=password -p 127.0.0.1:8530:8529 -d --name arangodb-instance-auth arangodb:3.7.2
```

Then, running tests is as simple as:

```
go test .
ok  	github.com/adamwasila/arangodb-adapter	0.847s
```

### Pull requests

If possible each PR should be linked to some issue (except trivial ones like typo fixes). Avoid unrelated changes. Redundant commits should be squashed together before merge.

## Getting Help

- [Casbin](https://github.com/casbin/casbin) - main library this adapter is extending

## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
