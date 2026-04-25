Timeseries data scraper example, implementing the ideas mentioned in [this tutorial](https://tompston.pages.dev/writing/2024-06-29-everything-about-timeseries-data-scraping).

Scrapes spot data from binance api and stores it in a mongodb database.

## Dependencies

- Go (min version 1.23)
- Running instance of mongodb. Default port, no auth. [installation link](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-on-ubuntu/)
- Reflex (for rebuilding the go app on changes) [(installation link)](https://github.com/cespare/reflex)

## Running the app

```bash
chmod +x run.sh

# start the binance-pooler/cmd/pooler app
./run.sh pooler
# run tests for the project (will be written under mongodb database called `test`)
./run.sh test
```
