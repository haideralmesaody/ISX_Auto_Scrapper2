from data_fetcher import DataFetcher
from app_config import (
    LOGGING_CONFIG, EDGE_DRIVER_PATH, BASE_URL, DEFAULT_DATE,
    TABLE_SELECTOR, WEBDRIVER_WAIT_TIME, DEFAULT_SMA_PERIOD,
    DEFAULT_ROW_COUNT, EXCEL_ENGINE
)
def main():
    data_fetcher = DataFetcher(driver_path=EDGE_DRIVER_PATH)
    ticker = 'TASC'  # replace with your ticker
    desired_rows = 100  # replace with your desired number of rows
    df = data_fetcher.fetch_data(ticker, desired_rows)
    print(df)

if __name__ == "__main__":
    main()