# Grand Tour

## Getting Started

Create the database:

```bash
cat schema.sql | sqlite3 app.db
```

You can find your client id and secret on the [my API application](https://www.strava.com/settings/api) page.

```toml
# config.toml
id = 123
secret = "baaaaaad"
```
