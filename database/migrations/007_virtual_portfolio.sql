-- ═══════════════════════════════════════════════════════════════
-- 模拟养鸡（基）系统 — 虚拟盘表结构
-- ═══════════════════════════════════════════════════════════════

-- 虚拟账户
CREATE TABLE IF NOT EXISTS virtual_account (
  id              BIGSERIAL PRIMARY KEY,
  user_id         BIGINT NOT NULL UNIQUE,
  initial_balance NUMERIC(14,2) NOT NULL,          -- 初始资金
  cash_balance    NUMERIC(14,2) NOT NULL,          -- 可用现金
  frozen_cash     NUMERIC(14,2) DEFAULT 0,         -- 冻结资金（待结算订单）
  total_profit    NUMERIC(14,2) DEFAULT 0,         -- 累计已实现收益
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 持仓
CREATE TABLE IF NOT EXISTS virtual_position (
  id              BIGSERIAL PRIMARY KEY,
  user_id         BIGINT NOT NULL,
  fund_code       VARCHAR(10) NOT NULL,
  fund_name       VARCHAR(100) NOT NULL DEFAULT '',
  shares          NUMERIC(14,4) DEFAULT 0,         -- 持有份额
  avg_cost        NUMERIC(14,8) DEFAULT 0,         -- 持仓均价
  total_cost      NUMERIC(14,2) DEFAULT 0,         -- 持仓总成本
  dividend_method VARCHAR(10) DEFAULT 'reinvest',  -- reinvest=红利再投 / cash=现金分红
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  updated_at      TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE (user_id, fund_code)
);

-- 订单（买/卖申请，T日提交，T+1确认）
CREATE TABLE IF NOT EXISTS virtual_order (
  id              BIGSERIAL PRIMARY KEY,
  user_id         BIGINT NOT NULL,
  fund_code       VARCHAR(10) NOT NULL,
  fund_name       VARCHAR(100) NOT NULL DEFAULT '',
  order_type      VARCHAR(4) NOT NULL,             -- buy / sell
  amount          NUMERIC(14,2) NOT NULL,          -- 买入金额 / 卖出份额
  status          VARCHAR(12) DEFAULT 'pending',   -- pending / confirmed / cancelled
  order_date      DATE NOT NULL,                   -- T日
  settle_nav      NUMERIC(14,8),                   -- 结算净值（T日晚公布）
  settle_shares   NUMERIC(14,4),                   -- 确认份额
  settle_amount   NUMERIC(14,2),                   -- 确认金额（扣费后）
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 交易流水
CREATE TABLE IF NOT EXISTS virtual_transaction (
  id              BIGSERIAL PRIMARY KEY,
  user_id         BIGINT NOT NULL,
  fund_code       VARCHAR(10) NOT NULL,
  tx_type         VARCHAR(12) NOT NULL,            -- buy / sell / dividend / dca_buy
  amount          NUMERIC(14,2),                   -- 金额
  shares          NUMERIC(14,4),                   -- 份额
  nav             NUMERIC(14,8),                   -- 成交净值
  created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 定投计划
CREATE TABLE IF NOT EXISTS virtual_dca_plan (
  id              BIGSERIAL PRIMARY KEY,
  user_id         BIGINT NOT NULL,
  fund_code       VARCHAR(10) NOT NULL,
  fund_name       VARCHAR(100) NOT NULL DEFAULT '',
  amount          NUMERIC(14,2) NOT NULL,          -- 每次定投金额
  frequency       VARCHAR(10) NOT NULL,            -- weekly / biweekly / monthly
  exec_day        INT NOT NULL,                    -- 周几(1-7) 或 几号(1-31)
  next_exec_date  DATE NOT NULL,                   -- 下次执行日
  status          VARCHAR(10) DEFAULT 'active',    -- active / paused
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_virtual_order_user_status ON virtual_order(user_id, status);
CREATE INDEX IF NOT EXISTS idx_virtual_order_date ON virtual_order(order_date);
CREATE INDEX IF NOT EXISTS idx_virtual_position_user ON virtual_position(user_id);
CREATE INDEX IF NOT EXISTS idx_virtual_transaction_user ON virtual_transaction(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_virtual_dca_next ON virtual_dca_plan(status, next_exec_date);
