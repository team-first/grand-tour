# Grand Tour

A quick and dirty collaborative test of Strava's API. Do not use.


## Getting Started

Install all dependencies (ensuring `GOPATH` is set):

```bash
go get github.com/BurntSushi/toml
go get github.com/mattn/go-sqlite3
go get github.com/strava/go.strava
```

Create the database:

```bash
cat schema.sql | sqlite3 app.db
```

Create the `config.toml` app config file. This sets your API client id and
secret, which you can find on the [my API application](https://www.strava.com/settings/api)
page. The file should have the following contents:

```toml
# config.toml
id = 123
secret = "baaaaaad"
```

Launch the server:

```bash
go run app.go
```

Connect at `localhost:8080`.


## In Depth

TBD


## License

TBD. Assume written by incompetant monkeys. All bets are off. No liability for
us.


## To Do
- Finish this readme


## Additional Notes
Issue / Feature / Roadmap tracking on [Trello](https://trello.com/b/DKVpD6aW) if
we care.
