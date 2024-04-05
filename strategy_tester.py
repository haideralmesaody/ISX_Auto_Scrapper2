#class that includes the function to backtest strategies
import os
from datetime import datetime
import pandas as pd
import numpy as np
from LoggerFunction import Logger  # Import your Logger class
from pandas.tseries.offsets import BDay
import matplotlib.pyplot as plt
import traceback


class StrategyTester:
    def __init__(self):
        self.logger = Logger()

    #function run all the backtest functions in the class
    def backtest_all_strategies(self):
        try:

            #loop throug all the strategies RSIMACD Strategy	RSICMF Strategy	RSI OBV Strategy	OBV Strategy	RSI Strategy	RSI Strategy2	MACD Strategy	CMF Strategy	RSI14_OBV_RoC Strategy	EMA5 PSAR Strategy
            # and apply summarize_strategy_results to them
            strategies = ['RSI Strategy', 'RSI Strategy2', 'RSI14_OBV_RoC Strategy', 'RSIMACD Strategy', 'RSICMF Strategy', 'RSI OBV Strategy', 'OBV Strategy', 'MACD Strategy', 'CMF Strategy', 'EMA5 PSAR Strategy','EMA5 PSAR Strategy2','Rolling Std10 Strategy','Rolling Std50 Strategy']
            #strategies = ['RSI Strategy', 'RSI Strategy2','CMF Strategy with TP']
            for strategy in strategies:
                self.logger.log_or_print(f"Backtesting {strategy}...", level="INFO")
                #self.summarize_strategy_results(strategy)
                self.summarize_strategy_results_liquid(strategy)
                self.logger.log_or_print(f"Backtesting {strategy} complete.", level="INFO")


            
        except Exception as e:
            self.logger.log_or_print(f"Error in backtest_all_strategies method: {e}", level="ERROR")


    #function to taske a strategy name column, loop through all the Strategies_{ticker}.cvs files and get summarize the strategy results to a csv file names Summary_Strategy_{Strategy Column Name}.csv
    def summarize_strategy_results(self, strategy_name):
        """
        Summarizes the backtest results for a given trading strategy across multiple tickers,
        including both closed and open trades at the end of the backtesting period.

        This function iterates through a list of tickers, retrieves the strategy results for each ticker,
        and compiles the outcomes of all trades (buy/sell pairs) into a summary report. It captures details
        of each trade, such as buy date, sell date (if applicable), buy price, sell price (if applicable),
        profit, profit percentage, and gain factor. Open positions at the end of the backtest are also included
        in the summary with relevant details, marked as 'Open' in the 'Trade Status' column.

        Parameters:
        - strategy_name (str): The name of the strategy column in the CSV files. This name is used to identify
        the strategy signals ('Buy', 'Sell', 'Hold') within each ticker's strategy CSV file.

        Outputs:
        - A CSV file named 'Summary_Strategy_{strategy_name}.csv' containing the compiled backtest results
        for all tickers. The file includes columns for ticker, buy date, sell date, buy price, sell price,
        profit, profit percentage, gain factor, and trade status (indicating whether the trade is 'Closed'
        or remains 'Open').

        The function logs various messages to indicate progress, any issues encountered with specific tickers,
        or the absence of strategy files. It also logs a message when the backtest results are successfully saved.

        Note:
        - The function expects a 'TICKERS.csv' file listing the tickers to be backtested.
        - Strategy CSV files are expected to be named in the format 'Strategies_{ticker}.csv' and should
        contain at least the columns 'Date', 'Close', and the strategy signal column as specified by
        `strategy_name`.
        - Open trade calculations assume the last close price available in the dataset as the sell price for profit
        calculation purposes.
        """
        try:
            #tickers_df = pd.read_csv('TICKERS.csv')
            tickers_df = pd.read_csv('TICKERS2.csv')
            backtest_results = []
            for ticker in tickers_df['Ticker']:
                try:
                    strategy_file_path = f'Strategies_{ticker}.csv'
                    if not os.path.exists(strategy_file_path):
                        self.logger.log_or_print(f"No strategy file found for {ticker}. Skipping...", level="WARNING")
                        continue
                    df = pd.read_csv(strategy_file_path, parse_dates=['Date'])
                    df.sort_values('Date', inplace=True)
                    # The strategy column name will have Buy, Sell, Hold Values, use the same conditions
                    df['Signal'] = df[f'{strategy_name}']
                    holding = False
                    for index, row in df.iterrows():
                        if row['Signal'] == 'Buy' and not holding:
                            buy_date = row['Date']
                            buy_price = row['Close']
                            #calculate the buy price with the 0.06% commission
                            buy_price_c = buy_price + (buy_price * 0.0006)
                            holding = True
                        elif row['Signal'] == 'Sell' and holding:
                            sell_date = row['Date']
                            sell_price = row['Close']
                            #calculate the sell price with the 0.06% commission
                            sell_price_c = sell_price - (sell_price * 0.0006)
                            profit = sell_price - buy_price
                            profit_percent = (profit / buy_price) * 100
                            gainfactor = sell_price / buy_price
                            #calculate profit profit percient and gain factor with the commission
                            profit_c = sell_price_c - buy_price_c
                            profit_percent_c = (profit_c / buy_price_c) * 100
                            gainfactor_c = sell_price_c / buy_price_c
                            


                            backtest_results.append({
                                'Ticker': ticker,
                                'Buy Date': buy_date,
                                'Sell Date': sell_date,
                                'Buy Price': buy_price,
                                'Sell Price': sell_price,
                                'Profit': profit,
                                'Profit Percent': profit_percent,
                                'Gain Factor': gainfactor,
                                'Trade Status': 'Closed',
                                'Buy Price Commission': buy_price_c,
                                'Sell Price Commission': sell_price_c,
                                'Profit Commission': profit_c,
                                'Profit Percent Commission': profit_percent_c,
                                'Gain Factor Commission': gainfactor_c




                            })
                            holding = False  # Reset for the next buy-sell cycle

                    # Handle the open position case
                    if holding:
                        # Assuming you want to consider the last available close price for the open position
                        last_close_price = df.iloc[-1]['Close']
                        profit = last_close_price - buy_price
                        profit_percent = (profit / buy_price) * 100
                        gainfactor = last_close_price / buy_price
                        backtest_results.append({
                            'Ticker': ticker,
                            'Buy Date': buy_date,
                            'Sell Date': 'N/A',  # No sell date for open positions
                            'Buy Price': buy_price,
                            'Sell Price': 'N/A',  # No sell price for open positions
                            'Profit': profit,
                            'Profit Percent': profit_percent,
                            'Gain Factor': gainfactor,
                            'Trade Status': 'Open',
                            'Buy Price Commission': buy_price_c,
                            'Sell Price Commission': sell_price_c,
                            'Profit Commission': profit_c,
                            'Profit Percent Commission': profit_percent_c,
                            'Gain Factor Commission': gainfactor_c
                        })
                        self.logger.log_or_print(f"Position in {ticker} remains open at the end of the backtest period.", level="INFO")
                    self.logger.log_or_print(f"Backtesting {ticker} for {strategy_name} complete.", level="INFO")
                except Exception as e:
                    self.logger.log_or_print(f"Error backtesting {ticker}: {e}", level="ERROR")

            if backtest_results:
                self.logger.log_or_print(f"Summarizing backtest results for {strategy_name}...", level="INFO")
                results_df = pd.DataFrame(backtest_results)
                #sort the results by buy date
                results_df = results_df.sort_values(by='Buy Date')
                results_path = f'Summary_Strategy_{strategy_name}.csv'
                self.logger.log_or_print(f"Saving backtest results to {results_path}...", level="INFO")
                results_df.to_csv(results_path, index=False)
                self.logger.log_or_print(f"Backtest results successfully saved to {results_path}.", level="INFO")
            else:
                self.logger.log_or_print("No backtest results to save.", level="WARNING")

        except Exception as e:
            self.logger.log_or_print(f"Error in summarize_strategy_results method: {e}", level="ERROR")


     #function to taske a strategy name column, loop through all the Strategies_{ticker}.cvs files and get summarize the strategy results to a csv file names Summary_Strategy_{Strategy Column Name}.csv
    def summarize_strategy_results_liquid(self, strategy_name):
        """
        Summarizes the backtest results for a given trading strategy across multiple tickers,
        including both closed and open trades at the end of the backtesting period.

        This function iterates through a list of tickers, retrieves the strategy results for each ticker,
        and compiles the outcomes of all trades (buy/sell pairs) into a summary report. It captures details
        of each trade, such as buy date, sell date (if applicable), buy price, sell price (if applicable),
        profit, profit percentage, and gain factor. Open positions at the end of the backtest are also included
        in the summary with relevant details, marked as 'Open' in the 'Trade Status' column.

        Parameters:
        - strategy_name (str): The name of the strategy column in the CSV files. This name is used to identify
        the strategy signals ('Buy', 'Sell', 'Hold') within each ticker's strategy CSV file.

        Outputs:
        - A CSV file named 'Summary_Strategy_{strategy_name}.csv' containing the compiled backtest results
        for all tickers. The file includes columns for ticker, buy date, sell date, buy price, sell price,
        profit, profit percentage, gain factor, and trade status (indicating whether the trade is 'Closed'
        or remains 'Open').

        The function logs various messages to indicate progress, any issues encountered with specific tickers,
        or the absence of strategy files. It also logs a message when the backtest results are successfully saved.

        Note:
        - The function expects a 'TICKERS.csv' file listing the tickers to be backtested.
        - Strategy CSV files are expected to be named in the format 'Strategies_{ticker}.csv' and should
        contain at least the columns 'Date', 'Close', and the strategy signal column as specified by
        `strategy_name`.
        - Open trade calculations assume the last close price available in the dataset as the sell price for profit
        calculation purposes.
        """
        try:
            #tickers_df = pd.read_csv('TICKERS.csv')
            tickers_df = pd.read_csv('TICKERS2.csv')
            backtest_results = []
            #read the liquidity score file
            liquidity_score = pd.read_csv('liquidity_scores.csv')
            for ticker in tickers_df['Ticker']:
                try:
                    #check if Liquidity Score% column is in the liquidity_score dataframe is equal or greater than 2% process it other wise skip it
                    #if liquidity_score[liquidity_score['Ticker'] == ticker]['Liquidity Score%'].values[0] < 0.02:
                    #    self.logger.log_or_print(f"Liquidity Score for {ticker} is less than 2%. Skipping...", level="WARNING")
                    #    continue


                    strategy_file_path = f'Strategies_{ticker}.csv'
                    if not os.path.exists(strategy_file_path):
                        self.logger.log_or_print(f"No strategy file found for {ticker}. Skipping...", level="WARNING")
                        continue
                    df = pd.read_csv(strategy_file_path, parse_dates=['Date'])
                    df.sort_values('Date', inplace=True)
                    # The strategy column name will have Buy, Sell, Hold Values, use the same conditions
                    df['Signal'] = df[f'{strategy_name}']
                    holding = False
                    for index, row in df.iterrows():
                        if row['Signal'] == 'Buy' and not holding:
                            buy_date = row['Date']
                            buy_price = row['Close']
                            #calculate the buy price with the 0.06% commission
                            buy_price_c = buy_price + (buy_price * 0.0006)
                            holding = True
                        elif row['Signal'] == 'Sell' and holding:
                            sell_date = row['Date']
                            sell_price = row['Close']
                            #calculate the sell price with the 0.06% commission
                            sell_price_c = sell_price - (sell_price * 0.0006)
                            profit = sell_price - buy_price
                            profit_percent = (profit / buy_price) * 100
                            gainfactor = sell_price / buy_price
                            #calculate profit profit percient and gain factor with the commission
                            profit_c = sell_price_c - buy_price_c
                            profit_percent_c = (profit_c / buy_price_c) * 100
                            gainfactor_c = sell_price_c / buy_price_c
                            


                            backtest_results.append({
                                'Ticker': ticker,
                                'Buy Date': buy_date,
                                'Sell Date': sell_date,
                                'Buy Price': buy_price,
                                'Sell Price': sell_price,
                                'Profit': profit,
                                'Profit Percent': profit_percent,
                                'Gain Factor': gainfactor,
                                'Trade Status': 'Closed',
                                'Buy Price Commission': buy_price_c,
                                'Sell Price Commission': sell_price_c,
                                'Profit Commission': profit_c,
                                'Profit Percent Commission': profit_percent_c,
                                'Gain Factor Commission': gainfactor_c




                            })
                            holding = False  # Reset for the next buy-sell cycle

                    # Handle the open position case
                    if holding:
                        # Assuming you want to consider the last available close price for the open position
                        last_close_price = df.iloc[-1]['Close']
                        # Calculate the sell price with the 0.06% commission for the open position
                        sell_price_c = last_close_price - (last_close_price * 0.0006)
                        profit = last_close_price - buy_price
                        profit_percent = (profit / buy_price) * 100
                        gainfactor = last_close_price / buy_price
                        # Recalculate profit, profit percent, and gain factor with commission for the open trade
                        profit_c = sell_price_c - buy_price_c
                        profit_percent_c = (profit_c / buy_price_c) * 100
                        gainfactor_c = sell_price_c / buy_price_c

                        backtest_results.append({
                            'Ticker': ticker,
                            'Buy Date': buy_date,
                            'Sell Date': 'Open Trade - Last Close Price Used',  # Indicate open trade and last close price used
                            'Buy Price': buy_price,
                            'Sell Price': last_close_price,  # Show last close price for open trades
                            'Profit': profit,
                            'Profit Percent': profit_percent,
                            'Gain Factor': gainfactor,
                            'Trade Status': 'Open',
                            'Buy Price Commission': buy_price_c,
                            'Sell Price Commission': sell_price_c,  # Reflecting the adjusted close price with commission
                            'Profit Commission': profit_c,
                            'Profit Percent Commission': profit_percent_c,
                            'Gain Factor Commission': gainfactor_c
                        })
                        self.logger.log_or_print(f"Position in {ticker} remains open at the end of the backtest period.", level="INFO")
                    self.logger.log_or_print(f"Backtesting {ticker} for {strategy_name} complete.", level="INFO")
                except Exception as e:
                    self.logger.log_or_print(f"Error backtesting {ticker}: {e}", level="ERROR")

            if backtest_results:
                self.logger.log_or_print(f"Summarizing backtest results for {strategy_name}...", level="INFO")
                results_df = pd.DataFrame(backtest_results)
                #sort the results by buy date
                results_df = results_df.sort_values(by='Buy Date')
                results_path = f'Summary_Strategy_Liquid_{strategy_name}.csv'
                self.logger.log_or_print(f"Saving backtest results to {results_path}...", level="INFO")
                results_df.to_csv(results_path, index=False)
                self.logger.log_or_print(f"Backtest results successfully saved to {results_path}.", level="INFO")
            else:
                self.logger.log_or_print("No backtest results to save.", level="WARNING")

        except Exception as e:
            self.logger.log_or_print(f"Error in summarize_strategy_results method: {e}", level="ERROR")

    #summarize the strategy summarization results
    def summarize_summary_strategy_sheets(self):
        """
        Aggregates the strategy summary data from multiple CSV files and saves 
        the combined results into a single CSV file.
        """
        try:
            # Get all the summary strategy files that match the pattern
            files = [f for f in os.listdir('.') if os.path.isfile(f) and f.startswith('Summary_Strategy_')]
            
            # Initialize a list to hold data for each strategy
            strategy_summaries = []
            
            # Iterate over each file to process and summarize the data
            for file in files:
                try:
                    # Read the strategy CSV file into a DataFrame
                    df = pd.read_csv(file)
                    
                    # Calculate various statistics based on commission-adjusted prices
                    stats = {
                        'Strategy': file.replace('Summary_Strategy_', '').replace('.csv', ''),
                        'Number of Trades': len(df),
                        'Number of Closed Trades': len(df[df['Trade Status'] == 'Closed']),
                        'Number of Open Trades': len(df[df['Trade Status'] == 'Open']),
                        'Number of Losing Trades': len(df[df['Profit Commission'] < 0]),
                        'Number of Winning Trades': len(df[df['Profit Commission'] > 0]),
                        'Winning Trade Percent': len(df[df['Profit Commission'] > 0]) / len(df) * 100,
                        'Average Profit': df['Profit Commission'].mean(),
                        'Average Profit Percent': df['Profit Percent Commission'].mean(),
                        'Average Gain Factor': df['Gain Factor Commission'].mean(),
                        'Product of Gain Factors': df['Gain Factor Commission'].prod(),
                    }
                    
                    # Append the calculated statistics to the list
                    strategy_summaries.append(stats)
                except Exception as e:
                    # Log any errors encountered during processing of a file
                    self.logger.log_or_print(f"Error processing file {file}: {e}", level="ERROR")
            
            # Convert the list of summaries to a DataFrame
            summary_df = pd.DataFrame(strategy_summaries)
            
            # Save the aggregated summary to a CSV file
            summary_df.to_csv('Summary_Strategy_All.csv', index=False)
            self.logger.log_or_print("Aggregated strategy summaries successfully saved to Summary_Strategy_All.csv", level="INFO")
        
        except Exception as e:
            # Log any errors encountered during the aggregation process
            self.logger.log_or_print(f"Error in summarize_summary_strategy_sheets method: {e}", level="ERROR")

            
    def simulate_strategy_results(self):
        files = [f for f in os.listdir('.') if os.path.isfile(f) and f.startswith('Summary_Strategy_')]
        #print the files to be simulated
        self.logger.log_or_print(f"Files to be simulated: {files}", level="INFO")
        if not files:
            self.logger.log_or_print("No strategy summary files found to process.", level="WARNING")
            return

        for file in files:
            try:
                df = pd.read_csv(file)

                if 'Gain Factor Commission' not in df.columns:
                    raise ValueError(f"'Gain Factor Commission' column not found in file {file}")

                df['Simulation Status'] = False
                df['Trade Iteration Number'] = 0
                #total number of trades
                number_of_trades= len(df)
                #number of simulated trades
                number_of_simulated_trades = 0
                #number of iterations 
                i=0

                #while simulated trades < total trades
                while number_of_simulated_trades < number_of_trades:
                    #current iteration number
                    i += 1

                    current_trade_index = df[(df['Simulation Status'] == False) & (df['Trade Iteration Number'] == 0)].index.min()
                    if current_trade_index is None:
                        self.logger.log_or_print(f"No trade found for iteration {i} in file {file}.", level="WARNING")
                        break
                    
                    while current_trade_index is not None:  # Check for None explicitly
                        df.at[current_trade_index, 'Simulation Status'] = True
                        df.at[current_trade_index, 'Trade Iteration Number'] = i
                        df.at[current_trade_index, f'{i}th Iteration'] = df.at[current_trade_index, 'Gain Factor Commission']
                        self.logger.log_or_print(f"Iteration {i}: Processing trade at index {current_trade_index}.", level="INFO")
                        
                        sell_date = df.at[current_trade_index, 'Sell Date']
                        self.logger.log_or_print(f"Current sell date: {sell_date}", level="INFO")
                        
                        next_trades = df[(df['Simulation Status'] == False) & (df['Buy Date'] > sell_date)]
                        current_trade_index = next_trades.index.min() if not next_trades.empty else None
                        #increment the number of simulated trades
                        number_of_simulated_trades += 1

                        

                df.to_csv(f'Simulated_{file}', index=False)
                self.logger.log_or_print(f"Strategy results successfully simulated and saved for {file}.", level="INFO")

            except Exception as e:
                self.logger.log_or_print(f"Unexpected error processing file {file}: {e}", level="ERROR")
                traceback.print_exc()

    def summarize_simulated_strategy_results(self):
        """
        Summarizes the simulated strategy results from files starting with 'Simulated_Summary_Strategy_'
        and saves the summary to a CSV file named 'Summarized_Simulations.csv'. The summary includes
        total trades, open trades, winning trades, losing trades, the winning trade percentage,
        the simulation result (average product of gain factors across iterations), and the number of iterations
        required for simulation.
        """
        try:
            files = [f for f in os.listdir('.') if os.path.isfile(f) and f.startswith('Simulated_Summary_Strategy_')]
            
            if not files:
                self.logger.log_or_print("No simulated strategy summary files found to process.", level="WARNING")
                return

            summaries = []
            for file in files:
                try:
                    df = pd.read_csv(file)

                    total_trades = len(df)
                    number_of_iterations = df['Trade Iteration Number'].max()

                    gain_factors = df.filter(regex='^\d+th Iteration$', axis=1)

                    dynamic_stats = {}
                    # Calculate dynamic fields first for each iteration
                    for i in range(1, number_of_iterations + 1):
                        if f'{i}th Iteration' in gain_factors.columns:
                            product = gain_factors[f'{i}th Iteration'].prod()
                            dynamic_stats[f'Product of Gain Factors - Iteration {i}'] = product
                            dynamic_stats[f'Number of Trades - Iteration {i}'] = len(df[df['Trade Iteration Number'] == i])

                    # Calculate 'Simulation Result' dynamically
                    simulation_result = sum(dynamic_stats.get(f'Product of Gain Factors - Iteration {i}', 0) for i in range(1, number_of_iterations + 1)) / number_of_iterations if number_of_iterations > 0 else 0

                    # Construct stats with both static and dynamic fields
                    stats = {
                        'File': file, 
                        'Number of Trades Total': total_trades, 
                        'Open Trades': len(df[df['Trade Status'] == 'Open']),
                        'Number of Winning Trades': len(df[df['Profit Commission'] > 0]),
                        'Number of Losing Trades': len(df[df['Profit Commission'] < 0]),
                        'Winning Trade Percent': (len(df[df['Profit Commission'] > 0]) / total_trades * 100 if total_trades > 0 else 0),
                        'Simulation Result': simulation_result,
                        'Number of Iterations': number_of_iterations,
                        **dynamic_stats  # Incorporate dynamic fields
                    }

                    summaries.append(stats)

                except Exception as e:
                    self.logger.log_or_print(f"Unexpected error processing file {file}: {e}", level="ERROR")
                    traceback.print_exc()

            summary_df = pd.DataFrame(summaries)
            summary_df.to_csv('Summarized_Simulations.csv', index=False)
            self.logger.log_or_print("Simulated strategy results successfully summarized and saved to Summarized_Simulations.csv.", level="INFO")
        except Exception as e:
            self.logger.log_or_print(f"Unexpected error in summarize_simulated_strategy_results: {e}", level="ERROR")   
            traceback.print_exc()
