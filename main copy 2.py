import argparse
import pandas as pd
from data_fetcher import DataFetcher
from app_config import EDGE_DRIVER_PATH

def main():
    parser = argparse.ArgumentParser(description='Extract data from a web page.')
    parser.add_argument('--mode', type=str, choices=['single', 'auto'], required=True, help='The mode to run the script in.')
    args = parser.parse_args()

    data_fetcher = DataFetcher(driver_path=EDGE_DRIVER_PATH)

    if args.mode == 'single':
        ticker = input('Enter the ticker: ')
        df = data_fetcher.fetch_data(ticker, 100)
        df.to_csv(f'{ticker}.csv', index=False)
    elif args.mode == 'auto':
        tickers_df = pd.read_csv('TICKERS.csv')
        for ticker in tickers_df['Ticker']:
            df = data_fetcher.fetch_data(ticker, 100)
            df.to_csv(f'{ticker}.csv', index=False)

if __name__ == "__main__":
    main()