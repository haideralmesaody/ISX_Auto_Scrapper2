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


class DataFetcher(QObject):

    def __init__(self, driver_path):
        super().__init__()
        self.driver_path = driver_path
        self.logger = Logger()

    def fetch_data(self, ticker, desired_rows):
        # Define the GMT+3 timezone
        gmt3 = tz.tzoffset('GMT+3', 3*3600)
        df = self.initialize_dataframe()
        # Get the current date and time in GMT+3
        current_datetime = datetime.now(gmt3)
        current_date = current_datetime.date()

        driver = None  # Initialize driver to None
        filename = f"raw_{ticker}.csv"

        try:
            df_existing, file_outdated = self.handle_existing_file(filename, desired_rows)
            if file_outdated:
                # If the existing file is outdated or has fewer rows, fetch new data
                driver = self.initialize_edge_driver(ticker)
            elif df_existing is not None:
                # If the existing file is valid and not outdated, use its data
                # ... [code to handle existing data] ...
            else:
                # If no existing file, fetch new data
                driver = self.initialize_edge_driver(ticker)
                # ... [code to fetch new data] ...
            if os.path.exists(filename):

                df_existing = self.read_data_from_file(filename)
                number_of_rows = df_existing.shape[0]

                # Ensure the 'Date' column is of datetime type
                df_existing['Date'] = pd.to_datetime(df_existing['Date'])
                # Get the maximum date
                # Convert to datetime.date
                max_date = df_existing['Date'].max().date()
                # Calculate the difference in days
                difference_in_days = (current_date - max_date).days

                # if the CSV is older than 10day the CSV or the number of rows are less than the desired rows be deleted and the to proceed normally to fetch the desired trading days
                if difference_in_days > 20 or number_of_rows < desired_rows-20:

                    os.remove(filename)
                    # Initialization and existing code

                    URL = f'{BASE_URL}?currLanguage=en&companyCode={ticker}&activeTab=0'

                    # Initialize Edge driver
                    driver_service = Service(EDGE_DRIVER_PATH)
                    driver = EdgeDriver(service=driver_service)

                    driver.get(URL)
                    self.dismiss_alert_if_present(driver)

                    # Adjust the value of the input field
                    driver.execute_script(
                        'document.querySelector("#fromDate").value = "1/1/2010";')
                    WebDriverWait(driver, WEBDRIVER_WAIT_TIME).until(lambda driver: driver.execute_script(
                        'return document.querySelector("#fromDate").value;') == "1/1/2010")

                    # Find the button and click it
                    update_button = driver.execute_script(
                        'return document.querySelector("#command > div.filterbox > div.button-all")')
                    update_button.click()

                    # Wait for a couple of seconds after pressing the button
                    time.sleep(2)

                    # Wait for table to load
                    self.wait_for_table_to_load(driver)

                    page_num = 1

                    while len(df) < desired_rows:

                        df = self.extract_data_from_page(df, driver)
                        df['Date'] = pd.to_datetime(df['Date'])
                        df = df.drop_duplicates(subset='Date', keep='first')
                        if len(df) >= desired_rows:
                            break

                        self.navigate_to_next_page(driver)
                        page_num += 1
                    df['Date'] = pd.to_datetime(df['Date'])
                    df = df.sort_values(by='Date', ascending=True)

                else:  # if the CSV is freash fetch only the first page, to be modified to check for the number of rows

                    os.remove(filename)
                    # Initialization and existing code
                    driver = self.initialize_edge_driver(ticker)

                    # Adjust the value of the input field
                    driver.execute_script(
                        'document.querySelector("#fromDate").value = "1/1/2010";')
                    WebDriverWait(driver, WEBDRIVER_WAIT_TIME).until(lambda driver: driver.execute_script(
                        'return document.querySelector("#fromDate").value;') == "1/1/2010")

                    # Find the button and click it
                    update_button = driver.execute_script(
                        'return document.querySelector("#command > div.filterbox > div.button-all")')
                    update_button.click()

                    # Wait for a couple of seconds after pressing the button
                    time.sleep(2)

                    # Wait for table to load
                    self.wait_for_table_to_load(driver)

                    df = self.extract_data_from_page(df, driver)
                    df_temp = pd.concat([df, df_existing],
                                        axis=0, ignore_index=True)

                    df = df_temp
                    df['Date'] = pd.to_datetime(df['Date'])
                    duplicate_dates = df[df.duplicated(
                        subset='Date', keep='first')]['Date']

                    df = df.drop_duplicates(subset='Date', keep='first')

                    df = df.sort_values(by='Date', ascending=False)
                    df = df.head(desired_rows)
                    df = df.sort_values(by='Date', ascending=True)

            else:

                # Initialization and existing code

                URL = f'{BASE_URL}?currLanguage=en&companyCode={ticker}&activeTab=0'

                # Initialize Edge driver
                driver_service = Service(EDGE_DRIVER_PATH)
                driver = EdgeDriver(service=driver_service)

                driver.get(URL)
                self.dismiss_alert_if_present(driver)

                # Adjust the value of the input field
                driver.execute_script(
                    'document.querySelector("#fromDate").value = "1/1/2010";')
                WebDriverWait(driver, WEBDRIVER_WAIT_TIME).until(lambda driver: driver.execute_script(
                    'return document.querySelector("#fromDate").value;') == "1/1/2010")

                # Find the button and click it
                update_button = driver.execute_script(
                    'return document.querySelector("#command > div.filterbox > div.button-all")')
                update_button.click()

                # Wait for a couple of seconds after pressing the button
                time.sleep(2)

                # Wait for table to load
                self.wait_for_table_to_load(driver)

                page_num = 1

            while len(df) < desired_rows:

                # Extract data into a temporary DataFrame
                df_temp = self.extract_data_from_page(df, driver)

                # Append the temporary DataFrame to the main DataFrame
                df = pd.concat([df, df_temp], ignore_index=True)

                # Remove duplicates based on the 'Date' column
                df = df.drop_duplicates(subset='Date', keep='first')

                if len(df) >= desired_rows:
                    break

                self.navigate_to_next_page(driver)
                page_num += 1

            df = df.sort_values(by='Date', ascending=True)
            # This will keep only the latest 'desired_rows'
            df = df.tail(desired_rows)

            # Compute the actual change and change% based on the Close prices
            df['Change'] = df['Close'].diff()
            df['Change%'] = df['Change'] / df['Close'].shift(1) * 100
            # Round the values to two decimal places
            df['Change'] = df['Change'].round(2)
            df['Change%'] = df['Change%'].round(2)

            filename = f"raw_{ticker}.csv"
            df.to_csv(filename, index=False)
            return df

        except Exception as e:
            # Log the exact exception details
            self.logger.log_or_print(
                f"An error occurred while processing ticker {ticker}: {str(e)}", level="ERROR", exc_info=True)

            # Log the current state of the WebDriver
            current_url = driver.current_url
            self.logger.log_or_print(
                f"WebDriver state at error: Current URL = {current_url}", level="DEBUG")
            # Optional: Save a screenshot to see where the browser is when the error occurs
            # driver.save_screenshot('error_screenshot.png')

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
        try:
            WebDriverWait(driver, WEBDRIVER_WAIT_TIME).until(
                EC.presence_of_element_located((By.CSS_SELECTOR, "#dispTable")))
        except TimeoutException:
            self.logger.log_or_print(
                "Table did not load in time.", level="ERROR")
            raise

    def extract_data_from_page(self, df, driver):
        """Extract data from the current web page and append to DataFrame."""
        try:

            table_html = driver.execute_script(
                'return document.querySelector("#dispTable").outerHTML;')
            soup = BeautifulSoup(table_html, 'html.parser')
            table = soup.find('table')

            # Extracting data and appending to DataFrame
            for row in table.find_all('tr')[1:]:
                cols = row.find_all('td')
                date = datetime.strptime(
                    cols[9].text.strip(), '%d/%m/%Y').date()
                open_price = float(cols[8].text.strip().replace(',', ''))
                high = float(cols[7].text.strip().replace(',', ''))
                low = float(cols[6].text.strip().replace(',', ''))
                close = float(cols[5].text.strip().replace(',', ''))
                change = float(cols[4].text.strip().replace(',', ''))
                change_percent = float(
                    cols[3].text.strip().replace('%', '').replace(',', ''))
                t_shares = int(cols[2].text.strip().replace(',', ''))
                volume = int(cols[1].text.strip().replace(',', ''))
                no_trades = int(cols[0].text.strip().replace(',', ''))

                row_data = [date, open_price, high, low, close,
                            change, change_percent, t_shares, volume, no_trades]
                df.loc[len(df)] = row_data

            return df

        except Exception as e:
            self.logger.log_or_print(
                f"An error occurred while extracting data from the page: {str(e)}", level="ERROR", exc_info=True)
            return df  # Return the DataFrame as is

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
            
    def initialize_edge_driver(self, ticker):
        """
        Initializes and returns a Edge WebDriver for the specified ticker symbol.

        This function sets up the Edge WebDriver, navigates to the URL constructed
        using the ticker symbol, and handles any initial alerts that may appear on the page.

        Args:
            ticker (str): The ticker symbol for which the data is to be fetched.

        Returns:
            EdgeDriver: An instance of the Edge WebDriver ready to be used for web scraping.

        Raises:
            Exception: If any error occurs during the WebDriver initialization or navigation.
        """
        try:
            # Construct the URL with the ticker symbol
            URL = f'{BASE_URL}?currLanguage=en&companyCode={ticker}&activeTab=0'

            # Initialize the Edge WebDriver
            driver_service = Service(EDGE_DRIVER_PATH)
            driver = EdgeDriver(service=driver_service)

            # Navigate to the URL
            driver.get(URL)

            # Dismiss any initial alerts on the page
            self.dismiss_alert_if_present(driver)

            # Return the initialized WebDriver
            return driver

        except Exception as e:
            # Log the error with details
            self.logger.log_or_print(
                f"An error occurred during Edge WebDriver initialization for ticker {ticker}: {str(e)}",
                level="ERROR", exc_info=True)

            # Rethrow the exception to be handled by the caller
            raise

    def handle_existing_file(self, filename, desired_rows):
        """
        Handles the existing CSV file for a given ticker.

        Checks if the file exists and performs operations based on the file's data.
        If the file is outdated or has fewer rows than desired, it will be removed.

        Args:
            filename (str): The name of the CSV file to handle.
            desired_rows (int): The desired number of rows in the DataFrame.

        Returns:
            tuple: A tuple containing:
                - DataFrame from the existing file or None if the file doesn't exist.
                - A boolean indicating whether the file was outdated or had fewer rows.

        Raises:
            Exception: If any error occurs during file handling.
        """
        try:
            if os.path.exists(filename):
                df_existing = self.read_data_from_file(filename)
                number_of_rows = df_existing.shape[0]
                max_date = pd.to_datetime(df_existing['Date']).max().date()
                current_date = self.get_current_date()
                difference_in_days = (current_date - max_date).days

                if difference_in_days > 20 or number_of_rows < desired_rows - 20:
                    os.remove(filename)
                    self.logger.log_or_print(f"Removed outdated or insufficient data file: {filename}", level="INFO")
                    return df_existing, True
                else:
                    return df_existing, False
            return None, False
        except Exception as e:
            self.logger.log_or_print(f"An error occurred while handling the file {filename}: {str(e)}", level="ERROR")
            raise
