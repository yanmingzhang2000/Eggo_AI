-- ============================================================
-- 「鸡生蛋」基金 Agent 平台 - PostgreSQL 初始化建表
-- 适用: PostgreSQL 15+
-- 编码: UTF-8
-- ============================================================

-- 启用 UUID 扩展（分布式友好主键）
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- 模糊搜索优化

-- ============================================================
-- 1. 用户表 (users)
-- ============================================================
CREATE TABLE users (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    username        VARCHAR(64) NOT NULL,
    email           VARCHAR(255) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    phone           VARCHAR(20),
    avatar_url      VARCHAR(512),
    status          SMALLINT    NOT NULL DEFAULT 1,  -- 1=正常 0=禁用
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_users_username UNIQUE (username),
    CONSTRAINT uq_users_email    UNIQUE (email)
);

CREATE INDEX idx_users_status      ON users (status);
CREATE INDEX idx_users_phone       ON users (phone) WHERE phone IS NOT NULL;

COMMENT ON TABLE  users             IS '用户表';
COMMENT ON COLUMN users.status      IS '1=正常 0=禁用';

-- ============================================================
-- 2. 基金基础信息表 (funds) — 公共字典表，非自选
-- ============================================================
CREATE TABLE funds (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    fund_code       VARCHAR(16) NOT NULL,            -- 基金代码 e.g. "110011"
    fund_name       VARCHAR(128) NOT NULL,
    fund_type       VARCHAR(32),                     -- 股票型/混合型/债券型/指数型/QDII...
    manager         VARCHAR(64),                     -- 基金经理
    custodian       VARCHAR(64),                     -- 托管人
    inception_date  DATE,                            -- 成立日期
    benchmark       VARCHAR(128),                    -- 业绩基准
    status          SMALLINT    NOT NULL DEFAULT 1,  -- 1=运作中 0=已清盘
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_funds_code UNIQUE (fund_code)
);

CREATE INDEX idx_funds_type   ON funds (fund_type);
CREATE INDEX idx_funds_name   ON funds USING gin (fund_name gin_trgm_ops);  -- 模糊搜索

COMMENT ON TABLE funds IS '基金基础信息字典表';

-- ============================================================
-- 3. 自选基金表 (watchlist) — 用户与基金的多对多关系
-- ============================================================
CREATE TABLE watchlist (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fund_id         UUID        NOT NULL REFERENCES funds(id) ON DELETE CASCADE,
    remark          VARCHAR(256),                    -- 用户备注
    sort_order      INT         NOT NULL DEFAULT 0,  -- 排序权重
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_watchlist_user_fund UNIQUE (user_id, fund_id)  -- 同一用户不能重复自选
);

CREATE INDEX idx_watchlist_user ON watchlist (user_id);
CREATE INDEX idx_watchlist_fund ON watchlist (fund_id);

COMMENT ON TABLE watchlist IS '用户自选基金表';

-- ============================================================
-- 4. 每日净值流水表 (fund_nav_daily)
-- ============================================================
CREATE TABLE fund_nav_daily (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    fund_id         UUID        NOT NULL REFERENCES funds(id) ON DELETE CASCADE,
    nav_date        DATE        NOT NULL,            -- 净值日期
    unit_nav        NUMERIC(12,4) NOT NULL,          -- 单位净值
    acc_nav         NUMERIC(12,4),                   -- 累计净值
    daily_return    NUMERIC(8,4),                    -- 日收益率 %
    total_return    NUMERIC(8,4),                    -- 累计收益率 %
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_nav_fund_date UNIQUE (fund_id, nav_date)  -- 同一基金同日唯一
);

-- 按日期范围查询净值的主力索引
CREATE INDEX idx_nav_fund_date ON fund_nav_daily (fund_id, nav_date DESC);
-- 按日期跨基金查询
CREATE INDEX idx_nav_date      ON fund_nav_daily (nav_date DESC);

