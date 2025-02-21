
Timeseries data scraper example, implementing the ideas mentioned in [this tutorial](https://tompston.pages.dev/writing/2024-06-29-everything-about-timeseries-data-scraping).

Scrapes spot data from binance api and stores it in a mongodb database.

Additional features:

- Logging
- Tracking of cron job state and executions (wrapper around the `robfig/cron/v3` package)

## Dependencies

- Go (min version 1.20)
- Running instance of mongodb. Default port, no auth. [installation link](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-on-ubuntu/)
- Reflex (for rebuilding the go app on changes) [(installation link)](https://github.com/cespare/reflex)

## Running the app

```bash
chmod +x run.sh

# start the binance-pooler/cmd/pooler app
./run.sh pooler
# start the binance-pooler/cmd/api app
./run.sh api
# run tests for the project (will be written under mongodb database called `test`)
./run.sh test
```

### Project structure

```bash
├── binance-pooler
│   ├── cmd               # entry points for the apps
│   ├── conf              # config files
│   ├── internal
│   │   ├── api           # api implementation
│   │   └── pooler        # data pooler implementations
│   └── pkg
│       ├── app           # main app struct and config
│       └── dto           # shared types of the parsed data
```

### TODOS

- [ ] backend view for
  - [ ] logs (including filtering)
  - [ ] cron job list
    - [ ] Ability to execute a cron job from the frontend?
    - [ ] Timeline of the cron executions.
  - [ ] cron job executions
  - [ ] user athentication?
- [ ] simple tracing extension for the logger?
- [ ] interface based kv store
- [x] store the log level as either a char or a smallint in the db for compression
- [ ] ideas from https://medium.com/@tjholowaychuk/apex-log-e8d9627f4a9a
- [ ] Api endpoint for storing event logs (so that they can be previewed)
- [ ] Extending the scheduler
  - [x] Add an optional desciption field.
  - [ ] Add an optional OnError callback function, which executes a specified function if the cron job fails.
  - [ ] Add an util function for checking the consumption of global resources?
  - [ ] Figure out if there's a way to implement remote executions of the jobs
  - [ ] How to map the logs of a cron job to the cron job itself? Meaning, clicking on the cron job in the frontend will show all of the logs from it.
- [ ] Frontend
  - [ ] Figure out how to host the admin view so that by using the syro package there would be an option to expose the admin view.
  - [ ] https://play.axiom.co/axiom-play-qf1k/datasets
  - [ ] https://medium.com/@tjholowaychuk/apex-logs-public-beta-48c683464054

<!--

psql -d postgres

 -->
