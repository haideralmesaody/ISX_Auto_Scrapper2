import pandas as pd
from datetime import datetime
from LoggerFunction import Logger  # Ensure this import matches your logger setup
import os
class LiquidityCalculator():

    def __init__(self):
        super().__init__()
        self.logger = Logger()

    def calculate_liquidity_score(self):
        """
        Calculate liquidity scores for stocks listed in 'TICKERS.csv' based on their trading
        activity over the past 12 months, excluding the top 5% and lowest 5% of days by trading volume,
        considering volume, trading frequency, and volatility, and saves the results to 'liquidity_scores.csv'.
        """
        tickers_df = pd.read_csv('TICKERS.csv')
        liquidity_scores = pd.DataFrame(columns=['Ticker', 'Average Volume' , 'Average Traded Volume','Volume STD',  'Days Traded','Days Traded2','Trading Activity Score', 'Relative Volume Score' ,'Liqudity Score'])

        for ticker in tickers_df['Ticker']:
            df_path = f'raw_{ticker}.csv'
            if not os.path.exists(df_path):
                self.logger.log_or_print(f"Data file for {ticker} does not exist.", level="WARNING")
                continue

            df = pd.read_csv(df_path, parse_dates=['Date'])
            df.sort_values('Date', inplace=True)

            one_year_ago = datetime.now() - pd.DateOffset(years=1)
            #ge only the rows within the previous 12 months
            df_last_12_months = df[df['Date'] > one_year_ago]

            if df_last_12_months.empty:
                self.logger.log_or_print(f"No trading data for {ticker} in the past 12 months.", level="WARNING")
                continue
            #get traded days for the last 12 months
            days_traded = df_last_12_months['Date'].nunique()
            #calculate average volume traded
            Average_Volume = df_last_12_months['Volume'].mean()
            #get the standard deviation of the close price
            Volume_SDT = df_last_12_months['Volume'].pct_change().std()
            #keep the data within plus minus 2 standard deviation of the average volume
            #df_last_12_months = df_last_12_months[(df_last_12_months['Volume'] > (Average_Volume - 3 * Volume_SDT)) & (df_last_12_months['Volume'] < (Average_Volume + 3 * Volume_SDT))]
            #remove the to and bottom 5% of the volume
            df_last_12_months = df_last_12_months[(df_last_12_months['Volume'] > df_last_12_months['Volume'].quantile(0.05)) & (df_last_12_months['Volume'] < df_last_12_months['Volume'].quantile(0.95))]
            #get the number of days traded after removing the outliers
            days_traded2 = df_last_12_months['Date'].nunique()
            #get the average volume
            average_volume_traded = df_last_12_months['Volume'].mean()
            #pring or log the average volume
            self.logger.log_or_print(f"Average volume traded for {ticker} is {average_volume_traded}", level="INFO")
            #calculate trading activity score
            trading_activity_score = days_traded / 252
            #if trading days less than 100, set trading activity score to 0
            if days_traded < 100:
                trading_activity_score = 0
            
            new_row = {
                'Ticker': ticker, 
                'Average Volume': Average_Volume,
                'Average Traded Volume': average_volume_traded,
                'Volume STD': Volume_SDT,
                'Days Traded': days_traded,
                'Days Traded2': days_traded2,
                'Trading Activity Score': trading_activity_score,
                'Relative Volume Score': 0,
                'Liqudity Score': 0,
                'Liquidity Score%': 0

            }
            liquidity_scores = pd.concat([liquidity_scores, pd.DataFrame([new_row])], ignore_index=True)
        #calculate relative volume score
        liquidity_scores['Relative Volume Score'] = liquidity_scores['Average Traded Volume'] / liquidity_scores['Average Traded Volume'].max()
        #calculate liquidity score
        liquidity_scores['Liqudity Score'] = (liquidity_scores['Trading Activity Score'] * liquidity_scores['Relative Volume Score'])*100
        #calculate liquidity score percentage by dividing liquidity score by the sum of liquidity score column
        liquidity_scores['Liquidity Score%'] = liquidity_scores['Liqudity Score'] / liquidity_scores['Liqudity Score'].sum()


        liquidity_scores.to_csv('liquidity_scores.csv', index=False)
        self.logger.log_or_print("Liquidity scores successfully calculated and saved.", level="INFO")

