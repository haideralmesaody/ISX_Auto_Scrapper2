#class to summarize {Ticker}_Indicators.csv files and give summary report
import os

from datetime import datetime
import pandas as pd
import pandas_ta as ta
from bs4 import BeautifulSoup
from selenium.common.exceptions import UnexpectedAlertPresentException
from selenium.webdriver.common.by import By
from selenium.webdriver.edge.service import Service
from selenium.webdriver.edge.webdriver import WebDriver as EdgeDriver
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait
from selenium.common.exceptions import TimeoutException
from selenium.common.exceptions import NoAlertPresentException
from app_config import (
    LOGGING_CONFIG, EDGE_DRIVER_PATH, BASE_URL, DEFAULT_DATE,
    TABLE_SELECTOR, WEBDRIVER_WAIT_TIME, DEFAULT_SMA_PERIOD,
    DEFAULT_ROW_COUNT, EXCEL_ENGINE
)
from PyQt5.QtCore import QObject, pyqtSignal, pyqtSlot
from LoggerFunction import Logger  # Import your Logger class
import numpy as np
import pandas_ta as ta


class Strategies():

    def __init__(self):
        self.logger = Logger()
    #Apply Alternative States to Strategies 
    def apply_strategies_and_save(self):
        """
        Applies trading indicators and strategies to each ticker's data for the past 12 months,
        then saves the results to a Strategies_{ticker}.csv file.

        This method reads trading data from indicators_{ticker}.csv files, filters data to the last 12 months,
        applies predefined trading strategies, and saves the modified DataFrame with applied strategies
        to Strategies_{ticker}.csv files. It also handles cases where indicator files are missing or empty
        and logs all significant actions and errors.
        """
        try:
            tickers_df = pd.read_csv('TICKERS.csv')
            if tickers_df.empty:
                self.logger.log_or_print('TICKERS.csv is empty. No tickers to process.', level='WARNING')
                return

            for ticker in tickers_df['Ticker']:
                file_path = f'indicators_{ticker}.csv'
                if not os.path.exists(file_path):
                    self.logger.log_or_print(f'indicators_{ticker}.csv does not exist', level='ERROR')
                    continue

                try:
                    df = pd.read_csv(file_path, parse_dates=['Date'])
                    # Keep only the rows within the previous 12 months
                    one_year_ago = datetime.now() - pd.DateOffset(years=1)
                    df = df[df['Date'] > one_year_ago]

                    if df.empty:
                        self.logger.log_or_print(f"No trading data for {ticker} in the past 12 months.", level="WARNING")
                        continue

                    # Filter and apply strategies to the DataFrame
                    df = self.filter_and_apply_strategies(df)
                    df = self.TradingStrategies(df)

                    # Save the DataFrame to a new file
                    strategies_file_path = f'Strategies_{ticker}.csv'
                    df.to_csv(strategies_file_path, index=False)
                    self.logger.log_or_print(f"Trading strategies successfully added and saved for {ticker}.", level="INFO")

                except Exception as e:
                    self.logger.log_or_print(f'Error processing {ticker}: {e}', level='ERROR')

            self.logger.log_or_print('Trading strategies successfully added and saved for all processed tickers.', level='INFO')

        except Exception as e:
            self.logger.log_or_print(f'Error applying trading strategies: {e}', level='ERROR')

    def filter_and_apply_strategies(self, df):
        """
        Filters columns relevant to trading strategies and applies strategies on a DataFrame.

        Parameters:
        - df (DataFrame): DataFrame containing ticker trading data.

        Returns:
        - DataFrame: Modified DataFrame with trading strategies applied.
        """
        relevant_columns = [
            'Date', 'Close', 'Open', 'High', 'Low', 'Change', 'Change%', 'T.Shares', 
            'Volume', 'SMA10', 'SMA50', 'SMA200', 'RSI_14', 'CMF_20', 'MACD_12_26_9', 
            'MACDs_12_26_9', 'MACDh_12_26_9', 'OBV_SMA_Diff', 'EMA5', 'EMA10', 'EMA20', 
            'EMA50', 'PSARl_0.02_0.2', 'OBV_RoC','ATR'
        ]
        # Ensure all relevant columns are present
        missing_columns = [col for col in relevant_columns if col not in df.columns]
        if missing_columns:
            self.logger.log_or_print(f"Missing columns in DataFrame: {missing_columns}", level="ERROR")
            return df

        df = df[relevant_columns]
        # Placeholder for strategy application logic
        # Example: df = self.TradingStrategies(df)
        return df

 
    def TradingStrategies(self, df):
        """
        Applies predefined trading strategies to each row of a DataFrame. This method initializes several
        strategy columns with default values and applies specific trading logic to update these values based
        on trading indicators and conditions.

        Parameters:
        - df (DataFrame): The DataFrame to which trading strategies will be applied. It is expected to contain
                          necessary columns for each strategy's indicators (e.g., RSI, MACD values).

        Returns:
        - DataFrame: The modified DataFrame with trading strategies applied, including new columns for each strategy
                     indicating 'Buy', 'Sell', or 'Hold' signals.
        
        Raises:
        - Exception: Propagates exceptions that may occur during the application of strategies, with an error message
                     logged to the provided logger object.
        """
        if df is None or df.empty:
            raise ValueError("The input DataFrame is either None or empty.")


        
        try:
            # apply RSI Strategy
            df['RSI Strategy'] = 'Hold'
            df= self.apply_rsi_with_trailing_stop(df, atr_multiplier=1)
            # apply RSI Strategy2
            df['RSI Strategy2'] = 'Hold'
            df= self.apply_rsi2_with_trailing_stop(df, atr_multiplier=2)
            # apply RSI14_OBV_RoC Strategy
            df['RSI14_OBV_RoC Strategy'] = 'Hold'
            df= self.apply_rsi14_obv_roc_with_trailing_stop(df, atr_multiplier=1)
            # apply RSIMACD Strategy
            df['RSIMACD Strategy'] = 'Hold'
            df= self.apply_rsimacd_with_trailing_stop(df, atr_multiplier=1)
            # apply RSICMF Strategy
            df['RSICMF Strategy'] = 'Hold'
            df= self.apply_rsicmf_with_trailing_stop(df, atr_multiplier=1)
            # apply RSI OBV Strategy
            df['RSI OBV Strategy'] = 'Hold'
            df= self.apply_rsi_obv_with_trailing_stop(df, atr_multiplier=1)
            # apply OBV Strategy
            df['OBV Strategy'] = 'Hold'
            df= self.apply_obv_with_trailing_stop(df, atr_multiplier=1)
            # apply MACD Strategy
            df['MACD Strategy'] = 'Hold'
            df= self.apply_macd_with_trailing_stop(df, atr_multiplier=1)
            # apply CMF Strategy
            df['CMF Strategy'] = 'Hold'
            df= self.apply_cmf_with_trailing_stop(df, atr_multiplier=1)
            # apply EMA5 PSAR Strategy
            df['EMA5 PSAR Strategy'] = 'Hold'
            df= self.apply_ema5_psar_with_trailing_stop(df, atr_multiplier=1)



            self.logger.log_or_print('Trading strategies successfully added.', level='INFO')
            return df

        except Exception as e:
            self.logger.log_or_print(f'Error adding trading strategies: {e}', level='ERROR')
            raise







    def apply_alternative_strategy_states(self):
        """
        Apply alternative strategy states to strategy CSV files for different tickers based on specified strategy columns.
        
        This function searches for CSV files in the current directory with the pattern 'Strategies_{ticker}.csv'.
        For each CSV file found, it checks if the specified strategy columns exist, creates an alternative strategy column
        for each, and applies logic to populate these alternative strategy columns based on the current and previous states.
        """
        try:
            tickers_df = pd.read_csv('TICKERS.csv')
            # List of all strategies columns
            strategies = [
                'RSIMACD Strategy', 'RSICMF Strategy', 'RSI OBV Strategy', 
                'OBV Strategy', 'RSI Strategy', 'RSI Strategy2', 
                'MACD Strategy', 'MACD Strategy2','CMF Strategy','RSI14_OBV_RoC Strategy' , 'EMA5 PSAR Strategy','EMA5 EMA10 Strategy','EMA5 EMA20 Strategy','EMA10 EMA20 Strategy','EMA10 EMA50 Strategy','EMA20 EMA50 Strategy'
                ,'EMA5 EMA10 OBV_RoC Strategy','EMA5 EMA20 OBV_RoC Strategy','EMA10 EMA20 OBV_RoC Strategy','EMA10 EMA50 OBV_RoC Strategy','EMA20 EMA50 OBV_RoC Strategy']   
            
            for strategy_column in strategies:
                for ticker in tickers_df['Ticker']:
                    file_path = f'Strategies_{ticker}.csv'
                    if not os.path.exists(file_path):
                        self.logger.log_or_print(f'File {file_path} does not exist',level='ERROR')

                        continue

                    strategies_df = pd.read_csv(file_path)
                    if strategy_column not in strategies_df.columns:
                        self.logger.log_or_print(f'Column {strategy_column} does not exist in {file_path}',level='WARNING')
                        continue  # Skip to the next strategy or file

                    alt_strategy_column = f'{strategy_column} Alt'
                    strategies_df[alt_strategy_column] = 'Monitor-Monitor'  # Default state
                    self.logger.log_or_print(f'Applying Alternative Strategy States for {ticker} using {strategy_column}...',level='INFO')

                    for index in range(2, len(strategies_df)):
                        prev_alt_state = strategies_df.at[index - 1, alt_strategy_column]
                        current_state = strategies_df.at[index, strategy_column]
                        next_state = self.determine_next_state(prev_alt_state, current_state)
                        strategies_df.at[index, alt_strategy_column] = next_state

                    strategies_df.to_csv(file_path, index=False)
                    self.logger.log_or_print(f'Alternative strategy states successfully applied and saved for {ticker}.',level='INFO')

        except pd.errors.EmptyDataError:
            self.logger.log_or_print('No data found in TICKERS.csv', level='ERROR')
        except pd.errors.ParserError:
            self.logger.log_or_print('Error parsing TICKERS.csv', level='ERROR')
        except Exception as e:
            self.logger.log_or_print(f'An unexpected error occurred: {e}', level='ERROR')


    def determine_next_state(self, prev_alt_state, current_state):
        """
        Determine the next alternative state based on previous alternative state and current strategy state.

        :param prev_alt_state: The previous state in the alternative strategy column.
        :param current_state: The current state in the original strategy column.
        :return: The next state to be applied to the alternative strategy column.
        """
        # Placeholder for actual state determination logic
        # Replace the following logic with your specific rules for determining the next state
        if prev_alt_state == 'Monitor-Monitor' and current_state == 'Buy':
            return 'Buy'
        elif prev_alt_state == 'Monitor-Monitor' and current_state == 'Hold':
            return 'Monitor-Monitor'
        elif prev_alt_state == 'Monitor-Monitor' and current_state == 'Sell':
            return 'Sell-Monitor'
        
        elif prev_alt_state == 'Buy' and current_state == 'Sell':
            return 'Sell'
        elif prev_alt_state == 'Buy' and current_state == 'Hold':
            return 'Hold-Hold'
        elif prev_alt_state == 'Buy' and current_state == 'Buy':
            return 'Buy-Hold'
        
        
        elif prev_alt_state == 'Hold-Hold' and current_state == 'Sell':
            return 'Sell'
        elif prev_alt_state == 'Hold-Hold' and current_state == 'Hold':
            return 'Hold-Hold'
        elif prev_alt_state == 'Hold-Hold' and current_state == 'Buy':
            return 'Buy-Hold'
        
        elif prev_alt_state == 'Sell' and current_state == 'Buy':
            return 'Buy'
        elif prev_alt_state == 'Sell' and current_state == 'Hold':
            return 'Monitor-Monitor'
        elif prev_alt_state == 'Sell' and current_state == 'Sell':
            return 'Sell-Monitor'
        
        elif prev_alt_state == 'Sell-Monitor' and current_state == 'Buy':
            return 'Buy'
        elif prev_alt_state == 'Sell-Monitor' and current_state == 'Hold':
            return 'Monitor-Monitor'
        elif prev_alt_state == 'Sell-Monitor' and current_state == 'Sell':
            return 'Sell-Monitor'
        
        elif prev_alt_state == 'Buy-Hold' and current_state == 'Buy':
            return 'Buy-Hold'
        elif prev_alt_state == 'Buy-Hold' and current_state == 'Hold':
            return 'Hold-Hold'
        elif prev_alt_state == 'Buy-Hold' and current_state == 'Sell':
            return 'Sell'
        
    #summarize strategies actions, go through all the Strategies_{ticker}.cvs files and get summarize the strategy actionlist by getting the last row of each file and add the ticker name to teh start of the row and save the results to date_strategies_action_summary.csv
    def summarize_strategy_actions(self):
        try:
            tickers_df=pd.read_csv('TICKERS.csv')
            strategies_actions = []
            for ticker in tickers_df['Ticker']:
                try:
                    strategy_file_path = f'Strategies_{ticker}.csv'
                    if not os.path.exists(strategy_file_path):
                        self.logger.log_or_print(f"No strategy file found for {ticker}. Skipping...", level="WARNING")
                        continue
                    #geat column names from the first file read
                    
                    if not strategies_actions:
                        df = pd.read_csv(strategy_file_path, parse_dates=['Date'])
                        strategies_actions.append(df.columns)
                        #add the ticker column to the column names first column
                        strategies_actions[0] = np.insert(strategies_actions[0], 0, 'Ticker')
                    df = pd.read_csv(strategy_file_path, parse_dates=['Date'])
                    df.sort_values('Date', inplace=True)
                    last_row = df.iloc[-1]
                    last_row = last_row.to_list()
                    last_row.insert(0, ticker)
                    strategies_actions.append(last_row)
                except Exception as e:
                    self.logger.log_or_print(f"Error summarizing strategy actions for {ticker}: {e}", level="ERROR")
            if strategies_actions:
                results_df = pd.DataFrame(strategies_actions)
                results_path = f'{datetime.now().strftime("%Y-%m-%d")}_strategies_action_summary.csv'
                results_df.to_csv(results_path, index=False)
                self.logger.log_or_print(f"Strategy actions summary successfully saved to {results_path}.", level="INFO")
            else:
                self.logger.log_or_print("No strategy actions summary to save.", level="WARNING")
        except Exception as e:
            self.logger.log_or_print(f"Error in summarize_strategy_actions method: {e}", level="ERROR")

    ## Detailed Strategies Function
    #1. RSIMACD Strategy

    def apply_rsimacd_with_trailing_stop(self, df, atr_multiplier=1):
        """
        Applies the RSIMACD trading strategy along with a trailing stop loss mechanism,
        enhancing it to require both MACD and MACD signal line to be below zero for buy conditions
        and above zero for sell conditions.

        Parameters:
        - df (DataFrame): The DataFrame containing the stock data, including 'RSI_14', 'MACDh_12_26_9', 'MACD_12_26_9' (MACDk), and 'MACDs_12_26_9'.
        - atr_multiplier (float): The multiplier of ATR to set the trailing stop loss.

        Adds a 'Trade_Signal' column to the DataFrame indicating 'Buy', 'Sell', or 'Hold' signals
        and manages a trailing stop loss for open positions.
        """
        if df is None or df.empty:
            self.logger.log_or_print("DataFrame is empty or None", level="ERROR")
            return

        if 'ATR' not in df.columns or 'MACD_12_26_9' not in df.columns or 'MACDs_12_26_9' not in df.columns:
            self.logger.log_or_print("Required columns (ATR, MACD_12_26_9, MACDs_12_26_9) are missing from DataFrame", level="ERROR")
            return

        try:
            trailing_stop_loss = None
            trade_status = 'none'  # Possible values: 'none', 'open'

            for i, row in df.iterrows():
                if trade_status == 'open':
                    current_stop_loss = row['Close'] - (row['ATR'] * atr_multiplier)
                    if row['Close'] > df.loc[i - 1, 'Close']:  # Price moved in our favor
                        trailing_stop_loss = max(trailing_stop_loss, current_stop_loss)
                    if row['Close'] <= trailing_stop_loss:
                        df.at[i, 'RSI'] = 'Sell'
                        trade_status = 'none'
                        continue

                # Enhancing Buy and Sell conditions with MACDk and MACDs requirements
                buy_condition = (row['RSI_14'] < 35) and (row['MACDh_12_26_9'] > 0) and \
                                (row['MACD_12_26_9'] < 0) and (row['MACDs_12_26_9'] < 0)
                sell_condition = (row['RSI_14'] > 65) and (row['MACDh_12_26_9'] < 0) and \
                                 (row['MACD_12_26_9'] > 0) and (row['MACDs_12_26_9'] > 0)

                if buy_condition:
                    df.at[i, 'RSIMACD Strategy'] = 'Buy'
                    trade_status = 'open'
                    trailing_stop_loss = row['Close'] - (row['ATR'] * atr_multiplier)
                elif sell_condition:
                    df.at[i, 'RSIMACD Strategy'] = 'Sell'
                    trade_status = 'none'
                else:
                    df.at[i, 'RSIMACD Strategy'] = 'Hold'

            self.logger.log_or_print("RSIMACD Strategy with Enhanced Conditions and Trailing Stop Loss applied successfully.", level="INFO")
        except Exception as e:
            self.logger.log_or_print(f"An error occurred while applying RSIMACD Strategy: {e}", level="ERROR")
            raise  # Optionally re-raise the exception after logging for external handling
        #log the results how many buy and sell signals were generated
        buy_signals = df[df['RSIMACD Strategy'] == 'Buy'].shape[0]
        sell_signals = df[df['RSIMACD Strategy'] == 'Sell'].shape[0]
        self.logger.log_or_print(f"RSIMACD Strategy generated {buy_signals} Buy signals and {sell_signals} Sell signals.", level="INFO")
        return df
    #2. RSICMF Strategy
    def apply_rsicmf_with_trailing_stop(self, df, atr_multiplier=1):
        """
        Applies the RSICMF trading strategy along with a trailing stop loss mechanism,
        based on RSI and Chaikin Money Flow (CMF) indicators, enhanced with an ATR-based trailing stop loss.

        Parameters:
        - df (DataFrame): The DataFrame containing the stock data, including 'RSI_14', 'CMF_20', 'High', 'Low', 'Close'.
        - atr_multiplier (float): The multiplier of ATR to set the trailing stop loss distance.

        Enhancements:
        - The function checks that 'CMF_20' complies with the buy/sell condition alongside 'RSI_14'.
        - Manages a trailing stop loss for open positions based on the specified ATR multiplier.
        """
        if df is None or df.empty:
            self.logger.log_or_print("DataFrame is empty or None", level="ERROR")
            return

        required_columns = ['RSI_14', 'CMF_20', 'High', 'Low', 'Close']
        if not all(column in df.columns for column in required_columns):
            self.logger.log_or_print("One or more required columns are missing from DataFrame", level="ERROR")
            return

        try:
            df['ATR'] = ta.atr(df['High'], df['Low'], df['Close'], length=14)  # Calculate ATR for trailing stop loss
            trailing_stop_loss = None
            trade_status = 'none'  # Possible values: 'none', 'open'

            for i, row in df.iterrows():
                current_price = row['Close']
                if trade_status == 'open':
                    current_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                    trailing_stop_loss = max(trailing_stop_loss, current_stop_loss) if current_price > df.loc[i - 1, 'Close'] else trailing_stop_loss
                    if current_price <= trailing_stop_loss:
                        df.at[i, 'RSICMF Strategy'] = 'Sell'
                        trade_status = 'none'
                        continue

                buy_condition = row['RSI_14'] < 35 and row['CMF_20'] > 0
                sell_condition = row['RSI_14'] > 65 and row['CMF_20'] < 0

                if buy_condition:
                    df.at[i, 'RSICMF Strategy'] = 'Buy'
                    trade_status = 'open'
                    trailing_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                elif sell_condition:
                    df.at[i, 'RSICMF Strategy'] = 'Sell'
                    trade_status = 'none'
                else:
                    df.at[i, 'RSICMF Strategy'] = 'Hold'

            self.logger.log_or_print("RSICMF Strategy with Trailing Stop Loss applied successfully.", level="INFO")
        except Exception as e:
            self.logger.log_or_print(f"An error occurred while applying RSICMF Strategy: {e}", level="ERROR")
            raise  # Optionally re-raise the exception for external handling

        return df
    
    #3. RSIOBV Strategy
    def apply_rsi_obv_with_trailing_stop(self, df, atr_multiplier=1):
        """
        Applies the RSI OBV trading strategy along with a trailing stop loss mechanism,
        based on RSI and On-Balance Volume (OBV) moving average difference indicators,
        enhanced with an ATR-based trailing stop loss.

        Parameters:
        - df (DataFrame): The DataFrame containing the stock data, including 'RSI_14', 'OBV_SMA_Diff', 'High', 'Low', 'Close'.
        - atr_multiplier (float): The multiplier of ATR to set the trailing stop loss distance.

        Enhancements:
        - The function checks that 'OBV_SMA_Diff' complies with the buy/sell condition alongside 'RSI_14'.
        - Manages a trailing stop loss for open positions based on the specified ATR multiplier.
        """
        if df is None or df.empty:
            self.logger.log_or_print("DataFrame is empty or None", level="ERROR")
            return

        required_columns = ['RSI_14', 'OBV_SMA_Diff', 'High', 'Low', 'Close']
        if not all(column in df.columns for column in required_columns):
            self.logger.log_or_print("One or more required columns are missing from DataFrame", level="ERROR")
            return

        try:
            df['ATR'] = ta.atr(df['High'], df['Low'], df['Close'], length=14)  # Calculate ATR for trailing stop loss
            trailing_stop_loss = None
            trade_status = 'none'  # Possible values: 'none', 'open'

            for i, row in df.iterrows():
                current_price = row['Close']
                if trade_status == 'open':
                    current_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                    trailing_stop_loss = max(trailing_stop_loss, current_stop_loss) if current_price > df.loc[i - 1, 'Close'] else trailing_stop_loss
                    if current_price <= trailing_stop_loss:
                        df.at[i, 'RSI OBV Strategy'] = 'Sell'
                        trade_status = 'none'
                        continue

                buy_condition = row['RSI_14'] < 35 and row['OBV_SMA_Diff'] > 0
                sell_condition = row['RSI_14'] > 65 and row['OBV_SMA_Diff'] < 0

                if buy_condition:
                    df.at[i, 'RSI OBV Strategy'] = 'Buy'
                    trade_status = 'open'
                    trailing_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                elif sell_condition:
                    df.at[i, 'RSI OBV Strategy'] = 'Sell'
                    trade_status = 'none'
                else:
                    df.at[i, 'RSI OBV Strategy'] = 'Hold'

            self.logger.log_or_print("RSI OBV Strategy with Trailing Stop Loss applied successfully.", level="INFO")
        except Exception as e:
            self.logger.log_or_print(f"An error occurred while applying RSI OBV Strategy: {e}", level="ERROR")
            raise  # Optionally re-raise the exception for external handling

        return df
    #4. OBV Strategy
    def apply_obv_with_trailing_stop(self, df, atr_multiplier=1):
        """
        Applies the OBV trading strategy along with a trailing stop loss mechanism,
        based on the On-Balance Volume (OBV) moving average difference indicator,
        enhanced with an ATR-based trailing stop loss.

        Parameters:
        - df (DataFrame): The DataFrame containing the stock data, including 'OBV_SMA_Diff', 'High', 'Low', 'Close'.
        - atr_multiplier (float): The multiplier of ATR to set the trailing stop loss distance.

        Enhancements:
        - The function applies a trailing stop loss for open positions based on the specified ATR multiplier.
        """
        if df is None or df.empty:
            self.logger.log_or_print("DataFrame is empty or None", level="ERROR")
            return

        required_columns = ['OBV_SMA_Diff', 'High', 'Low', 'Close']
        if not all(column in df.columns for column in required_columns):
            self.logger.log_or_print("One or more required columns are missing from DataFrame", level="ERROR")
            return

        try:
            df['ATR'] = ta.atr(df['High'], df['Low'], df['Close'], length=14)  # Calculate ATR for trailing stop loss
            trailing_stop_loss = None
            trade_status = 'none'  # Possible values: 'none', 'open'

            for i, row in df.iterrows():
                current_price = row['Close']
                if trade_status == 'open':
                    current_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                    trailing_stop_loss = max(trailing_stop_loss, current_stop_loss) if current_price > df.loc[i - 1, 'Close'] else trailing_stop_loss
                    if current_price <= trailing_stop_loss:
                        df.at[i, 'OBV Strategy'] = 'Sell'
                        trade_status = 'none'
                        continue

                buy_condition = row['OBV_SMA_Diff'] > 0
                sell_condition = row['OBV_SMA_Diff'] < 0

                if buy_condition:
                    df.at[i, 'OBV Strategy'] = 'Buy'
                    trade_status = 'open'
                    trailing_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                elif sell_condition:
                    df.at[i, 'OBV Strategy'] = 'Sell'
                    trade_status = 'none'
                else:
                    df.at[i, 'OBV Strategy'] = 'Hold'

            self.logger.log_or_print("OBV Strategy with Trailing Stop Loss applied successfully.", level="INFO")
        except Exception as e:
            self.logger.log_or_print(f"An error occurred while applying OBV Strategy: {e}", level="ERROR")
            raise  # Optionally re-raise the exception for external handling

        return df
    
    #5. RSI Strategy
    def apply_rsi_with_trailing_stop(self, df, atr_multiplier=1):
        """
        Applies the RSI trading strategy along with a trailing stop loss mechanism,
        based on the RSI indicator, enhanced with an ATR-based trailing stop loss.

        Parameters:
        - df (DataFrame): The DataFrame containing the stock data, including 'RSI_14', 'High', 'Low', 'Close'.
        - atr_multiplier (float): The multiplier of ATR to set the trailing stop loss distance.

        The function adds a 'RSI Strategy' column to the DataFrame indicating 'Buy', 'Sell', or 'Hold' signals
        and manages a trailing stop loss for open positions.
        """
        if df is None or df.empty:
            self.logger.log_or_print("DataFrame is empty or None", level="ERROR")
            return

        required_columns = ['RSI_14', 'High', 'Low', 'Close']
        if not all(column in df.columns for column in required_columns):
            self.logger.log_or_print("One or more required columns are missing from DataFrame", level="ERROR")
            return

        try:
            df['ATR'] = ta.atr(df['High'], df['Low'], df['Close'], length=14)  # Calculate ATR for trailing stop loss
            trailing_stop_loss = None
            trade_status = 'none'  # Possible values: 'none', 'open'

            for i, row in df.iterrows():
                current_price = row['Close']
                if trade_status == 'open':
                    current_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                    trailing_stop_loss = max(trailing_stop_loss, current_stop_loss) if current_price > df.loc[i - 1, 'Close'] else trailing_stop_loss
                    if current_price <= trailing_stop_loss:
                        df.at[i, 'RSI Strategy'] = 'Sell'
                        trade_status = 'none'
                        continue

                buy_condition = row['RSI_14'] < 35
                sell_condition = row['RSI_14'] > 65

                if buy_condition:
                    df.at[i, 'RSI Strategy'] = 'Buy'
                    trade_status = 'open'
                    trailing_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                elif sell_condition:
                    df.at[i, 'RSI Strategy'] = 'Sell'
                    trade_status = 'none'
                else:
                    df.at[i, 'RSI Strategy'] = 'Hold'

            self.logger.log_or_print("RSI Strategy with Trailing Stop Loss applied successfully.", level="INFO")
        except Exception as e:
            self.logger.log_or_print(f"An error occurred while applying RSI Strategy: {e}", level="ERROR")
            raise  # Optionally re-raise the exception for external handling

        return df
    #6. RSI Strategy2
    def apply_rsi2_with_trailing_stop(self, df, atr_multiplier=1):
        """
        Applies the RSI2 trading strategy along with a trailing stop loss mechanism,
        based on the RSI indicator, enhanced with an ATR-based trailing stop loss.
        This strategy uses modified RSI thresholds to generate buy/sell signals.

        Parameters:
        - df (DataFrame): The DataFrame containing the stock data, with columns for
                          'RSI_14', 'High', 'Low', 'Close'.
        - atr_multiplier (float): The multiplier of ATR to set the trailing stop loss distance.

        Adds a 'RSI Strategy2' column to the DataFrame indicating 'Buy', 'Sell', or 'Hold' signals
        and manages a trailing stop loss for open positions.

        Returns:
        - DataFrame: The modified DataFrame with 'RSI Strategy2' signals applied.
        """
        if df is None or df.empty:
            self.logger.log_or_print("DataFrame is empty or None.", level="ERROR")
            return df

        required_columns = ['RSI_14', 'High', 'Low', 'Close']
        missing_columns = [col for col in required_columns if col not in df.columns]
        if missing_columns:
            self.logger.log_or_print(f"Missing required columns: {missing_columns}", level="ERROR")
            return df

        # Ensure 'ATR' column exists, calculate if necessary
        if 'ATR' not in df.columns:
            df['ATR'] = ta.atr(df['High'], df['Low'], df['Close'], length=14)

        # Initialize variables for tracking trailing stop loss and trade status
        trailing_stop_loss = -float('inf')  # Use negative infinity to ensure any price is higher initially
        trade_status = 'none'  # Possible values: 'none', 'open'

        for i, row in df.iterrows():
            current_price = row['Close']
            current_stop_loss = current_price - (row['ATR'] * atr_multiplier)
            
            # Update trailing stop loss for open trades
            if trade_status == 'open':
                trailing_stop_loss = max(trailing_stop_loss, current_stop_loss)
                if current_price <= trailing_stop_loss:
                    df.at[i, 'RSI Strategy2'] = 'Sell'
                    trade_status = 'none'  # Close trade
                    continue  # Skip further checks

            # Define buy and sell conditions based on RSI values
            buy_condition = row['RSI_14'] < 40
            sell_condition = row['RSI_14'] > 60

            # Apply buy, sell, or hold decisions
            if buy_condition:
                df.at[i, 'RSI Strategy2'] = 'Buy'
                trade_status = 'open'  # Open a new trade
                trailing_stop_loss = current_stop_loss  # Initialize trailing stop loss for the new trade
            elif sell_condition and trade_status != 'open':  # Only sell if no open trade based on trailing stop loss
                df.at[i, 'RSI Strategy2'] = 'Sell'
                trade_status = 'none'
            else:
                df.at[i, 'RSI Strategy2'] = 'Hold'

        self.logger.log_or_print("RSI Strategy2 with Trailing Stop Loss applied successfully.", level="INFO")
        return df
    #7. MACD Strategy
    def apply_macd_with_trailing_stop(self, df, atr_multiplier=1):
        """
        Applies an enhanced MACD trading strategy along with a trailing stop loss mechanism,
        requiring both MACD and MACD signal lines to be below zero for a 'Buy' signal and above zero for a 'Sell' signal,
        enhanced with an ATR-based trailing stop loss.

        Parameters:
        - df (DataFrame): The DataFrame containing the stock data, including 'MACDh_12_26_9', 'MACD_12_26_9', 'MACDs_12_26_9', 'High', 'Low', 'Close'.
        - atr_multiplier (float): The multiplier of ATR to set the trailing stop loss distance.

        The function adds a 'MACD Strategy' column to the DataFrame indicating 'Buy', 'Sell', or 'Hold' signals
        and manages a trailing stop loss for open positions.
        """
        if df is None or df.empty:
            self.logger.log_or_print("DataFrame is empty or None", level="ERROR")
            return

        required_columns = ['MACDh_12_26_9', 'MACD_12_26_9', 'MACDs_12_26_9', 'High', 'Low', 'Close']
        if not all(column in df.columns for column in required_columns):
            self.logger.log_or_print("One or more required columns are missing from DataFrame", level="ERROR")
            return

        try:
            df['ATR'] = ta.atr(df['High'], df['Low'], df['Close'], length=14)  # Calculate ATR for trailing stop loss
            trailing_stop_loss = None
            trade_status = 'none'  # Possible values: 'none', 'open'

            for i, row in df.iterrows():
                current_price = row['Close']
                if trade_status == 'open':
                    current_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                    trailing_stop_loss = max(trailing_stop_loss, current_stop_loss) if current_price > df.loc[i - 1, 'Close'] else trailing_stop_loss
                    if current_price <= trailing_stop_loss:
                        df.at[i, 'MACD Strategy'] = 'Sell'
                        trade_status = 'none'
                        continue

                buy_condition = (row['MACD_12_26_9'] < 0) and (row['MACDs_12_26_9'] < 0) and (row['MACDh_12_26_9'] > 0)
                sell_condition = (row['MACD_12_26_9'] > 0) and (row['MACDs_12_26_9'] > 0) and (row['MACDh_12_26_9'] < 0)

                if buy_condition:
                    df.at[i, 'MACD Strategy'] = 'Buy'
                    trade_status = 'open'
                    trailing_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                elif sell_condition:
                    df.at[i, 'MACD Strategy'] = 'Sell'
                    trade_status = 'none'
                else:
                    df.at[i, 'MACD Strategy'] = 'Hold'

            self.logger.log_or_print("Enhanced MACD Strategy with Trailing Stop Loss applied successfully.", level="INFO")
        except Exception as e:
            self.logger.log_or_print(f"An error occurred while applying Enhanced MACD Strategy: {e}", level="ERROR")
            raise  # Optionally re-raise the exception for external handling

        return df
    #8. CMF Strategy
    def apply_cmf_with_trailing_stop(self, df, atr_multiplier=1):
        """
        Applies the CMF trading strategy along with a trailing stop loss mechanism,
        based on the Chaikin Money Flow (CMF) indicator, enhanced with an ATR-based trailing stop loss.

        Parameters:
        - df (DataFrame): The DataFrame containing the stock data, including 'CMF_20', 'High', 'Low', 'Close'.
        - atr_multiplier (float): The multiplier of ATR to set the trailing stop loss distance.

        The function adds a 'CMF Strategy' column to the DataFrame indicating 'Buy', 'Sell', or 'Hold' signals
        and manages a trailing stop loss for open positions.
        """
        if df is None or df.empty:
            self.logger.log_or_print("DataFrame is empty or None", level="ERROR")
            return

        required_columns = ['CMF_20', 'High', 'Low', 'Close']
        if not all(column in df.columns for column in required_columns):
            self.logger.log_or_print("One or more required columns are missing from DataFrame", level="ERROR")
            return

        try:
            df['ATR'] = ta.atr(df['High'], df['Low'], df['Close'], length=14)  # Calculate ATR for trailing stop loss
            trailing_stop_loss = None
            trade_status = 'none'  # Possible values: 'none', 'open'

            for i, row in df.iterrows():
                current_price = row['Close']
                if trade_status == 'open':
                    current_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                    trailing_stop_loss = max(trailing_stop_loss, current_stop_loss) if current_price > df.loc[i - 1, 'Close'] else trailing_stop_loss
                    if current_price <= trailing_stop_loss:
                        df.at[i, 'CMF Strategy'] = 'Sell'
                        trade_status = 'none'
                        continue

                buy_condition = row['CMF_20'] > 0
                sell_condition = row['CMF_20'] < 0

                if buy_condition:
                    df.at[i, 'CMF Strategy'] = 'Buy'
                    trade_status = 'open'
                    trailing_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                elif sell_condition:
                    df.at[i, 'CMF Strategy'] = 'Sell'
                    trade_status = 'none'
                else:
                    df.at[i, 'CMF Strategy'] = 'Hold'

            self.logger.log_or_print("CMF Strategy with Trailing Stop Loss applied successfully.", level="INFO")
        except Exception as e:
            self.logger.log_or_print(f"An error occurred while applying CMF Strategy: {e}", level="ERROR")
            raise  # Optionally re-raise the exception for external handling

        return df
    #9. EMA5 PSAR Strategy
    def apply_ema5_psar_with_trailing_stop(self, df, atr_multiplier=1):
        """
        Applies the EMA5 PSAR trading strategy along with a trailing stop loss mechanism,
        based on the comparison between Parabolic SAR (PSAR) and Exponential Moving Average (EMA) of 5 periods,
        enhanced with an ATR-based trailing stop loss.

        Parameters:
        - df (DataFrame): The DataFrame containing the stock data, including 'PSARl_0.02_0.2', 'EMA5', 'High', 'Low', 'Close'.
        - atr_multiplier (float): The multiplier of ATR to set the trailing stop loss distance.

        The function adds an 'EMA5 PSAR Strategy' column to the DataFrame indicating 'Buy', 'Sell', or 'Hold' signals
        and manages a trailing stop loss for open positions.
        """
        if df is None or df.empty:
            self.logger.log_or_print("DataFrame is empty or None", level="ERROR")
            return

        required_columns = ['PSARl_0.02_0.2', 'EMA5', 'High', 'Low', 'Close']
        if not all(column in df.columns for column in required_columns):
            self.logger.log_or_print("One or more required columns are missing from DataFrame", level="ERROR")
            return

        try:
            df['ATR'] = ta.atr(df['High'], df['Low'], df['Close'], length=14)  # Calculate ATR for trailing stop loss
            trailing_stop_loss = None
            trade_status = 'none'  # Possible values: 'none', 'open'

            for i, row in df.iterrows():
                current_price = row['Close']
                if trade_status == 'open':
                    current_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                    trailing_stop_loss = max(trailing_stop_loss, current_stop_loss) if current_price > df.loc[i - 1, 'Close'] else trailing_stop_loss
                    if current_price <= trailing_stop_loss:
                        df.at[i, 'EMA5 PSAR Strategy'] = 'Sell'
                        trade_status = 'none'
                        continue

                buy_condition = row['PSARl_0.02_0.2'] < row['EMA5']
                sell_condition = row['PSARl_0.02_0.2'] > row['EMA5']

                if buy_condition:
                    df.at[i, 'EMA5 PSAR Strategy'] = 'Buy'
                    trade_status = 'open'
                    trailing_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                elif sell_condition:
                    df.at[i, 'EMA5 PSAR Strategy'] = 'Sell'
                    trade_status = 'none'
                else:  # Default to 'Sell' if conditions are not met for 'Buy'
                    df.at[i, 'EMA5 PSAR Strategy'] = 'Sell'

            self.logger.log_or_print("EMA5 PSAR Strategy with Trailing Stop Loss applied successfully.", level="INFO")
        except Exception as e:
            self.logger.log_or_print(f"An error occurred while applying EMA5 PSAR Strategy: {e}", level="ERROR")
            raise  # Optionally re-raise the exception for external handling

        return df
    #10. RSI14_OBV_RoC Strategy
    def apply_rsi14_obv_roc_with_trailing_stop(self, df, atr_multiplier=1):
        """
        Applies the RSI14 OBV Rate of Change (RoC) trading strategy along with a trailing stop loss mechanism,
        based on the RSI and the Rate of Change of the On-Balance Volume (OBV), enhanced with an ATR-based trailing stop loss.

        Parameters:
        - df (DataFrame): The DataFrame containing the stock data, including 'RSI_14', 'OBV_RoC', 'High', 'Low', 'Close'.
        - atr_multiplier (float): The multiplier of ATR to set the trailing stop loss distance.

        The function adds a 'RSI14_OBV_RoC Strategy' column to the DataFrame indicating 'Buy', 'Sell', or 'Hold' signals
        and manages a trailing stop loss for open positions.
        """
        if df is None or df.empty:
            self.logger.log_or_print("DataFrame is empty or None", level="ERROR")
            return

        required_columns = ['RSI_14', 'OBV_RoC', 'High', 'Low', 'Close']
        if not all(column in df.columns for column in required_columns):
            self.logger.log_or_print("One or more required columns are missing from DataFrame", level="ERROR")
            return

        try:
            df['ATR'] = ta.atr(df['High'], df['Low'], df['Close'], length=14)  # Calculate ATR for trailing stop loss
            trailing_stop_loss = None
            trade_status = 'none'  # Possible values: 'none', 'open'

            for i, row in df.iterrows():
                current_price = row['Close']
                if trade_status == 'open':
                    current_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                    trailing_stop_loss = max(trailing_stop_loss, current_stop_loss) if current_price > df.loc[i - 1, 'Close'] else trailing_stop_loss
                    if current_price <= trailing_stop_loss:
                        df.at[i, 'RSI14_OBV_RoC Strategy'] = 'Sell'
                        trade_status = 'none'
                        continue

                buy_condition = row['RSI_14'] < 35 and row['OBV_RoC'] > 0
                sell_condition = row['RSI_14'] > 65 and row['OBV_RoC'] < 0

                if buy_condition:
                    df.at[i, 'RSI14_OBV_RoC Strategy'] = 'Buy'
                    trade_status = 'open'
                    trailing_stop_loss = current_price - (row['ATR'] * atr_multiplier)
                elif sell_condition:
                    df.at[i, 'RSI14_OBV_RoC Strategy'] = 'Sell'
                    trade_status = 'none'
                else:
                    df.at[i, 'RSI14_OBV_RoC Strategy'] = 'Hold'

            self.logger.log_or_print("RSI14_OBV_RoC Strategy with Trailing Stop Loss applied successfully.", level="INFO")
        except Exception as e:
            self.logger.log_or_print(f"An error occurred while applying RSI14_OBV_RoC Strategy: {e}", level="ERROR")
            raise  # Optionally re-raise the exception for external handling

        return df