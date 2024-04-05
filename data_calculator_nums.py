import os

from datetime import datetime
import pandas as pd
import pandas_ta as ta
import numpy as np
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


class DataCalculatorNums(QObject):

    def __init__(self):
        super().__init__()
        self.logger = Logger()
    def calculate_all(self, ticker):
        """
        Calculates financial indicators for a given ticker and saves them to an Excel file.
        If the indicators file already exists, it checks if the data is up-to-date.
        
        Parameters:
        - ticker: The stock ticker symbol for which to calculate indicators.
        """
        # Check if the raw data CSV file exists
        raw_file_path = f"raw_{ticker}.csv"
        if not os.path.exists(raw_file_path):
            self.logger.log_or_print(f"File {raw_file_path} does not exist.", level="INFO", module="Data_Calculator")
            return

        # Read the raw data from CSV file
        df = pd.read_csv(raw_file_path)
        if df.empty:
            self.logger.log_or_print("The DataFrame from raw data is empty.", level="ERROR", module="Data_Calculator")
            return

        # Define the path for the Excel file with indicators
        indicators_file_path = f"indicators2_{ticker}.csv"
        
        # Check if the indicators Excel file already exists
        if os.path.exists(indicators_file_path):
            self.logger.log_or_print(f"File {indicators_file_path} already exists.", level="INFO", module="Data_Calculator")
            
            # Read the indicators from the Excel file
            excel_df = pd.read_csv(indicators_file_path)
            if excel_df.empty:
                self.logger.log_or_print("The Excel DataFrame is empty.", level="ERROR", module="Data_Calculator")
                return
            
            # Compare the last date in the Excel file with the last date in the new data
            last_excel_date = excel_df['Date'].iloc[-1]
            last_df_date = df['Date'].iloc[-1]
            if last_df_date == last_excel_date:
                self.logger.log_or_print("The data is up to date.", level="INFO", module="Data_Calculator")
                return

        # If the Excel file does not exist or the data is not up-to-date, proceed with calculations
        try:
            # Calculate all the indicators
            #calculate sma
            self.calculate_sma(df)
            #calculate rsi
            self.calculate_rsi(df)
            #calculate stochastic oscillator
            self.calculate_stochastic_oscillator(df)
            #calculate cmf
            self.calculate_cmf(df)
            #calculate macd
            self.calculate_macd(df)
            #calculate obv
            self.calculate_obv(df)
            #calcualte ema5
            self.calculate_ema5(df)
            #calculate ema10
            self.calculate_ema10(df)
            #calculate ema20
            self.calculate_ema20(df)
            #calculate ema50
            self.calculate_ema50(df)
            #calculate ema200
            self.calculate_ema200(df)
           

            #calculate psar 
            self.logger.log_or_print("Calculating PSAR", level="INFO", module="Data_Calculator")
            self.calculate_psar(df)
            #calculate atr
            self.calculate_atr(df)
            #calculate rolling std
            self.calculate_rolling_std(df)
            #Calculation completed 
            self.logger.log_or_print("Data calculation completed.", level="INFO", module="Data_Calculator")
            # After calculations, save the updated DataFrame to a csv file
            df.to_csv(indicators_file_path, index=False)
            #df.to_excel(indicators_file_path, engine=EXCEL_ENGINE, index=False)
            self.logger.log_or_print(f"Data calculation completed and saved to {indicators_file_path}.", level="INFO", module="Data_Calculator")
        except Exception as e:
            # Log any exceptions that occur during the calculations or saving process
            self.logger.log_or_print(f"An error occurred during data calculation or saving: {str(e)}", level="ERROR", exc_info=True)



    def calculate_sma(self, df):
        try:
            if df is None:
                self.logger.log_or_print(
                    "data calculatro: recieved dataframes SMA calculation is None.", level="ERROR", module="MainLogic")
            #Check if df have more than 10 record calucate SMA10
            
            df['SMA10'] = ta.sma(df['Close'], length=10).round(2)
            df['SMA50'] = ta.sma(df['Close'], length=50).round(2)
            # Attempt to calculate SMA200; will result in NaNs if df has less than 200 records
            if len(df) >= 200:
                df['SMA200'] = ta.sma(df['Close'], length=200).round(2)
            else:
                df['SMA200'] = float('nan')

           

            if df is None:
                self.logger.log_or_print(
                    "data calculatro: Returned DataFrame from SMA calculation is None.", level="ERROR", module="MainLogic")
        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred in calculate_sma: {str(e)}", level="ERROR", exc_info=True)
            # Return the original DataFrame if an error occurs

    def calculate_rsi(self, df):
        # Preliminary checks
        if df is None:

            return
        for period in [9, 14, 25]:
            try:

                # Calculate RSI
                column_name = f"RSI_{period}"
                df[column_name] = ta.rsi(df["Close"], length=period).round(2)



               
            except Exception as e:
                self.logger.log_or_print(
                    f"An error occurred in calculate_rsi: {str(e)}", level="ERROR", exc_info=True)

    def calculate_stochastic_oscillator(self, df, k_period=9, d_period=6):
        try:
            if df is None:
                self.logger.log_or_print(
                    "DataFrame is None in calculate_stochastic_oscillator", level="ERROR")
                return

            # Log initial DataFrame headers

            stoch_df = ta.stoch(df['High'], df['Low'],
                                df['Close'], k=k_period, d=d_period).round(2)
            for col in stoch_df.columns:
                df[col] = stoch_df[col]

            # Column identifiers based on periods
            stoch_id = f"STOCH_{k_period}_{d_period}_3"



            
        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred in calculate_stochastic_oscillator: {str(e)}", level="ERROR", exc_info=True)

    def calculate_cmf(self, df, window=20):
        try:
            if df is None:

                return

            # Ensure data is sorted by date in ascending order
            # df = df.sort_values(by='Date')
            delta = df['High'] - df['Low']
            zero_delta_indices = delta[delta == 0].index
            if len(zero_delta_indices) > 0:
                self.logger.log_or_print(
                    f"Identified rows with zero high-low difference at indices: {zero_delta_indices.tolist()}", level="WARNING")
            # replace 0 with a small number to avoid division by zero
            delta.replace({0: 0.0001}, inplace=True)
            # Money Flow Multiplier (MFM)

            MFM = ((df['Close'] - df['Low']) -
                   (df['High'] - df['Close'])) / delta
            # Money Flow Volume (MFV)
            MFV = MFM * df['T.Shares']
            # CMF
            df['CMF_' + str(window)] = ((MFV.rolling(window=window).sum() /
                                         df['T.Shares'].rolling(window=window).sum())).round(2)

            # Calculate CMF Rate of Change
            df['CMF_RoC'] = df['CMF_' + str(window)].pct_change() * 100

            # Replace NaN values with 0 or another specified value
            df['CMF_RoC'] = df['CMF_RoC'].fillna(0)

            # Replace inf and -inf values with a defined maximum or minimum value
            max_value = 999  # Example max value for positive infinity
            min_value = -999  # Example min value for negative infinity
            df['CMF_RoC'] = df['CMF_RoC'].replace([np.inf, -np.inf], [max_value, min_value])

        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred in calculate_cmf: {str(e)}", level="ERROR", exc_info=True)

    def calculate_macd(self, df, short_period=12, long_period=26, signal_period=9):
        try:
            if df is None:

                return

            # Compute MACD using pandas-ta
            macd_df = ta.macd(df['Close'], fast=short_period,
                              slow=long_period, signal=signal_period)

            # Extracting MACD, Signal Line, and Histogram from the computed DataFrame
            df['MACD_12_26_9'] = macd_df[f'MACD_{short_period}_{long_period}_{signal_period}'].round(
                2)
            df['MACDs_12_26_9'] = macd_df[f'MACDs_{short_period}_{long_period}_{signal_period}'].round(
                2)
            df['MACDh_12_26_9'] = macd_df[f'MACDh_{short_period}_{long_period}_{signal_period}'].round(
                2)
            # Calculate MACDh rate of change
            df['MACDh_RoC'] = df['MACDh_12_26_9'].pct_change() * 100

            # Replace NaN values with 0 (or another placeholder value of your choice)
            df['MACDh_RoC'] = df['MACDh_RoC'].fillna(0)

            # Replace inf and -inf values with a defined maximum or minimum value
            # This step is optional and depends on how you want to handle these cases
            max_value = 999  # Example max value for inf
            min_value = -999  # Example min value for -inf
            df['MACDh_RoC'] = df['MACDh_RoC'].replace([np.inf, -np.inf], [max_value, min_value])

        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred in calculate_macd: {str(e)}", level="ERROR", exc_info=True)
            # Assuming you have a signal for MACD like for stochastic

    def calculate_obv(self, df):
        try:
            if df is None:

                return

            # Rename column for OBV calculation
            df_temp = df.rename(columns={'T.Shares': 'Volume'})

            # Calculate OBV using pandas-ta with renamed DataFrame
            df['OBV'] = df_temp.ta.obv()

            # OBV Trend Analysis
            df['OBV_SMA'] = df['OBV'].rolling(window=20).mean()
            # Calculate OBV SMA Difference
            df['OBV_SMA_Diff'] = df['OBV'] - df['OBV_SMA']
            
            # Calculate OBV rate of change
            df['OBV_RoC'] = df['OBV'].pct_change() * 100

            # Replace NaN values with 0 (or another placeholder value of your choice)
            df['OBV_RoC'] = df['OBV_RoC'].fillna(0)

            # Replace inf and -inf values with a defined maximum or minimum value
            max_value = 999  # Example max value for inf
            min_value = -999  # Example min value for -inf
            df['OBV_RoC'] = df['OBV_RoC'].replace([np.inf, -np.inf], [max_value, min_value])

            


        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred in calculate_obv: {str(e)}", level="ERROR")

    #calculate ema5
    def calculate_ema5(self, df):
        try:
            if df is None:
                return
            df['EMA5'] = df['Close'].ewm(span=5, adjust=False).mean()
        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred in calculate_ema5: {str(e)}", level="ERROR")
    #calculate ema10
    def calculate_ema10(self, df):
        try:
            if df is None:
                return
            df['EMA10'] = df['Close'].ewm(span=10, adjust=False).mean()
        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred in calculate_ema10: {str(e)}", level="ERROR")
    #calculate ema20
    def calculate_ema20(self, df):
        try:
            if df is None:
                return
            df['EMA20'] = df['Close'].ewm(span=20, adjust=False).mean()
        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred in calculate_ema20: {str(e)}", level="ERROR")
    #calculate ema50
    def calculate_ema50(self, df):
        try:
            if df is None:
                return
            #check if there is 50 records else fill the column with NaN
            if len(df) < 50:
                df['EMA50'] = np.nan
                return
            else:
                df['EMA50'] = df['Close'].ewm(span=50, adjust=False).mean()
        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred in calculate_ema50: {str(e)}", level="ERROR")
    #calculate ema200
    def calculate_ema200(self, df):
        try:
            if df is None:
                return
            #check if there is 200 records else fill the column with NaN
            if len(df) < 200:
                df['EMA200'] = np.nan
                return
            else:
                df['EMA200'] = df['Close'].ewm(span=200, adjust=False).mean()
        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred in calculate_ema200: {str(e)}", level="ERROR")
    #calculate psar
    """
    def calculate_psar(self, df):
        try:
            if df is None:
                return
            df['PSAR'] = ta.psar(df['High'], df['Low'], df['Close'], acceleration=0.02, maximum=0.2)
        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred in calculate_psar: {str(e)}", level="ERROR")
    """
    def calculate_psar(self, df):
        """
        Calculate the Parabolic Stop and Reverse (PSAR) and add it to the DataFrame.

        The function calculates PSAR using the 'ta' library and adds the resulting columns
        directly to the input DataFrame. It handles potential errors and logs useful information
        and errors.

        Parameters:
        - df: A pandas DataFrame containing 'High', 'Low', and 'Close' columns.

        Returns:
        - None. The function modifies the DataFrame in place.
        """
        try:
            if df is None or df.empty:
                self.logger.log_or_print("DataFrame is empty or None.", level="ERROR")
                return
            
            # Ensure required columns are present
            if not all(column in df.columns for column in ['High', 'Low', 'Close']):
                self.logger.log_or_print("Required columns are not present in the DataFrame.", level="ERROR")
                return

            # Calculate PSAR
            psar_result = ta.psar(df['High'], df['Low'], df['Close'], acceleration=0.02, maximum=0.2)
            
            # Check if psar_result is not empty and has columns
            if psar_result.empty or psar_result.columns.size == 0:
                self.logger.log_or_print("PSAR calculation failed. The result is empty or has no columns.", level="ERROR")
                return
            
            # Log the columns being added for clarity
            self.logger.log_or_print(f"Adding the following PSAR columns to the DataFrame: {psar_result.columns.tolist()}", level="INFO")

            # Add each PSAR column to the original DataFrame with its default name
            for column in psar_result.columns:
                df[column] = psar_result[column]

            self.logger.log_or_print("PSAR calculation and column addition completed successfully.", level="INFO")

        except Exception as e:
            self.logger.log_or_print(f"An error occurred in calculate_psar: {str(e)}", level="ERROR", exc_info=True)
    def calculate_atr(self, df):
        """
        Calculates the Average True Range (ATR) for the given DataFrame and handles missing data.

        Parameters:
        - df (DataFrame): The input DataFrame containing 'High', 'Low', and 'Close' columns.

        Modifies the input DataFrame by adding an 'ATR' column representing the Average True Range
        calculated over a default period of 14. Handles missing values to ensure no 'None' values
        in the 'ATR' column.
        """
        try:
            if df is None or df.empty:
                raise ValueError("Input DataFrame cannot be None or empty.")

            # Ensure there are no missing values in 'High', 'Low', and 'Close' columns
            df.fillna(method='ffill', inplace=True)  # Forward fill missing values
            df.fillna(method='bfill', inplace=True)  # Backward fill to cover leading missing values

            # Calculate ATR and add it to the DataFrame
            df['ATR'] = ta.atr(df['High'], df['Low'], df['Close'], length=14)

            # After calculating ATR, ensure there are no NaN values. If there are, fill them.
            df['ATR'].fillna(method='ffill', inplace=True)
            df['ATR'].fillna(method='bfill', inplace=True)  # For the initial rows

            # Alternatively, set ATR to 0 or a small number where NaNs remain (e.g., all rows are NaN)
            df['ATR'].fillna(0, inplace=True)

        except Exception as e:
            self.logger.log_or_print(f"An error occurred while calculating ATR: {e}", level="ERROR")
            raise  # Optionally, re-raise the exception for external handling.
    #calculate rolling std
    def calculate_rolling_std(self, df):
        try:
            if df is None:
                return
            df['Rolling_Std_10'] = df['Close'].rolling(window=10).std()
            df['Rolling_Std_50'] = df['Close'].rolling(window=50).std()
        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred in calculate_rolling_std: {str(e)}", level="ERROR")
