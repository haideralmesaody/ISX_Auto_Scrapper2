from strategy_common import BaseStrategies

class Strategies(BaseStrategies):
    """Wrapper for subset of strategies with trailing stops."""

    def TradingStrategies(self, df):
        if df is None or df.empty:
            raise ValueError("The input DataFrame is either None or empty.")
        try:
            df['RSI Strategy'] = 'Hold'
            df = self.apply_rsi_with_trailing_stop(df, atr_multiplier=1)
            df['RSI Strategy2'] = 'Hold'
            df = self.apply_rsi2_with_trailing_stop(df, atr_multiplier=2)
            df['RSI14_OBV_RoC Strategy'] = 'Hold'
            df = self.apply_rsi14_obv_roc_with_trailing_stop(df, atr_multiplier=1)
            df['RSIMACD Strategy'] = 'Hold'
            df = self.apply_rsimacd_with_trailing_stop(df, atr_multiplier=1)
            df['RSICMF Strategy'] = 'Hold'
            df = self.apply_rsicmf_with_trailing_stop(df, atr_multiplier=1)
            df['RSI OBV Strategy'] = 'Hold'
            df = self.apply_rsi_obv_with_trailing_stop(df, atr_multiplier=1)
            df['OBV Strategy'] = 'Hold'
            df = self.apply_obv_with_trailing_stop(df, atr_multiplier=1)
            df['MACD Strategy'] = 'Hold'
            df = self.apply_macd_with_trailing_stop(df, atr_multiplier=1)
            df['CMF Strategy'] = 'Hold'
            df = self.apply_cmf_with_trailing_stop(df, atr_multiplier=1)
            df['EMA5 PSAR Strategy'] = 'Hold'
            df = self.apply_ema5_psar_with_trailing_stop(df, atr_multiplier=1)
            self.logger.log_or_print('Trading strategies successfully added.', level='INFO')
            return df
        except Exception as e:
            self.logger.log_or_print(f'Error adding trading strategies: {e}', level='ERROR')
            raise
