#!/usr/bin/python3
# -*- coding: utf-8 -*-
import numpy as np
import pandas as pd
import argparse
import hashlib
import json
import talib as ta
try:
    import pandas_ta as pta
except ImportError:  # current maintained package name
    import pandas_ta_classic as pta
import MyTT as mytt

tcol, ocol, hcol, lcol, ccol, vcol = 0, 1, 2, 3, 4, 5

with open("testdata/fixture_btc_58.json", "rb") as fixture_file:
    fixture_bytes = fixture_file.read()
    candles = json.loads(fixture_bytes)
fixture_sha256 = hashlib.sha256(fixture_bytes).hexdigest()
ohlcv_arr = np.array(candles)
open_arr, high_arr, low_arr = ohlcv_arr[:, ocol], ohlcv_arr[:, hcol], ohlcv_arr[:, lcol]
close_arr, vol_arr = ohlcv_arr[:, ccol], ohlcv_arr[:, vcol]
bar_idx = list(range(len(ohlcv_arr)))
open_col = pd.Series(open_arr, index=bar_idx)
high_col = pd.Series(high_arr, index=bar_idx)
low_col = pd.Series(low_arr, index=bar_idx)
close_col = pd.Series(close_arr, index=bar_idx)
vol_col = pd.Series(vol_arr, index=bar_idx)
comparison_result = {}


def _json_values(value):
    if isinstance(value, (tuple, list)):
        return [_json_values(x) for x in value]
    if hasattr(value, 'to_numpy'):
        value = value.to_numpy()
    if isinstance(value, np.ndarray):
        return [None if not np.isfinite(x) else float(x) for x in value]
    return value


def print_tares(ta_cres=None, ta_mres=None, mytt_res=None, pta_res=None):
    comparison_result.clear()
    for label, value in (('talib', ta_cres), ('talib_compat', ta_mres),
                         ('mytt', mytt_res), ('pandas_ta', pta_res)):
        if value is not None:
            comparison_result[label] = _json_values(value)
    if ta_cres is not None:
        print('\n' + ' Ta-lib Classic '.center(60, '='))
        print(ta_cres)
    if ta_mres is not None:
        print('\n' + ' Ta-lib MetaStock '.center(60, '='))
        print(ta_mres)
    if mytt_res is not None:
        print('\n' + ' MyTT '.center(60, '='))
        print(mytt_res)
    if pta_res is not None:
        if hasattr(pta_res, 'to_numpy'):
            pta_res = pta_res.to_numpy()
        print('\n' + ' Pandas-TA '.center(60, '='))
        print(pta_res)


def test_sma():
    period = 5
    ta.set_compatibility(1)
    ta_res = ta.SMA(close_arr, timeperiod=period)
    ta.set_compatibility(0)
    ta2_res = ta.SMA(close_arr, timeperiod=period)
    # mytt的SMA初始120周期不精确
    mtt_res = mytt.SMA(close_arr, period)
    mtt_res = np.array(mtt_res)
    pta_res = pta.sma(close_col, period, talib=False)
    print_tares(ta_res, ta2_res, mtt_res, pta_res)
    

def test_ema():
    period = 12
    ta.set_compatibility(1)
    ta1_res = ta.EMA(close_arr, timeperiod=period)
    ta.set_compatibility(0)
    ta2_res = ta.EMA(close_arr, timeperiod=period)
    mtt_res = mytt.EMA(close_arr, period)
    pta_res = pta.ema(close_col, period, talib=False)
    print_tares(ta1_res, ta2_res, mtt_res, pta_res)
    

def test_rma():
    period = 12
    pta_res = pta.rma(close_col, period, talib=False)
    print_tares(pta_res=pta_res)
    

def test_tr():
    ta_res = ta.TRANGE(high_arr, low_arr, close_arr)
    pta_res = pta.true_range(high_col, low_col, close_col, talib=False)
    print_tares(ta_res, pta_res=pta_res)
    

def test_atr():
    period = 14
    mtt_res = mytt.ATR(close_arr, high_arr, low_arr, period)
    ta.set_compatibility(1)
    ta2_res = ta.ATR(high_arr, low_arr, close_arr, timeperiod=period)
    ta.set_compatibility(0)
    ta_res = ta.ATR(high_arr, low_arr, close_arr, timeperiod=period)
    pta_res = pta.atr(high_col, low_col, close_col, period, talib=False)
    print_tares(ta_res, ta2_res, mytt_res=mtt_res, pta_res=pta_res)
    

