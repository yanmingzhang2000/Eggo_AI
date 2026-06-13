from datetime import datetime

from pydantic import BaseModel, Field


# ── 请求模型 ───────────────────────────────────────────────────
class RawNewsItem(BaseModel):
    """爬虫抓取的原始新闻条目"""
    id: str = Field(..., description="新闻唯一标识")
    source: str = Field(..., description="来源：财联社 / 新浪财经 / 36氪 等")
    title: str = Field(..., description="新闻标题")
    summary: str = Field(default="", description="新闻摘要/正文片段")
    url: str = Field(default="", description="原文链接")
    published_at: str = Field(default="", description="发布时间 ISO 格式")


class FundInfo(BaseModel):
    """用户持仓基金信息"""
    fund_code: str = Field(..., description="基金代码")
    fund_name: str = Field(..., description="基金名称")
    fund_type: str = Field(default="", description="基金类型：股票型/混合型/债券型/指数型")
    top_holdings: list[str] = Field(default_factory=list, description="重仓股票/行业")
    top_sectors: list[str] = Field(default_factory=list, description="重仓行业板块")


class AIProcessRequest(BaseModel):
    """AI 过滤处理请求"""
    news_list: list[RawNewsItem] = Field(..., min_length=1, max_length=200, description="原始新闻列表")
    user_funds: list[FundInfo] = Field(..., min_length=1, max_length=50, description="用户持仓基金列表")


# ── 响应模型 ───────────────────────────────────────────────────
class FilteredNewsItem(BaseModel):
    """AI 过滤后的新闻条目"""
    original_id: str = Field(..., description="原始新闻 ID")
    title: str = Field(..., description="新闻标题")
    source: str = Field(..., description="来源")
    url: str = Field(default="", description="原文链接")
    published_at: str = Field(default="", description="发布时间")
    sentiment: int = Field(..., ge=-1, le=1, description="情感倾向：-1=利空 0=中性 1=利好")
    importance: int = Field(..., ge=1, le=5, description="重要性评分 1~5")
    relevance_reason: str = Field(..., description="一句话说明：为什么这条新闻和你的基金有关")
    related_funds: list[str] = Field(default_factory=list, description="关联的基金代码列表")
    tags: list[str] = Field(default_factory=list, description="标签：政策/利率/行业/个股 等")


class AIProcessResponse(BaseModel):
    """AI 过滤处理响应"""
    total_input: int = Field(..., description="输入新闻总数")
    total_filtered: int = Field(..., description="过滤后保留的新闻数")
    filtered_news: list[FilteredNewsItem] = Field(default_factory=list, description="过滤后的新闻列表")
    market_sentiment: str = Field(default="", description="整体市场情绪概要")
    processing_time_ms: float = Field(..., description="处理耗时（毫秒）")
    model_used: str = Field(..., description="使用的模型名称")
