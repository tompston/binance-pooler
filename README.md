# Syro

Timeseries data scraper example, implementing the ideas mentioned in [this tutorial](https://tompston.pages.dev/writing/2024-06-29-everything-about-timeseries-data-scraping).

Scrapes futures data from binance api and stores it in a mongodb database.

Additional features:

- [ ] Logging
- [ ] Tracking of cron job state and executions

## Dependencies

- Go (min version 1.20)
- Running instance of mongodb. Default port, no auth. [installation link](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-on-ubuntu/)
- Reflex (for rebuilding the go app on changes) [(installation link)](https://github.com/cespare/reflex)

## Running the app

```bash
chmod +x run.sh

# start the syro/cmd/pooler app
./run.sh pooler
# start the syro/cmd/api app
./run.sh api
# run tests for the project (will be written under mongodb database called `test`)
./run.sh test
```

### Project structure

```bash
├── syro
│   ├── cmd               # entry points for the apps
│   ├── conf              # config files
│   ├── internal
│   │   ├── api           # api implementation
│   │   └── pooler        # scraper implementations
│   └── pkg
│       ├── app           # main app struct and config
│       ├── lib
│       │   ├── encoder   # faster marshalling and unmarshalling of json
│       │   ├── errgroup  # package for grouping errors
│       │   ├── fetcher   # utlil for doing http requests
│       │   ├── logger    # logger package, using interfaces
│       │   ├── messenger # sending messages to telegram or google chat groups
│       │   ├── mongodb   # mongodb utils
│       │   ├── scheduler # cron scheduler + tracker
│       └── models        # models for the project
```

### TODOS

- [ ] Minimal ui for previewing the logs and crons
- [ ] Api endpoint for storing event logs (so that they can be previewed)
- [ ] write openapi spec for the api
- [ ] test api endpoints