def test_macd():
    ta.set_compatibility(1)
    ta_mres = ta.MACD(close_arr, fastperiod=12, slowperiod=26, signalperiod=9)[0]
    ta.set_compatibility(0)
    ta_cres = ta.MACD(close_arr, fastperiod=12, slowperiod=26, signalperiod=9)[0]
    mtt_res = mytt.MACD(close_arr)[0]
    pta_res = pta.macd(close_col, 12, 26, 9, talib=False)['MACD_12_26_9'].to_numpy()
    print_tares(ta_cres, ta_mres, mtt_res, pta_res)
    

def test_rsi():
    '''
    和Ta-lib、Pandas-TA的计算一致
    MyTT的RSI受SMA影响，前120周期不准确
    '''
    period = 14
    ta.set_compatibility(1)
    ta_res = ta.RSI(close_arr, timeperiod=period)
    ta.set_compatibility(0)
    ta2_res = ta.RSI(close_arr, timeperiod=period)
    # MyTT的RSI因SMA影响，初始120周期不精确
    mtt_res = mytt.RSI(close_arr, period)
    pta_res = pta.rsi(pd.Series(close_arr), period, talib=False).to_numpy()
    print_tares(ta2_res, ta_res, mtt_res, pta_res)
    

def test_vwma():
    period = 9
    price_col = pta.hlc3(close_col, high_col, low_col)
    ta_res = pta.vwma(price_col, vol_col, period, talib=False).to_numpy()
    print(' vwma Pandas Res '.center(60, '='))
    print(ta_res)


def test_kdj():
    '''
    最流行的KDJ算法中，平滑应该使用RMA
    中国主流软件和MyTT使用EMA(2*period-1)且init_type=1。
    ta-lib中KDJ的平滑支持很多种方式，通过slowk_matype指定，默认的0是SMA，1是EMA；
    https://developer.hs.net/thread/2321
    '''
    # 这里使用2*period-1，EMA平滑，保持和MyTT一致
    ta_kdj_args = dict(fastk_period=9, slowk_period=5, slowd_period=5, slowk_matype=1, slowd_matype=1)
    ta.set_compatibility(1)
    ta_k, ta_d = ta.STOCH(high_arr, low_arr, close_arr, **ta_kdj_args)
    ta.set_compatibility(0)
    ta2_k, ta2_d = ta.STOCH(high_arr, low_arr, close_arr, **ta_kdj_args)
    # 使用mytt计算
    mk, mt, mj = mytt.KDJ(close_arr, high_arr, low_arr)
    pta_df = pta.kdj(high_col, low_col, close_col, 9, 3, talib=False)
    pta_k, pta_d, pta_j = pta_df['K_9_3'], pta_df['D_9_3'], pta_df['J_9_3']
    print_tares(ta2_k, ta_k, mk, pta_k)


def test_stoch():
    import talib.abstract as tb
    ta.set_compatibility(1)
    df = pd.DataFrame(dict(high=high_col, low=low_col, close=close_col))
    cols = tb.STOCHF(df, 5,3,0,3,0)
    ta.set_compatibility(0)
    cols2 = tb.STOCHF(df, 5,3,0,3,0)
    print_tares(cols2['fastk'].to_numpy(), cols['fastk'].to_numpy())


def test_bband():
    ta.set_compatibility(1)
    period, nbdevup, nbdevdn = 9, 2, 2
    ta_up, ta_md, ta_lo = ta.BBANDS(close_arr, timeperiod=period, nbdevup=nbdevup, nbdevdn=nbdevdn)
    ta.set_compatibility(0)
    ta2_up, ta2_md, ta2_lo = ta.BBANDS(close_arr, timeperiod=period, nbdevup=nbdevup, nbdevdn=nbdevdn)
    # 使用mytt计算
    m_up, m_md, m_lo = mytt.BOLL(close_arr, period, nbdevup)
    # pta 计算
    pta_df = pta.bbands(close_col, period, nbdevup, talib=False)
    pta_up, pta_md, pta_lo = pta_df['BBU_9_2.0'], pta_df['BBM_9_2.0'], pta_df['BBL_9_2.0']
    # 对比
    print_tares(ta2_up, ta_up, m_up, pta_up)
    

def test_adx():
    ta.set_compatibility(1)
    period = 9
    ta_res = ta.ADX(high_arr, low_arr, close_arr, timeperiod=period)
    ta.set_compatibility(0)
    ta2_res = ta.ADX(high_arr, low_arr, close_arr, timeperiod=period)
    pta_res = pta.adx(high_col, low_col, close_col, period, talib=False).to_numpy()
    print_tares(ta_res, ta2_res, None, pta_res)


