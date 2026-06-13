-- ============================================================
-- 鸡生蛋 Eggo · MySQL 8.0 初始化建表
-- 适用: MySQL 8.0+
-- 编码: UTF-8mb4
-- ============================================================

SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- ============================================================
-- 1. 用户表 (users)
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id              CHAR(36)     PRIMARY KEY DEFAULT (UUID()),
    username        VARCHAR(64)  NOT NULL,
    email           VARCHAR(255) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    phone           VARCHAR(20),
    avatar_url      VARCHAR(512),
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '1=正常 0=禁用',
    last_login_at   DATETIME,
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_users_username (username),
    UNIQUE KEY uq_users_email (email),
    INDEX idx_users_status (status),
    INDEX idx_users_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ============================================================
-- 2. 基金基础信息表 (funds)
-- ============================================================
CREATE TABLE IF NOT EXISTS funds (
    id              CHAR(36)     PRIMARY KEY DEFAULT (UUID()),
    fund_code       VARCHAR(16)  NOT NULL COMMENT '基金代码 e.g. 110011',
    fund_name       VARCHAR(128) NOT NULL,
    fund_type       VARCHAR(32)  COMMENT '股票型/混合型/债券型/指数型/QDII',
    manager         VARCHAR(64)  COMMENT '基金经理',
    custodian       VARCHAR(64)  COMMENT '托管人',
    inception_date  DATE         COMMENT '成立日期',
    benchmark       VARCHAR(128) COMMENT '业绩基准',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '1=运作中 0=已清盘',
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_funds_code (fund_code),
    INDEX idx_funds_type (fund_type),
    FULLTEXT INDEX ft_funds_name (fund_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='基金基础信息字典表';

-- ============================================================
-- 3. 自选基金表 (watchlist)
-- ============================================================
CREATE TABLE IF NOT EXISTS watchlist (
    id              CHAR(36)     PRIMARY KEY DEFAULT (UUID()),
    user_id         CHAR(36)     NOT NULL,
    fund_id         CHAR(36)     NOT NULL,
    remark          VARCHAR(256) COMMENT '用户备注',
    sort_order      INT          NOT NULL DEFAULT 0 COMMENT '排序权重',
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uq_watchlist_user_fund (user_id, fund_id),
    INDEX idx_watchlist_user (user_id),
    INDEX idx_watchlist_fund (fund_id),
    CONSTRAINT fk_watchlist_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_watchlist_fund FOREIGN KEY (fund_id) REFERENCES funds(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户自选基金表';

-- ============================================================
-- 4. 每日净值流水表 (fund_nav_daily)
-- ============================================================
CREATE TABLE IF NOT EXISTS fund_nav_daily (
    id              CHAR(36)      PRIMARY KEY DEFAULT (UUID()),
    fund_id         CHAR(36)      NOT NULL,
    nav_date        DATE          NOT NULL COMMENT '净值日期',
    unit_nav        DECIMAL(12,4) NOT NULL COMMENT '单位净值',
    acc_nav         DECIMAL(12,4) COMMENT '累计净值',
    daily_return    DECIMAL(8,4)  COMMENT '日收益率 %',
    total_return    DECIMAL(8,4)  COMMENT '累计收益率 %',
    created_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uq_nav_fund_date (fund_id, nav_date),
    INDEX idx_nav_fund_date (fund_id, nav_date DESC),
    INDEX idx_nav_date (nav_date DESC),
    CONSTRAINT fk_nav_fund FOREIGN KEY (fund_id) REFERENCES funds(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='基金每日净值流水表';

-- ============================================================
-- 5. AI 过滤新闻表 (ai_news)
-- ============================================================
CREATE TABLE IF NOT EXISTS ai_news (
    id              CHAR(36)      PRIMARY KEY DEFAULT (UUID()),
    source          VARCHAR(64)   NOT NULL COMMENT '来源: eastmoney / sina / 10jqka',
    source_url      VARCHAR(1024),
    title           VARCHAR(512)  NOT NULL,
    summary         TEXT          COMMENT 'AI 摘要',
    content         TEXT          COMMENT '原文',
    related_funds   JSON          DEFAULT (JSON_ARRAY()) COMMENT '关联基金代码数组',
    sentiment       TINYINT       COMMENT '-1=利空 0=中性 1=利好',
    importance      TINYINT       COMMENT '1~5 重要性评分',
    tags            JSON          DEFAULT (JSON_ARRAY()) COMMENT '标签数组',
    published_at    DATETIME      COMMENT '原文发布时间',
    processed_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'AI 处理时间',
    created_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_news_published (published_at DESC),
    INDEX idx_news_sentiment (sentiment),
    INDEX idx_news_importance (importance DESC),
    INDEX idx_news_processed (processed_at DESC),
    FULLTEXT INDEX ft_news_title (title)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 过滤新闻表';

-- ============================================================
-- 6. 买卖信号决策表 (trade_signals)
-- ============================================================
CREATE TABLE IF NOT EXISTS trade_signals (
    id              CHAR(36)      PRIMARY KEY DEFAULT (UUID()),
    user_id         CHAR(36)      NOT NULL,
    fund_id         CHAR(36)      NOT NULL,
    signal_type     VARCHAR(8)    NOT NULL COMMENT 'BUY / SELL / HOLD',
    strategy        VARCHAR(64)   NOT NULL COMMENT '策略名称',
    confidence      DECIMAL(5,2)  NOT NULL COMMENT '置信度 0.00~100.00',
    reason          TEXT          COMMENT '决策理由',
    related_news_ids JSON         DEFAULT (JSON_ARRAY()) COMMENT '关联新闻 ID',
    target_amount   DECIMAL(14,2) COMMENT '建议金额',
    execute_before  DATETIME      COMMENT '建议执行截止时间',
    status          TINYINT       NOT NULL DEFAULT 0 COMMENT '0=待执行 1=已执行 2=已过期 3=用户忽略',
    executed_at     DATETIME,
    created_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_signals_user_status (user_id, status),
    INDEX idx_signals_fund (fund_id),
    INDEX idx_signals_created (created_at DESC),
    INDEX idx_signals_strategy (strategy),
    CONSTRAINT fk_signals_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_signals_fund FOREIGN KEY (fund_id) REFERENCES funds(id) ON DELETE CASCADE,
    CONSTRAINT chk_signal_type CHECK (signal_type IN ('BUY', 'SELL', 'HOLD')),
    CONSTRAINT chk_confidence CHECK (confidence >= 0 AND confidence <= 100)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='买卖信号决策表';

-- ============================================================
-- 7. 基金每日指标汇总表 (fund_daily_metrics)
-- 用于三铁律决策引擎的快速查询
-- ============================================================
CREATE TABLE IF NOT EXISTS fund_daily_metrics (
    id              CHAR(36)      PRIMARY KEY DEFAULT (UUID()),
    fund_id         CHAR(36)      NOT NULL,
    fund_code       VARCHAR(16)   NOT NULL,
    fund_name       VARCHAR(128)  NOT NULL,
    metrics_date    DATE          NOT NULL,
    unit_nav        DECIMAL(12,4) NOT NULL,
    daily_return    DECIMAL(8,4)  NOT NULL COMMENT '日涨跌幅 %',
    acc_nav         DECIMAL(12,4),
    consecutive_up  INT           DEFAULT 0 COMMENT '连续上涨天数',
    week_return     DECIMAL(8,4)  DEFAULT 0 COMMENT '近5日累计涨幅 %',
    has_neg_news    TINYINT(1)    DEFAULT 0 COMMENT '是否有负面新闻',
    has_pos_policy  TINYINT(1)    DEFAULT 0 COMMENT '是否有政策利好',
    sentiment_cool  TINYINT(1)    DEFAULT 0 COMMENT '舆情是否降温',
    created_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_metrics_fund_date (fund_code, metrics_date),
    INDEX idx_metrics_date (metrics_date DESC),
    INDEX idx_metrics_fund (fund_code),
    INDEX idx_metrics_return (daily_return),
    INDEX idx_metrics_consecutive (consecutive_up),
    CONSTRAINT fk_metrics_fund FOREIGN KEY (fund_id) REFERENCES funds(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='基金每日指标汇总表';

-- ============================================================
-- 插入测试数据
-- ============================================================
INSERT INTO funds (id, fund_code, fund_name, fund_type, manager, status) VALUES
('f0000000-0000-0000-0000-000000000001', '110011', '易方达中小盘混合', '混合型', '张坤', 1),
('f0000000-0000-0000-0000-000000000002', '161725', '招商中证白酒指数', '指数型', '侯昊', 1),
('f0000000-0000-0000-0000-000000000003', '012414', '华夏中证动漫游戏ETF', 'ETF', '徐猛', 1);

INSERT INTO users (id, username, email, password_hash, status) VALUES
('u0000000-0000-0000-0000-000000000001', 'demo', 'demo@eggo.com', '$2a$10$dummy_hash_for_demo', 1);
