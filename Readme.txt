# ISX Auto Scrapper

## Description
ISX Auto Scrapper is a command-line tool designed to keep stock data updated by scraping information from the website [http://www.isx-iq.net/isxportal/portal/homePage.html](http://www.isx-iq.net/isxportal/portal/homePage.html). It utilizes the data provided in the sheet named TICKERS.csv to update the list of ticker/stock data.

## Installation
1. Clone the repository:
    ```shell
    git clone https://github.com/your-username/ISX-Auto-Scrapper.git
    ```

2. Install the required dependencies:
    ```shell
    npm install
    ```

## Usage
1. Make sure the TICKERS.csv file is present in the project directory and contains the list of ticker/stock data to be updated.

2. Run the following command to start the auto scrapper:
    ```shell
    node index.js
    ```

3. The tool will automatically scrape the stock data from the website and update the relevant information in the TICKERS.csv file.

## Contributing
Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.

## License
This project is licensed under the [MIT License](LICENSE).
