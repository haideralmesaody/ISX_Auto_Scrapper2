import argparse
import pandas as pd
from data_fetcher import DataFetcher
from data_calculator import DataCalculator
from file_manager import FileManager
from liquidity_calculator import LiquidityCalculator
from strategy_tester import StrategyTester
from strategies import Strategies
import time
from LoggerFunction import Logger  # Import your Logger class
from data_fetcher_ase import DataFetcherASE
from data_calculator_ase import DataCalculatorASE
from file_manager_ase import FileManagerASE
from liquidity_calculator_ase import LiquidityCalculatorASE
from strategies_ase import Strategies_ASE
from strategy_tester_ase import StrategyTester_ASE
from app_config import EDGE_DRIVER_PATH

def main():
    parser = argparse.ArgumentParser(description='Extract data from a web page.')
    parser.add_argument('--mode', type=str, choices=['single', 'auto' ,'liquidity' , 'breakout','strategies','backtest','single_ase','auto_ase','liquidity_ase','strategies_ase','backtest_ase'], required=True, help='The mode to run the script in.')
    args = parser.parse_args()

    # Create a logger object
    logger = Logger()


    #ISX Data Collection
    data_fetcher = DataFetcher(driver_path=EDGE_DRIVER_PATH)
    data_calculator = DataCalculator()
    file_manager = FileManager()
    #Analysis and Strategy for ISX
    liquidity_calculator = LiquidityCalculator()
    strategiesfunction = Strategies()
    strategytester = StrategyTester()
    #ASE Data Collection
    data_fetcher_ase = DataFetcherASE(driver_path=EDGE_DRIVER_PATH)
    data_calculator_ase = DataCalculatorASE()
    file_manager_ase = FileManagerASE()
    #analysis and strategy for ASE
    liquidity_calculator_ase = LiquidityCalculatorASE()
    strategiesfunction_ase = Strategies_ASE()
    strategytester_ase = StrategyTester_ASE()
    #ISX Modes
    if args.mode == 'single':
        ticker = input('Enter the ticker: ')
        data_fetcher.fetch_data(ticker)
        data_calculator.calculate_all(ticker)
        file_manager.generate_report(ticker)
        
    elif args.mode == 'auto':
        tickers_df = pd.read_csv('TICKERS.csv')
        for ticker in tickers_df['Ticker']:
            print(f'Calling Data Fethcer for {ticker}...')
            data_fetcher.fetch_data(ticker)
            print(f'Calling Data Calculator for {ticker}...')
            data_calculator.calculate_all(ticker)
            #print(f'Calling File Manager for {ticker}...')
            file_manager.generate_report(ticker)
    elif args.mode == 'liquidity':
        liquidity_calculator.calculate_liquidity_score()
    elif args.mode == 'strategies':
        strategiesfunction.apply_strategies_and_save()
        strategiesfunction.apply_alternative_strategy_states()
        strategiesfunction.summarize_strategy_actions()
    elif args.mode == 'backtest':
        strategytester.backtest_all_strategies()
    #ASE Modes
    elif args.mode == 'single_ase':
        ticker = input('Enter the ticker: ')
        data_fetcher_ase.fetch_data(ticker)
        data_calculator_ase.calculate_all(ticker)
        file_manager_ase.generate_report(ticker)
    elif args.mode == 'auto_ase':
        tickers_df = pd.read_csv('TICKERS_ASE.csv')
        # Total number of tickers
        tickers_count = len(tickers_df['Ticker'])
        current_ticker = 0
        # Start time of the loop
        start_time = time.time()
        for ticker in tickers_df['Ticker']:
            # Start ticker processing
            ticker_start_time = time.time()

            current_ticker += 1
            logger.log_or_print(f'Processing ticker {current_ticker} of {tickers_count}')
            logger.log_or_print(f'Calling Data Fethcer for {ticker}...')
            data_fetcher_ase.fetch_data(ticker)
            logger.log_or_print(f'Calling Data Calculator for {ticker}...')
            data_calculator_ase.calculate_all(ticker)
            # Calculate processing time for the ticker
            ticker_processing_time = time.time() - ticker_start_time
            logger.log_or_print(f'Time taken to process {ticker}: {ticker_processing_time:.2f} seconds')

            # Estimate time to finish based on the average time per ticker
            elapsed_time = time.time() - start_time
            average_time_per_ticker = elapsed_time / current_ticker
            remaining_tickers = tickers_count - current_ticker
            estimated_remaining_time = average_time_per_ticker * remaining_tickers
            logger.log_or_print(f'Estimated time remaining: {estimated_remaining_time:.2f} seconds')

    elif args.mode == 'liquidity_ase':
        liquidity_calculator_ase.calculate_liquidity_score()  

    elif args.mode == 'strategies_ase':
        strategiesfunction_ase.apply_strategies_and_save()
        strategiesfunction_ase.apply_alternative_strategy_states()
        strategiesfunction_ase.summarize_strategy_actions()
    elif args.mode == 'backtest_ase':
        strategytester_ase.backtest_all_strategies()
    else:
        print('Invalid mode.')


if __name__ == "__main__":
    main()