def test_minusDi():
    ta.set_compatibility(1)
    period = 9
    ta_res = ta.MINUS_DI(high_arr, low_arr, close_arr, timeperiod=period)
    print_tares(ta_res)


def test_pluMinDm():
    ta.set_compatibility(1)
    period = 9
    ta_res = ta.PLUS_DM(high_arr, low_arr, timeperiod=period)
    print_tares(ta_res)


def test_roc():
    ta.set_compatibility(1)
    period = 9
    mytt_res = mytt.ROC(close_arr, period)[0]
    ta_res = ta.ROC(close_arr, timeperiod=period)
    ta.set_compatibility(0)
    ta2_res = ta.ROC(close_arr, timeperiod=period)
    print_tares(ta_res, ta2_res, mytt_res)


def test_cci():
    ta.set_compatibility(1)
    period = 10
    mytt_res = mytt.CCI(close_arr, high_arr, low_arr, period)
    ta_res = ta.CCI(high_arr, low_arr, close_arr, timeperiod=period)
    ta.set_compatibility(0)
    ta2_res = ta.CCI(high_arr, low_arr, close_arr, timeperiod=period)
    pta_res = pta.cci(high_col, low_col, close_col, period, talib=False).to_numpy()
    print_tares(ta_res, ta2_res, mytt_res, pta_res)


def test_cmf():
    ta.set_compatibility(1)
    period = 10
    pta_res = pta.cmf(high_col, low_col, close_col, vol_col, length=period).to_numpy()
    print(pta_res)


def test_kama():
    ta.set_compatibility(1)
    period = 10
    ta_res = ta.KAMA(close_arr, timeperiod=period)
    ta.set_compatibility(0)
    ta2_res = ta.KAMA(close_arr, timeperiod=period)
    pta_res = pta.kama(close_col, length=period).to_numpy()
    print_tares(ta_res, ta2_res, None, pta_res)


def test_will_r():
    ta.set_compatibility(1)
    period = 10
    ta_res = ta.WILLR(high_arr, low_arr, close_arr, timeperiod=period)
    ta.set_compatibility(0)
    ta2_res = ta.WILLR(high_arr, low_arr, close_arr, timeperiod=period)
    pta_res = pta.willr(high_col, low_col, close_col, length=period).to_numpy()
    print_tares(ta_res, ta2_res, None, pta_res)


def test_stoch_rsi():
    ta.set_compatibility(1)
    period = 9
    rsi = ta.RSI(close_arr, timeperiod=period)
    ta_res = ta.STOCH(rsi, rsi, rsi, fastk_period=9, slowk_period=3, slowk_matype=0, slowd_period=3, slowd_matype=0)[0]
    ta.set_compatibility(0)
    rsi = ta.RSI(close_arr, timeperiod=period)
    ta2_res = ta.STOCH(rsi, rsi, rsi, fastk_period=9, slowk_period=3, slowk_matype=0, slowd_period=3, slowd_matype=0)[0]
    pta_res = pta.stochrsi(close_col, length=9, rsi_length=period, k=3, d=3).to_numpy()[:, 0]
    print_tares(ta_res, ta2_res, None, pta_res)


def test_mfi():
    ta.set_compatibility(1)
    period = 10
    ta_res = ta.MFI(high_arr, low_arr, close_arr, vol_arr, timeperiod=period)
    ta.set_compatibility(0)
    ta2_res = ta.MFI(high_arr, low_arr, close_arr, vol_arr, timeperiod=period)
    mytt_res = mytt.MFI(high_arr, low_arr, close_arr, vol_arr, period)
    pta_res = pta.mfi(high_col, low_col, close_col, vol_col, length=period, talib=False).to_numpy()
    print_tares(ta_res, ta2_res, mytt_res, pta_res)


def test_aroon():
    ta.set_compatibility(1)
    period = 9
    ta_res = ta.AROON(high_arr, low_arr, timeperiod=period)
    ta.set_compatibility(0)
    ta2_res = ta.AROON(high_arr, low_arr, timeperiod=period)
    arn_res = pta.aroon(high_col, low_col, period).to_numpy()
    print(ta_res)
    print(ta2_res)
    print(arn_res)


def test_wma():
    ta.set_compatibility(1)
    period = 10
    ta_res = ta.WMA(close_arr, timeperiod=period)
    ta.set_compatibility(0)
    ta2_res = ta.WMA(close_arr, timeperiod=period)
    mytt_res = mytt.WMA(close_arr, period)
    pta_res = pta.wma(close_col, period).to_numpy()
    print_tares(ta_res, ta2_res, mytt_res, pta_res)


