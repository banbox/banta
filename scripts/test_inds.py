#!/usr/bin/python3
# -*- coding: utf-8 -*-
import numpy as np
import pandas as pd
import talib as ta
import pandas_ta as pta
import MyTT as mytt

tcol, ocol, hcol, lcol, ccol, vcol = 0, 1, 2, 3, 4, 5

candles = [
    # 数据来自币安合约BTC/USDT.P 2023/07/01 - 2023/08/27
    [1688169600000, 30460.2, 30668.2, 30311.3, 30573.6, 135520.246],
    [1688256000000, 30573.6, 30800, 30149.9, 30612.7, 231866.18800000002],
    [1688342400000, 30612.7, 31395.2, 30559.4, 31149, 370293.2360000001],
    [1688428800000, 31149, 31319.4, 30600, 30756.1, 300832.26],
    [1688515200000, 30756.1, 30875.2, 30175.8, 30488.4, 294896.578],
    [1688601600000, 30488.4, 31568, 29818, 29874.4, 721267.2189999999],
    [1688688000000, 29874.3, 30443.6, 29680, 30327.9, 337801.65400000004],
    [1688774400000, 30328, 30380, 30026.8, 30269.3, 138734.49300000002],
    [1688860800000, 30269.2, 30443.4, 30042, 30147.8, 162296.263],
    [1688947200000, 30147.8, 31040, 29928.8, 30396.9, 429115.53699999995],
    [1689033600000, 30396.9, 30804.9, 30261.4, 30608.4, 298904.7469999999],
    [1689120000000, 30608.3, 30980.5, 30186, 30368.9, 425058.257],
    [1689206400000, 30368.9, 31850, 30233, 31441.7, 696023.5100000001],
    [1689292800000, 31441.6, 31640, 29876.6, 30293.3, 538692.2459999999],
    [1689379200000, 30293.3, 30380, 30220.2, 30276.4, 111622.47400000002],
    [1689465600000, 30276.5, 30441.6, 30050, 30216.8, 173805.94],
    [1689552000000, 30216.9, 30329.7, 29630, 30126.1, 344113.42699999997],
    [1689638400000, 30126, 30227.5, 29400, 29845.6, 337396.417],
    [1689724800000, 29845.7, 30178.5, 29745, 29895.5, 298418.854],
    [1689811200000, 29895.5, 30420, 29468.8, 29791, 410263.76399999997],
    [1689897600000, 29791, 30056.2, 29705.9, 29891.4, 221145.07399999994],
    [1689984000000, 29891.5, 29986, 29602.2, 29783.5, 143375.95600000003],
    [1690070400000, 29783.5, 30368.2, 29718.9, 30070.8, 224713.84599999996],
    [1690156800000, 30070.9, 30091.7, 28830, 29163.8, 434182.417],
    [1690243200000, 29163.8, 29379.5, 29033, 29216.3, 196438.618],
    [1690329600000, 29216.2, 29670.5, 29083, 29336, 318527.58199999994],
    [1690416000000, 29336, 29558.4, 29065.5, 29209.7, 223376.53400000007],
    [1690502400000, 29209.8, 29535.7, 29112.8, 29299.9, 221299.86699999997],
    [1690588800000, 29300, 29397.9, 29237.5, 29339.1, 87100.719],
    [1690675200000, 29339.1, 29442.5, 29006, 29271.1, 169901.253],
    [1690761600000, 29271.2, 29525.8, 29100, 29220.8, 230381.72199999998],
    [1690848000000, 29220.8, 29735, 28550, 29701.2, 442328.643],
    [1690934400000, 29701.2, 30059.9, 28906.3, 29170.1, 491725.49199999997],
    [1691020800000, 29170.2, 29440, 28955.4, 29180.2, 251967.94199999998],
    [1691107200000, 29180.1, 29323.8, 28780, 29101.1, 229395.67300000004],
    [1691193600000, 29101.1, 29145, 28960, 29057.7, 82508.624],
    [1691280000000, 29057.7, 29199, 28978.5, 29075.9, 103094.43999999999],
    [1691366400000, 29075.9, 29274.5, 28682.3, 29202.7, 297752.4390000001],
    [1691452800000, 29202.7, 30250, 29132.4, 29759, 459269.80100000004],
    [1691539200000, 29758.9, 30149.7, 29362, 29572.8, 362009.70100000006],
    [1691625600000, 29572.9, 29729.8, 29303.7, 29443.7, 234546.03800000003],
    [1691712000000, 29443.7, 29565.1, 29220, 29415.5, 179044.008],
    [1691798400000, 29415.5, 29470, 29360, 29420.7, 56238.034999999996],
    [1691884800000, 29420.8, 29463, 29256.6, 29293.3, 84190.505],
    [1691971200000, 29293.3, 29686.7, 29070, 29419.5, 295035.38300000003],
    [1692057600000, 29419.5, 29492.1, 29050, 29188.8, 191289.959],
    [1692144000000, 29188.9, 29257.4, 28705.1, 28714.4, 280543.0889999999],
    [1692230400000, 28714.4, 28775.9, 24581, 26609.7, 868508.2199999996],
    [1692316800000, 26609.7, 26818, 25600, 26042.1, 522375.31599999993],
    [1692403200000, 26042, 26269.4, 25783.4, 26088.3, 202219.153],
    [1692489600000, 26088.4, 26285, 25948.5, 26175.9, 141089.85],
    [1692576000000, 26175.9, 26280, 25800, 26115.4, 232505.26499999993],
    [1692662400000, 26115.4, 26128.2, 25280, 26044.4, 313113.576],
    [1692748800000, 26044.5, 26806, 25800, 26419.2, 412969.02800000005],
    [1692835200000, 26419.2, 26568.3, 25835, 26164.6, 286753.33699999994],
    [1692921600000, 26164.6, 26300, 25754.4, 26051.7, 274830.49399999995],
    [1693008000000, 26051.7, 26129.4, 25969, 26004.3, 63925.861999999994],
    [1693094400000, 26004.3, 26173.6, 25955.6, 26087.7, 86505.398],
]
ohlcv_arr = np.array(candles)
open_arr, high_arr, low_arr = ohlcv_arr[:, ocol], ohlcv_arr[:, hcol], ohlcv_arr[:, lcol]
close_arr, vol_arr = ohlcv_arr[:, ccol], ohlcv_arr[:, vcol]
bar_idx = list(range(len(ohlcv_arr)))
open_col = pd.Series(open_arr, index=bar_idx)
high_col = pd.Series(high_arr, index=bar_idx)
low_col = pd.Series(low_arr, index=bar_idx)
close_col = pd.Series(close_arr, index=bar_idx)
vol_col = pd.Series(vol_arr, index=bar_idx)


def print_tares(ta_cres=None, ta_mres=None, mytt_res=None, pta_res=None):
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


if __name__ == '__main__':
    test_pluMinDm()
