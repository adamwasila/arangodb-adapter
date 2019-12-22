# ArangoDB Adapter

[![GoDoc](https://godoc.org/github.com/adamwasila/arangodb-adapter?status.svg)](https://godoc.org/github.com/adamwasila/arangodb-adapter) [![Go Report Card](https://goreportcard.com/badge/adamwasila/arangodb-adapter)](https://goreportcard.com/report/adamwasila/arangodb-adapter) [![Build Status](https://travis-ci.com/adamwasila/arangodb-adapter.svg?branch=master)](https://travis-ci.com/adamwasila/arangodb-adapter)

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

### Pull requests

If possible each PR should be linked to some issue (except trivial ones like typo fixes). Avoid unrelated changes. Redundant commits should be squashed together before merge.

## Getting Help

- [Casbin](https://github.com/casbin/casbin) - main library this adapter is extending

## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
