import logging
import os
from dateutil import tz
import time
import datetime
import pytz
from datetime import datetime
import pandas as pd
import pandas_ta as ta
from bs4 import BeautifulSoup
from PyQt5.QtCore import QObject, pyqtSignal
from selenium.webdriver.edge.options import Options as EdgeOptions
from selenium.common.exceptions import TimeoutException, UnexpectedAlertPresentException, NoSuchElementException
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
from LoggerFunction import Logger  # Import your Logger class
from pandas import DataFrame
from selenium.common.exceptions import WebDriverException

class DataFetcher(QObject):

    def __init__(self, driver_path):
        super().__init__()
        self.driver_path = driver_path

        self.logger = Logger()
        
    def fetch_data(self, ticker):
        # Define the GMT+3 timezone
        gmt3 = tz.tzoffset('GMT+3', 3*3600)
        df = self.initialize_dataframe()
        # Get the current date and time in GMT+3
        current_datetime = datetime.now(gmt3)
        current_date = current_datetime.date()

        driver = None  # Initialize driver to None
        filename = f"raw_{ticker}.csv"

        try:
            if os.path.exists(filename):
                self.logger.log_or_print(f"File '{filename}' already exists.", level="INFO")

                df_existing =  pd.read_csv(filename)
                self.logger.log_or_print(f"Data for ticker {ticker} has {len(df_existing)} rows before sorting.", level="INFO")
                # Ensure the 'Date' column is of datetime type
                df_existing['Date'] = pd.to_datetime(df_existing['Date']).apply(lambda x: x.date())
                MAX_ROWS = 1000  # Define the maximum number of rows to fetch
                # Get the maximum date
                # Convert to datetime.date
                max_date = df_existing['Date'].max()
                # Calculate the difference in days
                difference_in_days = (current_date - max_date).days
                URL = f'{BASE_URL}?currLanguage=en&companyCode={ticker}&activeTab=0'

                # Initialize Edge driver
                try:
                    driver_service = Service(EDGE_DRIVER_PATH)
                    driver = EdgeDriver(service=driver_service)
                except Exception as e:
                    self.logger.log_or_print(
                        f"Failed to initialize Edge WebDriver for ticker {ticker}: {str(e)}", level="ERROR", exc_info=True)
                    raise RuntimeError(f"WebDriver initialization failed for {ticker}") from e

                try:
                    driver.get(URL)
                except TimeoutException as e:
                    self.logger.log_or_print(
                        f"Page load timed out for URL {URL}: {str(e)}", level="ERROR", exc_info=True)
                    raise
                except WebDriverException as e:
                    self.logger.log_or_print(
                        f"WebDriver error occurred navigating to {URL}: {str(e)}", level="ERROR", exc_info=True)
                    raise

                self.dismiss_alert_if_present(driver)

                # Adjust the value of the input field
                # Ensure the input field is present before adjusting its value
                from_date_input_selector = "#fromDate"
                WebDriverWait(driver, WEBDRIVER_WAIT_TIME).until(
                    EC.presence_of_element_located((By.CSS_SELECTOR, from_date_input_selector))
                )
                driver.execute_script(
                    f'document.querySelector("{from_date_input_selector}").value = "1/1/2010";'
                )
                # Verify the value was set correctly
                assert "1/1/2010" == driver.execute_script(
                    f'return document.querySelector("{from_date_input_selector}").value;'
                ), "Failed to set the fromDate input value."

                # Find the button and click it
                update_button = driver.execute_script(
                    'return document.querySelector("#command > div.filterbox > div.button-all")')
                update_button.click()

                # Wait for a couple of seconds after pressing the button
                #time.sleep(2)

                # Wait for table to load
                self.wait_for_table_to_load(driver)
                #loop to check the df of each page
                page_num = 1
                #loop to get the data frame to temp data frame
                #flag to check if next page is required
                next_page = True
                while next_page and len(df) < MAX_ROWS:
                    self.logger.log_or_print(f"Fetching data from page {page_num} for ticker {ticker}.", level="INFO")
                    # Extract data into a temporary DataFrame
                    df_temp = self.extract_data_from_page(df, driver)
                    #get max date form exisiting data frame
                    max_date = df['Date'].max()
                    self.logger.log_or_print(f"Max date from existing data frame: {max_date}", level="INFO")
                    #get max date from df_temp
                    temp_max_date = df_temp['Date'].max()
                    self.logger.log_or_print(f"Temp_Max date from existing data frame: {temp_max_date}", level="INFO")
                    #get min date from df_temp
                    temp_min_date = df_temp['Date'].min()
                    self.logger.log_or_print(f"Temp_Min date from existing data frame: {temp_min_date}", level="INFO")
                    #append the temp data frame to the existing data frame
                    self.logger.log_or_print(f"Appending temp data frame to the existing data frame", level="INFO") 
                    df = pd.concat([df_existing, df_temp], ignore_index=True)
                    # Remove duplicates based on the 'Date' column
                    self.logger.log_or_print(f"Removing duplicates based on the 'Date' column", level="INFO")
                    df = df.drop_duplicates(subset='Date', keep='first')
                    #sort the data frame
                    self.logger.log_or_print(f"Sorting the data frame", level="INFO")
                    df = df.sort_values(by='Date', ascending=True)
                    self.logger.log_or_print(f"Data for ticker {ticker} has {len(df)} rows after sorting.", level="INFO")
                    #check if existing data frame max date withing the min and max date for the temp data frame
                    self.logger.log_or_print(f"Checking if max date from existing data frame {max_date} is within the min and max date for the temp data frame", level="INFO")
                    if max_date >= temp_min_date and max_date <= temp_max_date:
                        self.logger.log_or_print(f"Max date from existing data frame {max_date} is within the min and max date for the temp data frame", level="INFO")
                        #exit the while loop
                        next_page = False
                    #check if we fetched 1000 lines, if yes exit the loop
                    elif len(df) >= MAX_ROWS:
                        self.logger.log_or_print(f"Data for ticker {ticker} has {len(df)} rows after sorting.", level="INFO")
                        self.logger.log_or_print(f"Data for ticker {ticker} has 1000 rows, exiting the loop", level="INFO")
                        break
                    else:
                        self.logger.log_or_print(f"Max date from existing data frame {max_date} is not within the min and max date for the temp data frame", level="INFO")

            else:
                self.logger.log_or_print(f"File '{filename}' does not exist.", level="INFO")
                # Initialization and existing code

                URL = f'{BASE_URL}?currLanguage=en&companyCode={ticker}&activeTab=0'

                # Initialize Edge driver
                try:
                    driver_service = Service(EDGE_DRIVER_PATH)
                    driver = EdgeDriver(service=driver_service)
                except Exception as e:
                    self.logger.log_or_print(
                        f"Failed to initialize Edge WebDriver for ticker {ticker}: {str(e)}", level="ERROR", exc_info=True)
                    raise RuntimeError(f"WebDriver initialization failed for {ticker}") from e

                self.logger.log_or_print(f"Fetching data from URL {URL} for ticker {ticker}.", level="INFO")
                try:
                    driver.get(URL)
                except TimeoutException as e:
                    self.logger.log_or_print(
                        f"Page load timed out for URL {URL}: {str(e)}", level="ERROR", exc_info=True)
                    raise
                except WebDriverException as e:
                    self.logger.log_or_print(
                        f"WebDriver error occurred navigating to {URL}: {str(e)}", level="ERROR", exc_info=True)
                    raise

                self.dismiss_alert_if_present(driver)

                # Adjust the value of the input field
                # Ensure the input field is present before adjusting its value
                from_date_input_selector = "#fromDate"
                WebDriverWait(driver, WEBDRIVER_WAIT_TIME).until(
                    EC.presence_of_element_located((By.CSS_SELECTOR, from_date_input_selector))
                )
                driver.execute_script(
                    f'document.querySelector("{from_date_input_selector}").value = "1/1/2010";'
                )
                # Verify the value was set correctly
                assert "1/1/2010" == driver.execute_script(
                    f'return document.querySelector("{from_date_input_selector}").value;'
                ), "Failed to set the fromDate input value."

                # Find the button and click it
                update_button = driver.execute_script(
                    'return document.querySelector("#command > div.filterbox > div.button-all")')
                update_button.click()

                # Wait for a couple of seconds after pressing the button
                #time.sleep(2)

                # Wait for table to load
                self.wait_for_table_to_load(driver)

                page_num = 1
                
                while self.can_navigate_to_next_page(driver):
                    self.logger.log_or_print(f"Fetching data from page {page_num} for ticker {ticker}.", level="INFO")
                    # Extract data into a temporary DataFrame
                    df_temp = self.extract_data_from_page(df, driver)

                    # Append the temporary DataFrame to the main DataFrame
                    df = pd.concat([df, df_temp], ignore_index=True)

                    # Remove duplicates based on the 'Date' column
                    df = df.drop_duplicates(subset='Date', keep='first')

                    self.navigate_to_next_page(driver)
                    page_num += 1
            self.logger.log_or_print(f"Fetched a total of {len(df)} rows from the site for ticker {ticker}.", level="INFO")
            df = df.sort_values(by='Date', ascending=True)
            # This will keep only the latest 'desired_rows'
            # df = df.tail(desired_rows)
            self.logger.log_or_print(f"Data for ticker {ticker} has {len(df)} rows after sorting.", level="INFO")
            # Compute the actual change and change% based on the Close prices
            df['Change'] = df['Close'].diff()
            df['Change%'] = df['Change'] / df['Close'].shift(1) * 100
            # Round the values to two decimal places
            df['Change'] = df['Change'].round(2)
            df['Change%'] = df['Change%'].round(2)

            filename = f"raw_{ticker}.csv"
            self.logger.log_or_print(f"Saving {len(df)} rows to CSV file '{filename}' for ticker {ticker}.", level="INFO")
            df.to_csv(filename, index=False)
            self.logger.log_or_print(f"Data successfully saved to '{filename}' for ticker {ticker}.", level="INFO")
            return df

        except (TimeoutException, UnexpectedAlertPresentException, NoSuchElementException) as e:
            self.logger.log_or_print(
                f"Specific web scraping error occurred for ticker {ticker}: {str(e)}", level="ERROR", exc_info=True)
        except Exception as e:
            self.logger.log_or_print(
                f"Unexpected error occurred for ticker {ticker}: {str(e)}", level="ERROR", exc_info=True)
            return None

        finally:
            if driver:
                self.release_webdriver_resource(driver)

    def initialize_dataframe(self):
        # Initialize DataFrame
        df = pd.DataFrame(columns=[
            "Date", "Close", "Open", "High", "Low", "Change", "Change%", "T.Shares", "Volume", "No. Trades"
        ])
        return df

    def wait_for_table_to_load(self, driver):
        start_time = time.time()  # Start time measurement
        try:
            WebDriverWait(driver, WEBDRIVER_WAIT_TIME).until(
                EC.presence_of_element_located((By.CSS_SELECTOR, "#dispTable")))
        except TimeoutException:
            self.logger.log_or_print("Table did not load in time.", level="ERROR")
            raise
        finally:
            end_time = time.time()  # End time measurement
            self.logger.log_or_print(f"wait_for_table_to_load executed in {end_time - start_time:.2f} seconds.", level="INFO")

    def extract_data_from_page(self, df, driver):
        """
        Extracts data from the current page and appends it to the DataFrame.
        
        This method iterates over each row in the table found on the page. For each row,
        it attempts to extract data for various fields. If a field in the `open` column contains 
        a dash ('-') or zero ('0'), indicating missing or invalid data, the row is skipped, 
        and `self.desired_rows` is decremented to adjust the number of rows that are ultimately desired.
        
        Parameters:
            df (pandas.DataFrame): The DataFrame to which extracted data will be appended.
            driver (selenium.webdriver): The WebDriver instance used for web page interaction.
        
        Returns:
            pandas.DataFrame: The DataFrame with the newly appended data.
        
        Note:
            This method directly modifies `self.desired_rows` if rows are skipped due to missing or invalid data.
        """
        try:
            table_html = driver.execute_script('return document.querySelector("#dispTable").outerHTML;')
            soup = BeautifulSoup(table_html, 'html.parser')
            table = soup.find('table')

            for row in table.find_all('tr')[1:]:  # Iterate over table rows, skipping the header row
                cols = row.find_all('td')
                # Adjust the index for the 'open' column as necessary
                open_col_text = cols[8].text.strip()  # Assuming 'open_price' is in cols[8] based on your parse_row_data definition
                if open_col_text == '-' or open_col_text == '0':
                    self.logger.log_or_print("Skipping row due to missing or invalid 'open' data", level="INFO")
                    #self.desired_rows -= 1
                    continue  # Skip this row

                try:
                    row_data = self.parse_row_data(cols)
                    df.loc[len(df)] = row_data
                except ValueError as e:
                    self.logger.log_or_print(f"Skipping row due to conversion error: {str(e)}", level="WARNING")
                    #self.desired_rows -= 1

            return df
        except AttributeError as e:
            self.logger.log_or_print(f"Attribute error, likely due to changes in the page structure: {str(e)}", level="ERROR")
            return df
        except Exception as e:
            self.logger.log_or_print(f"Unexpected error during data extraction: {str(e)}", level="ERROR")
            return pd.DataFrame()  # Return an empty DataFrame to indicate failure

    def parse_row_data(self, cols):
        """
        Parses data from table columns into a format suitable for appending to the DataFrame.
        
        Parameters:
            cols (list): A list of BeautifulSoup elements representing table columns.
        
        Returns:
            list: A list containing parsed data from the table row.
        """
        date = datetime.strptime(cols[9].text.strip(), '%d/%m/%Y').date()
        open_price = self.parse_float(cols[8].text.strip())
        high = self.parse_float(cols[7].text.strip())
        low = self.parse_float(cols[6].text.strip())
        close = self.parse_float(cols[5].text.strip())
        change = self.parse_float(cols[4].text.strip())
        change_percent = self.parse_float(cols[3].text.strip().replace('%', ''))
        t_shares = self.parse_int(cols[2].text.strip())
        volume = self.parse_int(cols[1].text.strip())
        no_trades = self.parse_int(cols[0].text.strip())

        # The DataFrame initialized in ``initialize_dataframe`` expects the
        # following column order:
        # Date, Close, Open, High, Low, Change, Change%, T.Shares, Volume, No. Trades
        # Ensure the returned values match this order to avoid column
        # misalignment when ``row_data`` is appended to the DataFrame.
        return [
            date,
            close,
            open_price,
            high,
            low,
            change,
            change_percent,
            t_shares,
            volume,
            no_trades,
        ]

    def parse_float(self, value):
        """
        Converts a string to a float, replacing non-numeric values with 0.0.
        
        Parameters:
            value (str): The string value to convert.
        
        Returns:
            float: The converted float value, or 0.0 if conversion is not possible.
        """
        try:
            return 0.0 if value == '-' else float(value.replace(',', ''))
        except ValueError:
            return 0.0

    def parse_int(self, value):
        """
        Converts a string to an integer, replacing non-numeric values with 0.
        
        Parameters:
            value (str): The string value to convert.
        
        Returns:
            int: The converted integer value, or 0 if conversion is not possible.
        """
        try:
            return 0 if value == '-' else int(value.replace(',', ''))
        except ValueError:
            return 0

    def navigate_to_next_page(self, driver):
        """Navigate to the next page of data."""
        try:

            next_page_btn_selector = "#ajxDspId > div > span.pagelinks > a:nth-child(11)"
            next_page_btn = driver.find_element(
                By.CSS_SELECTOR, next_page_btn_selector)

            if next_page_btn:
                next_page_btn.click()
                WebDriverWait(driver, WEBDRIVER_WAIT_TIME).until(
                    EC.presence_of_element_located((By.CSS_SELECTOR, "#dispTable")))

            else:
                self.logger.log_or_print(
                    "Next page button not found.", level="WARNING")

        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred while navigating to the next page: {str(e)}", level="ERROR", exc_info=True)

    def release_webdriver_resource(self, driver):

        driver.quit()

    def dismiss_alert_if_present(self, driver):
        try:
            alert = driver.switch_to.alert
            alert.dismiss()

        except NoAlertPresentException:
            self.logger.log_or_print("No alert was present.", level="INFO")
                
    def can_navigate_to_next_page(self, driver):
        """
        Checks if the next page button exists and is clickable.
        
        Args:
            driver: Selenium WebDriver instance.
        
        Returns:
            bool: True if the next page button exists and is clickable, False otherwise.
        """
        try:
            next_page_btn_selector = "#ajxDspId > div > span.pagelinks > a:nth-child(11)"
            WebDriverWait(driver, WEBDRIVER_WAIT_TIME).until(EC.element_to_be_clickable((By.CSS_SELECTOR, next_page_btn_selector)))
            return True
        except TimeoutException:
            self.logger.log_or_print("Next page button not found or not clickable.", level="WARNING")
            return False