COMMENT ON TABLE  fund_nav_daily          IS '基金每日净值流水表';
COMMENT ON COLUMN fund_nav_daily.unit_nav IS '单位净值';
COMMENT ON COLUMN fund_nav_daily.acc_nav  IS '累计净值';

-- ============================================================
-- 5. AI 过滤新闻表 (ai_news)
-- ============================================================
CREATE TABLE ai_news (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    source          VARCHAR(64)  NOT NULL,           -- 来源: eastmoney / sina / 10jqka ...
    source_url      VARCHAR(1024),
    title           VARCHAR(512) NOT NULL,
    summary         TEXT,                            -- AI 摘要
    content         TEXT,                            -- 原文
    related_funds   UUID[]       DEFAULT '{}',       -- 关联基金 ID 数组
    sentiment       SMALLINT,                        -- -1=利空 0=中性 1=利好
    importance      SMALLINT,                        -- 1~5 重要性评分
    tags            VARCHAR(64)[] DEFAULT '{}',      -- 标签: ["政策","利率","科技"]
    published_at    TIMESTAMPTZ,                     -- 原文发布时间
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- AI 处理时间
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_news_published  ON ai_news (published_at DESC);
CREATE INDEX idx_news_sentiment  ON ai_news (sentiment);
CREATE INDEX idx_news_importance ON ai_news (importance DESC);
CREATE INDEX idx_news_tags       ON ai_news USING gin (tags);
CREATE INDEX idx_news_related    ON ai_news USING gin (related_funds);  -- 数组字段查询
CREATE INDEX idx_news_title_trgm ON ai_news USING gin (title gin_trgm_ops);

COMMENT ON TABLE  ai_news            IS 'AI 过滤新闻表';
COMMENT ON COLUMN ai_news.sentiment  IS '-1=利空 0=中性 1=利好';
COMMENT ON COLUMN ai_news.importance IS '1~5 重要性评分';

-- ============================================================
-- 6. 买卖信号决策表 (trade_signals)
-- ============================================================
CREATE TABLE trade_signals (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fund_id         UUID        NOT NULL REFERENCES funds(id) ON DELETE CASCADE,
    signal_type     VARCHAR(8)  NOT NULL,            -- BUY / SELL / HOLD
    strategy        VARCHAR(64) NOT NULL,            -- 策略名称: "均线突破" / "AI新闻驱动" / "定投补仓"
    confidence      NUMERIC(5,2) NOT NULL,           -- 置信度 0.00~100.00
    reason          TEXT,                            -- 决策理由（可含 AI 推理链）
    related_news_ids UUID[]     DEFAULT '{}',        -- 关联新闻 ID
    target_amount   NUMERIC(14,2),                   -- 建议金额
    execute_before  TIMESTAMPTZ,                     -- 建议执行截止时间
    status          SMALLINT    NOT NULL DEFAULT 0,  -- 0=待执行 1=已执行 2=已过期 3=用户忽略
    executed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_signal_type CHECK (signal_type IN ('BUY', 'SELL', 'HOLD')),
    CONSTRAINT chk_confidence  CHECK (confidence >= 0 AND confidence <= 100)
);

CREATE INDEX idx_signals_user_status ON trade_signals (user_id, status);
CREATE INDEX idx_signals_fund        ON trade_signals (fund_id);
CREATE INDEX idx_signals_created     ON trade_signals (created_at DESC);
CREATE INDEX idx_signals_strategy    ON trade_signals (strategy);

COMMENT ON TABLE  trade_signals           IS '买卖信号决策表';
COMMENT ON COLUMN trade_signals.signal_type IS 'BUY=买入 SELL=卖出 HOLD=持有观望';
COMMENT ON COLUMN trade_signals.status   IS '0=待执行 1=已执行 2=已过期 3=用户忽略';

-- ============================================================
-- 通用 updated_at 自动更新触发器
-- ============================================================
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_funds_updated_at
    BEFORE UPDATE ON funds
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_trade_signals_updated_at
    BEFORE UPDATE ON trade_signals
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
