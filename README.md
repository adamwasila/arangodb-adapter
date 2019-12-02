# ArangoDB Adapter

ArangoDB Adapter is the [Arango DB](https://www.arangodb.com/) adapter for [Casbin](https://github.com/casbin/casbin).

> **WARNING**: this is merely a hack; product of few hours spent learning go driver for arango, arango itself and adopting it to casbin needs. Definitely not a "production ready" quality yet. Use at your own risk.

## TODO

- Adapter cleanup (closing connections). See [this issue](https://github.com/arangodb/go-driver/issues/43).
- ~~Remove hardcoded db & collection names.~~
- Indexes.
- Filtered policies.
- Policy removal.
- Unit tests.
- Better README (examples of use).

## Getting Help

- [Casbin](https://github.com/casbin/casbin)

## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