def test_hma():
    period = 10
    pta_res = pta.hma(close_col, period).to_numpy()
    print(pta_res)


def test_cti():
    period = 10
    pta_res = pta.cti(close_col, period).to_numpy()
    print(pta_res)


def test_lingreg():
    period = 10
    pta_res = pta.linreg(close_col, period).to_numpy()
    print(pta_res)


def test_cmo():
    ta.set_compatibility(1)
    period = 10
    ta_res = ta.CMO(close_arr, timeperiod=period)
    ta.set_compatibility(0)
    ta2_res = ta.CMO(close_arr, timeperiod=period)
    pta_res = pta.cmo(close_col, period).to_numpy()
    print_tares(ta_res, ta2_res, None, pta_res)


def test_chop():
    period = 10
    pta_res = pta.chop(high_col, low_col, close_col, period).to_numpy()
    print_tares(None, None, None, pta_res)


def test_alma():
    period = 10
    pta_res = pta.alma(close_col, period).to_numpy()
    print_tares(None, None, None, pta_res)


def test_crsi():
    chg = close_col / close_col.shift(1)
    updown = np.where(chg.gt(1), 1.0, np.where(chg.lt(1), -1.0, 0.0))
    rsi = ta.RSI(close_arr, timeperiod=3)
    ud = ta.RSI(updown, timeperiod=2)
    roc = ta.ROC(close_arr, 20)
    crsi = (rsi + ud + roc) / 3
    print_tares(crsi)


def dv2(c, h, l, length=126):
    d1 = (c/((h+l)/2))-1
    dv = d1.rolling(window=2).mean()
    return dv.rolling(window=length).apply(lambda x: pd.Series(x).rank(pct=True).iloc[-1])*100

def test_dv2():
    dv2_res = dv2(close_col, high_col, low_col, 30)
    print_tares(None, None, None, dv2_res)


def test_ao():
    """Awesome Oscillator: SMA(HL2, 5) - SMA(HL2, 34)."""
    pta_res = pta.ao(high_col, low_col, fast=5, slow=34, talib=False)
    print_tares(None, None, None, pta_res)


def test_adosc():
    """Chaikin A/D oscillator; compare TA-Lib and pandas-ta when available."""
    ta_res = ta.ADOSC(high_arr, low_arr, close_arr, vol_arr, fastperiod=3, slowperiod=10)
    pta_res = pta.adosc(high_col, low_col, close_col, vol_col, fast=3, slow=10, talib=False)
    print_tares(ta_res, pta_res=pta_res)


def test_ht_sine():
    sine, leadsine = ta.HT_SINE(close_arr)
    print_tares((sine, leadsine))


def test_linearreg_angle():
    period = 14
    ta_res = ta.LINEARREG_ANGLE(close_arr, timeperiod=period)
    pta_res = pta.linreg(close_col, length=period, angle=True, talib=False)
    print_tares(ta_res, pta_res=pta_res)


def test_dpo():
    period = 20
    pta_res = pta.dpo(close_col, length=period, centered=False, talib=False)
    print_tares(None, None, None, pta_res)


def test_median():
    period = 14
    pta_res = pta.median(close_col, length=period, talib=False)
    print_tares(None, None, None, pta_res)


def test_efi():
    period = 13
    pta_res = pta.efi(close_col, vol_col, length=period, talib=False)
    print_tares(None, None, None, pta_res)


def test_mom():
    period = 2
    ta_res = ta.MOM(close_arr, timeperiod=period)
    pta_res = pta.mom(close_col, length=period, talib=False)
    print_tares(ta_res, pta_res=pta_res)


def test_obv():
    ta_res = ta.OBV(close_arr, vol_arr)
    pta_res = pta.obv(close_col, vol_col, talib=False)
    print_tares(ta_res, pta_res=pta_res)


def test_dema():
    period = 3
    ta_res = ta.DEMA(close_arr, timeperiod=period)
    pta_res = pta.dema(close_col, length=period, talib=False)
    print_tares(ta_res, pta_res=pta_res)


def test_t3():
    period = 3
    ta_res = ta.T3(close_arr, timeperiod=period, vfactor=0.7)
    pta_res = pta.t3(close_col, length=period, a=0.7, talib=False)
    print_tares(ta_res, pta_res=pta_res)


def test_aroonosc():
    period = 9
    ta_res = ta.AROONOSC(high_arr, low_arr, timeperiod=period)
    pta_df = pta.aroon(high_col, low_col, length=period, talib=False)
    pta_res = pta_df.iloc[:, 0] - pta_df.iloc[:, 1]
    print_tares(ta_res, pta_res=pta_res)


