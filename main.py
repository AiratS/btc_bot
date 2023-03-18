import json
import matplotlib.pyplot as plt
import pandas as pd

# Configs
EXTRA_ROWS = 600
DATA_FILE = 'data.json'
DATASET_FILE = 'datasets/BTCUSDT-1m-2023-01.csv'

# Open files
f = open('data.json')
json_data = json.load(f)

df = pd.read_csv(DATASET_FILE)
df.columns = [
    'OPEN_TIME',
    'OPEN_PRICE',
    'HIGH_PRICE',
    'LOW_PRICE',
    'CLOSE_PRICE',
    'VOLUME',
    'CLOSE_TIME',
    'QUOTE_ASSET_VOLUME',
    'NUMBER_OF_TRADES',
    'TAKER_BUY_BASE_ASSET_VOLUME',
    'TAKER_BUY_QUOTE_ASSET_VOLUME',
    'IGNORE',
]

# Define column types
df['IGNORE'] = df['IGNORE'].astype('bool')
df['CLOSE_TIME'] = df['CLOSE_TIME'].values.astype(dtype='datetime64[ms]')

TOTAL_ROWS_COUNT = df.shape[0]


def create_plot(data):
    # Find buy and sell rows
    buy_df = df[df['CLOSE_TIME'] == (data['buyTime'] + '.999')]
    sell_df = df[df['CLOSE_TIME'] == (data['sellTime'] + '.999')]

    if buy_df.shape[0] == 0:
        print("No buy Data")
        return

    if sell_df.shape[0] == 0:
        print("No sell Data")
        return

    MIN_INDEX = buy_df.index.values[0] - EXTRA_ROWS
    if MIN_INDEX < 0:
        MIN_INDEX = 0

    MAX_INDEX = sell_df.index.values[0] + EXTRA_ROWS
    if MAX_INDEX > TOTAL_ROWS_COUNT:
        MAX_INDEX = TOTAL_ROWS_COUNT

    # Plot
    plot_df = df[MIN_INDEX:MAX_INDEX]

    ax = plot_df.plot(x='CLOSE_TIME', y='CLOSE_PRICE')

    min_price = plot_df['CLOSE_PRICE'].min()
    max_price = plot_df['CLOSE_PRICE'].max()
    ax.set_ylim(min_price - 10, max_price + 10)

    buy_price = buy_df.iloc[0]['CLOSE_PRICE']
    ax.annotate(
        'BUY: ' + str(buy_price),
        xy=(data['buyTime'], buy_price),
        xytext=(data['buyTime'], buy_price + 5),
        arrowprops=dict(facecolor='red', shrink=0.05)
    )

    sell_price = sell_df.iloc[0]['CLOSE_PRICE']
    ax.annotate(
        'SELL: ' + str(sell_price),
        xy=(data['sellTime'], sell_price),
        xytext=(data['sellTime'], sell_price + 5),
        arrowprops=dict(facecolor='green', shrink=0.05)
    )

    # Create image files
    fig = ax.get_figure()
    fig.savefig('plots/buy_' + str(data['buyId']) + '.png')


for item in json_data:
    create_plot(item)