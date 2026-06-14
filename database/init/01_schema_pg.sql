-- ============================================================
-- 鸡生蛋 Eggo · PostgreSQL 初始化建表
-- ============================================================

-- 1. 用户表
CREATE TABLE IF NOT EXISTS users (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(64)  NOT NULL,
    email           VARCHAR(255) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    phone           VARCHAR(20),
    avatar_url      VARCHAR(512),
    status          SMALLINT     NOT NULL DEFAULT 1,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    UNIQUE (username),
    UNIQUE (email)
);

CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_phone ON users(phone);

-- 2. 基金基础信息表
CREATE TABLE IF NOT EXISTS funds (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    fund_code       VARCHAR(16)  NOT NULL,
    fund_name       VARCHAR(128) NOT NULL,
    fund_type       VARCHAR(32),
    manager         VARCHAR(64),
    custodian       VARCHAR(64),
    inception_date  DATE,
    benchmark       VARCHAR(128),
    status          SMALLINT     NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    UNIQUE (fund_code)
);

CREATE INDEX idx_funds_type ON funds(fund_type);

-- 3. 自选基金表
CREATE TABLE IF NOT EXISTS watchlist (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID         NOT NULL,
    fund_id         UUID         NOT NULL,
    remark          VARCHAR(256),
    sort_order      INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, fund_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (fund_id) REFERENCES funds(id) ON DELETE CASCADE
);

-- 4. 每日净值流水表
CREATE TABLE IF NOT EXISTS fund_nav_daily (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    fund_id         UUID           NOT NULL,
    nav_date        DATE           NOT NULL,
    unit_nav        DECIMAL(12,4)  NOT NULL,
    acc_nav         DECIMAL(12,4),
    daily_return    DECIMAL(8,4),
    total_return    DECIMAL(8,4),
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    UNIQUE (fund_id, nav_date),
    FOREIGN KEY (fund_id) REFERENCES funds(id) ON DELETE CASCADE
);

CREATE INDEX idx_nav_fund_date ON fund_nav_daily(fund_id, nav_date DESC);
CREATE INDEX idx_nav_date ON fund_nav_daily(nav_date DESC);

-- 5. AI 过滤新闻表
CREATE TABLE IF NOT EXISTS ai_news (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    source          VARCHAR(64)    NOT NULL,
    source_url      VARCHAR(1024),
    title           VARCHAR(512)   NOT NULL,
    summary         TEXT,
    content         TEXT,
    related_funds   JSONB          DEFAULT '[]'::jsonb,
    sentiment       SMALLINT,
    importance      SMALLINT,
    tags            JSONB          DEFAULT '[]'::jsonb,
    published_at    TIMESTAMPTZ,
    processed_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_news_published ON ai_news(published_at DESC);
CREATE INDEX idx_news_sentiment ON ai_news(sentiment);
CREATE INDEX idx_news_importance ON ai_news(importance DESC);

-- 6. 买卖信号决策表
CREATE TABLE IF NOT EXISTS trade_signals (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID           NOT NULL,
    fund_id         UUID           NOT NULL,
    signal_type     VARCHAR(8)     NOT NULL,
    strategy        VARCHAR(64)    NOT NULL,
    confidence      DECIMAL(5,2)   NOT NULL,
    reason          TEXT,
    related_news_ids JSONB         DEFAULT '[]'::jsonb,
    target_amount   DECIMAL(14,2),
    execute_before  TIMESTAMPTZ,
    status          SMALLINT       NOT NULL DEFAULT 0,
    executed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (fund_id) REFERENCES funds(id) ON DELETE CASCADE,
    CHECK (signal_type IN ('BUY', 'SELL', 'HOLD')),
    CHECK (confidence >= 0 AND confidence <= 100)
);

-- 7. 基金每日指标汇总表
CREATE TABLE IF NOT EXISTS fund_daily_metrics (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    fund_id         UUID           NOT NULL,
    fund_code       VARCHAR(16)    NOT NULL,
    fund_name       VARCHAR(128)   NOT NULL,
    metrics_date    DATE           NOT NULL,
    unit_nav        DECIMAL(12,4)  NOT NULL,
    daily_return    DECIMAL(8,4)   NOT NULL,
    acc_nav         DECIMAL(12,4),
    consecutive_up  INT            DEFAULT 0,
    week_return     DECIMAL(8,4)   DEFAULT 0,
    has_neg_news    SMALLINT       DEFAULT 0,
    has_pos_policy  SMALLINT       DEFAULT 0,
    sentiment_cool  SMALLINT       DEFAULT 0,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    UNIQUE (fund_code, metrics_date),
    FOREIGN KEY (fund_id) REFERENCES funds(id) ON DELETE CASCADE
);

-- 插入测试数据
INSERT INTO funds (id, fund_code, fund_name, fund_type, manager, status) VALUES
('f0000000-0000-0000-0000-000000000001', '110011', '易方达中小盘混合', '混合型', '张坤', 1),
('f0000000-0000-0000-0000-000000000002', '161725', '招商中证白酒指数', '指数型', '侯昊', 1),
('f0000000-0000-0000-0000-000000000003', '012414', '华夏中证动漫游戏ETF', 'ETF', '徐猛', 1);

INSERT INTO users (id, username, email, password_hash, status) VALUES
('u0000000-0000-0000-0000-000000000001', 'demo', 'demo@eggo.com', '$2a$10$dummy_hash_for_demo', 1);
