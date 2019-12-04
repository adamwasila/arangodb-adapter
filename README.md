# ArangoDB Adapter

[![Go Report Card](https://goreportcard.com/badge/adamwasila/arangodb-adapter)](https://goreportcard.com/report/adamwasila/arangodb-adapter) [![Build Status](https://travis-ci.com/adamwasila/arangodb-adapter.svg?branch=master)](https://travis-ci.com/adamwasila/arangodb-adapter)

ArangoDB Adapter is the [Arango DB](https://www.arangodb.com/) adapter for [Casbin](https://github.com/casbin/casbin).

> **WARNING**: this is merely a hack; product of few hours spent learning go driver for arango, arango itself and adopting it to casbin needs. Definitely not a "production ready" quality yet. Use at your own risk.

## TODO

- Adapter cleanup (closing connections). See [this issue](https://github.com/arangodb/go-driver/issues/43).
- ~~Remove hardcoded db & collection names.~~
- ~~Indexes.~~
- Filtered policies.
- ~~Policy removal.~~
- Add partial policy removal.
- Unit tests.
- Better README (examples of use).

## Getting Help

- [Casbin](https://github.com/casbin/casbin)

## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
