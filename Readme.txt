# ISX Auto Scrapper

## Description
ISX Auto Scrapper is a command-line tool designed to keep stock data updated by scraping information from the website [http://www.isx-iq.net/isxportal/portal/homePage.html](http://www.isx-iq.net/isxportal/portal/homePage.html). It utilizes the data provided in the sheet named TICKERS.csv to update the list of ticker/stock data.

## Installation
1. Clone the repository:
    ```shell
    git clone https://github.com/your-username/ISX-Auto-Scrapper.git
    ```

2. Install the required Python dependencies:
    ```shell
    pip install -r requirements.txt
    ```

## Usage
1. Make sure the TICKERS.csv file is present in the project directory and contains the list of ticker/stock data to be updated.

2. Run the following command to start the auto scrapper:
    ```shell
    python main.py --mode <option>
    ```
    Replace `<option>` with one of the modes listed below.

3. The tool will automatically scrape the stock data from the website and update the relevant information in the TICKERS.csv file.

### Available Modes
- `single`: Fetch data for a single ticker interactively.
- `auto`: Process all tickers from `TICKERS.csv` and run the full analysis.
- `liquidity`: Calculate liquidity scores only.
- `strategies`: Apply predefined strategies to recent data.
- `backtest`: Backtest strategies and summarize results.
- `simulate`: Simulate strategy outcomes.
- `calculate_num`: Generate numeric indicator files without descriptions.
- `breakout`, `train_tensor`, `stock_predictor`, `predict_close_price`: Reserved for future features.

## Contributing
Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.

## License
This project is licensed under the [MIT License](LICENSE).
