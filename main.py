import argparse
import pandas as pd
from data_fetcher import DataFetcher
from data_calculator import DataCalculator
from data_calculator_nums import DataCalculatorNums
from file_manager import FileManager
from liquidity_calculator import LiquidityCalculator
from strategy_tester import StrategyTester
from strategies import Strategies

import time
from LoggerFunction import Logger  # Import your Logger class

from app_config import EDGE_DRIVER_PATH

def main():
    parser = argparse.ArgumentParser(description='Extract data from a web page.')
    parser.add_argument('--mode', type=str, choices=['single', 'auto' ,'liquidity' , 'breakout','strategies','backtest','simulate', 'calculate_num','train_tensor', 'stock_predictor','predict_close_price'], required=True, help='The mode to run the script in.')
    args = parser.parse_args()

    # Create a logger object
    logger = Logger()


    #ISX Data Collection
    data_fetcher = DataFetcher(driver_path=EDGE_DRIVER_PATH)
    data_calculator = DataCalculator()
    data_calculator_nums = DataCalculatorNums()
    file_manager = FileManager()
    #Analysis and Strategy for ISX
    liquidity_calculator = LiquidityCalculator()
    strategiesfunction = Strategies()
    strategytester = StrategyTester()

    #Tensor Model



    #ISX Modes
    if args.mode == 'single':
        ticker = input('Enter the ticker: ')
        data_fetcher.fetch_data(ticker)
        #data_calculator.calculate_all(ticker)
        #file_manager.generate_report(ticker)
        
    elif args.mode == 'auto':
        #print the remaining time hh:mm:ss
        tickers_df = pd.read_csv('TICKERS.csv')

        num_tickers = len(tickers_df['Ticker'])
        for ticker in tickers_df['Ticker']:
            #processing ticker number x of num_tickers
            print(f'Processing {ticker} ({tickers_df.index[tickers_df["Ticker"] == ticker].tolist()[0]+1}/{num_tickers})')

            data_fetcher.fetch_data(ticker)
            data_calculator.calculate_all(ticker)
            file_manager.generate_report(ticker)

        liquidity_calculator.calculate_liquidity_score()
        strategiesfunction.apply_strategies_and_save()
        strategiesfunction.apply_alternative_strategy_states()
        strategiesfunction.summarize_strategy_actions()
        strategytester.backtest_all_strategies()
        strategytester.simulate_strategy_results()
        strategytester.summarize_simulated_strategy_results()
 
    elif args.mode == 'liquidity':
        liquidity_calculator.calculate_liquidity_score()
    elif args.mode == 'strategies':
        strategiesfunction.apply_strategies_and_save()
        strategiesfunction.apply_alternative_strategy_states()
        strategiesfunction.summarize_strategy_actions()
    elif args.mode == 'backtest':
        strategytester.backtest_all_strategies()
        strategytester.summarize_simulated_strategy_results()
        strategytester.summarize_summary_strategy_sheets()
    elif args.mode == 'simulate':
        strategytester.simulate_strategy_results()
        strategytester.summarize_simulated_strategy_results()
    elif args.mode == 'calculate_num':
        #Calculate all the indicators withut description to produce Indicators2_{ticker}.csv files
        tickers_df = pd.read_csv('TICKERS.csv')
        #number of ticker to process
        num_tickers = len(tickers_df['Ticker'])
        for ticker in tickers_df['Ticker']:        
            data_calculator_nums.calculate_all(ticker)

    else:
        print('Invalid mode.')


if __name__ == "__main__":
    main()