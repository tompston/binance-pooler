get all symbols
# https://api.binance.com/api/v3/exchangeInfo


# 1min query data
# https://binance-docs.github.io/apidocs/spot/en/#kline-candlestick-data
https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1m&startTime=1633833600000&endTime=1633833900000&limit=1000


# https://api.binance.com/api/v3/ticker/price

#
https://api.binance.com/sapi/v1/convert/exchangeInfo
{
  "fromAsset": "1INCH",
  "toAsset": "BTC",
  "fromAssetMinAmount": "0.029",
  "fromAssetMaxAmount": "61000",
  "toAssetMinAmount": "0.00000026",
  "toAssetMaxAmount": "0.58"
},


# symbol price ticker
https://binance-docs.github.io/apidocs/spot/en/#symbol-price-ticker
https://api.binance.com/api/v3/ticker/price

get a list of all symbols