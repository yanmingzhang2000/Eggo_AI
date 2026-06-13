/** 母鸡状态总响应 */
export interface EggStatusResponse {
  /** 母鸡状态概览 */
  chickenStatus: ChickenStatus
  /** 今日基金指标列表 */
  todayMetrics: FundMetric[]
  /** AI 过滤新闻线索 */
  newsClues: NewsClue[]
  /** 最终决策建议 */
  decision: Decision
  /** 数据生成时间 ISO8601 */
  generatedAt: string
}

/** 母鸡状态概览 */
export interface ChickenStatus {
  /** 状态标识：laying | resting | molting | soaking */
  overallState: ChickenState
  /** 状态 Emoji */
  stateEmoji: string
  /** 状态中文描述 */
  stateDesc: string
  /** 告警等级：none | info | warning | critical */
  alertLevel: AlertLevel
  /** 告警信息 */
  alertMessage: string
}

/** 母鸡状态枚举 */
export type ChickenState = 'laying' | 'resting' | 'molting' | 'soaking'

/** 告警等级枚举 */
export type AlertLevel = 'none' | 'info' | 'warning' | 'critical'

/** 单只基金今日指标 */
export interface FundMetric {
  /** 基金代码 */
  fundCode: string
  /** 基金名称 */
  fundName: string
  /** 单位净值 */
  unitNav: number
  /** 日涨跌幅 % */
  dailyReturn: number
  /** 连续上涨天数 */
  consecutiveUp: number
  /** 近5日累计涨幅 % */
  weekReturn: number
  /** 是否有负面新闻 */
  hasNegNews: boolean
  /** 是否有政策利好 */
  hasPosPolicy: boolean
  /** 舆情是否降温 */
  sentimentCool: boolean
}

/** AI 过滤新闻线索 */
export interface NewsClue {
  /** 新闻 ID */
  newsId: string
  /** 标题 */
  title: string
  /** 来源 */
  source: string
  /** 情感：-1 利空 | 0 中性 | 1 利好 */
  sentiment: SentimentType
  /** 重要性 1~5 */
  importance: number
  /** 关联性说明（一句话） */
  relevanceReason: string
  /** 关联基金代码 */
  relatedFunds: string[]
  /** 标签 */
  tags: string[]
  /** 发布时间 ISO8601 */
  publishedAt: string
}

/** 情感类型：-1 利空 | 0 中性 | 1 利好 */
export type SentimentType = -1 | 0 | 1

/** 最终决策建议 */
export interface Decision {
  /** 触发规则：rule1_dca | rule2_hold | rule3_take_profit | none */
  triggeredRule: TriggeredRule
  /** 规则中文名 */
  ruleName: string
  /** 操作：dca | hold | take_profit | observe */
  action: ActionType
  /** 操作 Emoji */
  actionEmoji: string
  /** 建议文案 */
  suggestion: string
  /** 置信度 0~1 */
  confidence: number
  /** 决策依据 */
  reason: string
  /** 涉及基金代码 */
  targetFunds: string[]
}

/** 触发规则类型 */
export type TriggeredRule = 'rule1_dca' | 'rule2_hold' | 'rule3_take_profit' | 'none'

/** 操作类型 */
export type ActionType = 'dca' | 'hold' | 'take_profit' | 'observe'

// ── API 通用响应包装 ────────────────────────────────────────────

/** API 统一响应结构 */
export interface ApiResponse<T> {
  code: number
  message: string
  data?: T
}
