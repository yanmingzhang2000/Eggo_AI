<script setup lang="ts">
import { computed } from 'vue'
import { useEggStore } from '@/stores/egg'
import type { NewsClue, SentimentType } from '@/types/egg'

const store = useEggStore()
const news = computed(() => store.newsClues)

/** 情感标签配置 */
const sentimentConfig: Record<SentimentType, { label: string; color: string; bg: string }> = {
  '-1': { label: '利空', color: '#ff4d4f', bg: 'rgba(255, 77, 79, 0.1)' },
  '0':  { label: '中性', color: '#666', bg: 'rgba(255, 255, 255, 0.05)' },
  '1':  { label: '利好', color: '#52c41a', bg: 'rgba(82, 196, 26, 0.1)' },
}

function getSentimentStyle(s: SentimentType) {
  return sentimentConfig[s] || sentimentConfig['0']
}

/** 重要性星级 */
function importanceStars(level: number): string {
  return '★'.repeat(level) + '☆'.repeat(5 - level)
}

/** 时间格式化 */
function formatTime(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  const h = d.getHours().toString().padStart(2, '0')
  const m = d.getMinutes().toString().padStart(2, '0')
  return `${h}:${m}`
}
</script>

<template>
  <view v-if="news.length > 0" class="section">
    <view class="section__header">
      <text class="section__title">🧠 AI 滤噪情报</text>
      <text class="section__count">{{ news.length }} 条</text>
    </view>

    <view class="news-list">
      <view
        v-for="item in news"
        :key="item.newsId"
        class="news-card"
      >
        <!-- 顶部：来源 + 时间 + 情感 -->
        <view class="news-card__header">
          <view class="news-card__source-row">
            <text class="news-card__source">{{ item.source }}</text>
            <text class="news-card__time">{{ formatTime(item.publishedAt) }}</text>
          </view>
          <view
            class="news-card__sentiment"
            :style="{
              background: getSentimentStyle(item.sentiment).bg,
              color: getSentimentStyle(item.sentiment).color,
            }"
          >
            <text class="news-card__sentiment-text">
              {{ getSentimentStyle(item.sentiment).label }}
            </text>
          </view>
        </view>

        <!-- 标题 -->
        <text class="news-card__title">{{ item.title }}</text>

        <!-- 关联性说明（核心） -->
        <view class="news-card__relevance">
          <text class="news-card__relevance-icon">🔗</text>
          <text class="news-card__relevance-text">{{ item.relevanceReason }}</text>
        </view>

        <!-- 底部：重要性 + 标签 -->
        <view class="news-card__footer">
          <text class="news-card__importance">{{ importanceStars(item.importance) }}</text>
          <view class="news-card__tags">
            <text
              v-for="tag in item.tags"
              :key="tag"
              class="news-card__tag"
            >
              #{{ tag }}
            </text>
          </view>
        </view>
      </view>
    </view>
  </view>

  <!-- 空态 -->
  <view v-else class="section">
    <view class="section__header">
      <text class="section__title">🧠 AI 滤噪情报</text>
    </view>
    <view class="empty">
      <text class="empty__emoji">📭</text>
      <text class="empty__text">今日暂无相关新闻</text>
    </view>
  </view>
</template>

<style scoped>
.section {
  margin: 32rpx 32rpx 0;
}

.section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.section__title {
  font-size: 28rpx;
  color: var(--text-secondary);
  font-weight: 600;
  letter-spacing: 1rpx;
}

.section__count {
  font-size: 22rpx;
  color: var(--text-tertiary);
}

.news-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.news-card {
  background: var(--bg-card);
  border: 1rpx solid var(--border-color);
  border-radius: 20rpx;
  padding: 28rpx;
}

.news-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16rpx;
}

.news-card__source-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.news-card__source {
  font-size: 22rpx;
  color: var(--text-tertiary);
  font-weight: 500;
}

.news-card__time {
  font-size: 20rpx;
  color: var(--text-tertiary);
  opacity: 0.6;
}

.news-card__sentiment {
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
}

.news-card__sentiment-text {
  font-size: 20rpx;
  font-weight: 600;
}

.news-card__title {
  font-size: 28rpx;
  color: var(--text-primary);
  font-weight: 600;
  line-height: 1.5;
  margin-bottom: 16rpx;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.news-card__relevance {
  display: flex;
  align-items: flex-start;
  gap: 10rpx;
  background: rgba(255, 215, 0, 0.04);
  border: 1rpx solid rgba(255, 215, 0, 0.1);
  border-radius: 12rpx;
  padding: 16rpx 20rpx;
  margin-bottom: 16rpx;
}

.news-card__relevance-icon {
  font-size: 24rpx;
  flex-shrink: 0;
  margin-top: 2rpx;
}

.news-card__relevance-text {
  font-size: 24rpx;
  color: var(--accent);
  line-height: 1.5;
  font-weight: 500;
}

.news-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.news-card__importance {
  font-size: 20rpx;
  color: var(--accent);
  letter-spacing: 2rpx;
}

.news-card__tags {
  display: flex;
  gap: 12rpx;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.news-card__tag {
  font-size: 20rpx;
  color: var(--text-tertiary);
  opacity: 0.7;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 60rpx 0;
}

.empty__emoji {
  font-size: 60rpx;
}

.empty__text {
  margin-top: 16rpx;
  font-size: 26rpx;
  color: var(--text-tertiary);
}
</style>