def test_stochf():
    k, d = ta.STOCHF(high_arr, low_arr, close_arr, fastk_period=5,
                     fastd_period=3, fastd_matype=0)
    print_tares((k, d))


def test_ultosc():
    ta_res = ta.ULTOSC(high_arr, low_arr, close_arr, timeperiod1=7,
                       timeperiod2=14, timeperiod3=28)
    pta_res = pta.uo(high_col, low_col, close_col, fast=7, medium=14,
                     slow=28, talib=False)
    print_tares(ta_res, pta_res=pta_res)


def test_sar():
    ta_res = ta.SAR(high_arr, low_arr, acceleration=0.02, maximum=0.2)
    print_tares(ta_res)


def test_rocr():
    period = 9
    ta_res = ta.ROCR(close_arr, timeperiod=period)
    pta_res = pta.rocr(close_col, length=period, talib=False)
    print_tares(ta_res, pta_res=pta_res)


def test_natr():
    period = 14
    ta_res = ta.NATR(high_arr, low_arr, close_arr, timeperiod=period)
    pta_res = pta.natr(high_col, low_col, close_col, length=period, talib=False)
    print_tares(ta_res, pta_res=pta_res)


def test_trima():
    period = 10
    ta_res = ta.TRIMA(close_arr, timeperiod=period)
    pta_res = pta.trima(close_col, length=period, talib=False)
    print_tares(ta_res, pta_res=pta_res)


def test_swma():
    pta_res = pta.swma(close_col, length=4, talib=False)
    print_tares(None, None, None, pta_res)


def test_zlma():
    period = 10
    pta_res = pta.zlma(close_col, length=period, talib=False)
    print_tares(None, None, None, pta_res)


def test_pivothigh():
    print('No callable TA-Lib/pandas-ta implementation; Pine pivot definition is the oracle.')


def test_pivotlow():
    print('No callable TA-Lib/pandas-ta implementation; Pine pivot definition is the oracle.')

def test_ichimoku():
    fn = getattr(pta, 'ichimoku', None)
    print_tares(None, None, None, fn(high_col, low_col, close_col) if fn else None)

def test_mama():
    fn = getattr(ta, 'MAMA', None)
    print_tares(fn(close_arr) if fn else None)

def test_kst():
    fn = getattr(pta, 'kst', None)
    print_tares(None, None, None, fn(close_col) if fn else None)

def test_donchian_pband():
    fn = getattr(pta, 'donchian', None)
    d = fn(high_col, low_col, close_col) if fn else None
    print_tares(None, None, None, d.iloc[:, -1] if d is not None else None)

def test_keltner_wband():
    fn = getattr(pta, 'kc', None)
    d = fn(high_col, low_col, close_col) if fn else None
    print_tares(None, None, None, d.iloc[:, -1] if d is not None else None)

def test_vpci():
    fn = getattr(pta, 'vpci', None)
    print_tares(None, None, None, fn(close_col, vol_col) if fn else None)

def test_williams_percent():
    r = ta.WILLR(high_arr, low_arr, close_arr, timeperiod=14)
    print_tares(r)

def test_dx():
    print_tares(ta.DX(high_arr, low_arr, close_arr, timeperiod=14))

def test_fisher():
    fn = getattr(pta, 'fisher', None)
    print_tares(None, None, None, fn(high_col, low_col) if fn else None)

def test_correlation():
    print_tares(pd.Series(close_arr).rolling(14).corr(pd.Series(vol_arr)))


TESTS = {name[5:]: fn for name, fn in globals().items() if name.startswith('test_') and callable(fn)}

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Compare indicators across TA-Lib/pandas-ta/MyTT.')
    parser.add_argument('--indicator', choices=sorted(TESTS), default='mfi')
    parser.add_argument('--list', action='store_true', help='list callable comparison tests')
    parser.add_argument('--json', metavar='PATH', help='save the selected comparison result as JSON')
    args = parser.parse_args()
    if args.list:
        print('\n'.join(sorted(TESTS)))
    else:
        TESTS[args.indicator]()
        if args.json:
            libraries = {}
            for label, module in (('talib', ta), ('pandas_ta', pta), ('MyTT', mytt)):
                libraries[label] = getattr(module, '__version__', getattr(module, 'version', 'unknown'))
            with open(args.json, 'w', encoding='utf-8') as f:
                json.dump({'indicator': args.indicator, 'fixture': 'fixture_btc_58.json',
                           'fixture_sha256': fixture_sha256,
                           'fixture_bars': len(candles), 'libraries': libraries,
                           'results': comparison_result},
                          f, allow_nan=False, indent=2)
