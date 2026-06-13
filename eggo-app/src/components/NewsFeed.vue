<script setup lang="ts">
import { computed } from 'vue'
import { useEggStore } from '@/stores/egg'
import type { SentimentType } from '@/types/egg'

const store = useEggStore()
const news = computed(() => store.newsClues)

const sentimentConfig: Record<SentimentType, { label: string; color: string; bg: string }> = {
  '-1': { label: '利空', color: '#ff4d4f', bg: 'rgba(255, 77, 79, 0.1)' },
  '0':  { label: '中性', color: '#666', bg: 'rgba(255, 255, 255, 0.05)' },
  '1':  { label: '利好', color: '#52c41a', bg: 'rgba(82, 196, 26, 0.1)' },
}

function getSentimentStyle(s: SentimentType) {
  return sentimentConfig[s] || sentimentConfig['0']
}

function importanceStars(level: number): string {
  return '★'.repeat(level) + '☆'.repeat(5 - level)
}

function formatTime(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}
</script>

<template>
  <section class="section">
    <div class="section__header">
      <h2 class="section__title">🧠 AI 滤噪情报</h2>
      <span class="section__count">{{ news.length }} 条</span>
    </div>

    <div v-if="news.length > 0" class="news-list">
      <div v-for="item in news" :key="item.newsId" class="news-card">
        <!-- 顶部 -->
        <div class="news-card__header">
          <div class="news-card__source-row">
            <span class="news-card__source">{{ item.source }}</span>
            <span class="news-card__time">{{ formatTime(item.publishedAt) }}</span>
          </div>
          <span
            class="news-card__sentiment"
            :style="{
              background: getSentimentStyle(item.sentiment).bg,
              color: getSentimentStyle(item.sentiment).color,
            }"
          >
            {{ getSentimentStyle(item.sentiment).label }}
          </span>
        </div>

        <!-- 标题 -->
        <h3 class="news-card__title">{{ item.title }}</h3>

        <!-- 关联性说明 -->
        <div class="news-card__relevance">
          <span class="news-card__relevance-icon">🔗</span>
          <p class="news-card__relevance-text">{{ item.relevanceReason }}</p>
        </div>

        <!-- 底部 -->
        <div class="news-card__footer">
          <span class="news-card__importance">{{ importanceStars(item.importance) }}</span>
          <div class="news-card__tags">
            <span v-for="tag in item.tags" :key="tag" class="news-card__tag">#{{ tag }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="empty">
      <p class="empty__emoji">📭</p>
      <p class="empty__text">今日暂无相关新闻</p>
    </div>
  </section>
</template>

<style scoped>
.section {
  margin-bottom: 32px;
}

.section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.section__title {
  font-size: 16px;
  color: var(--text-secondary);
  font-weight: 600;
  margin: 0;
}

.section__count {
  font-size: 12px;
  color: var(--text-tertiary);
}

.news-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.news-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 20px;
  transition: border-color 0.2s;
}

.news-card:hover {
  border-color: rgba(255, 215, 0, 0.2);
}

.news-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.news-card__source-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.news-card__source {
  font-size: 12px;
  color: var(--text-tertiary);
  font-weight: 500;
}

.news-card__time {
  font-size: 11px;
  color: var(--text-tertiary);
  opacity: 0.6;
}

.news-card__sentiment {
  padding: 3px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
}

.news-card__title {
  font-size: 15px;
  color: var(--text-primary);
  font-weight: 600;
  line-height: 1.5;
  margin: 0 0 12px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.news-card__relevance {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  background: rgba(255, 215, 0, 0.04);
  border: 1px solid rgba(255, 215, 0, 0.1);
  border-radius: 10px;
  padding: 12px 16px;
  margin-bottom: 12px;
}

.news-card__relevance-icon {
  font-size: 14px;
  flex-shrink: 0;
  margin-top: 2px;
}

.news-card__relevance-text {
  font-size: 13px;
  color: var(--accent);
  line-height: 1.5;
  font-weight: 500;
  margin: 0;
}

.news-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.news-card__importance {
  font-size: 11px;
  color: var(--accent);
  letter-spacing: 2px;
}

.news-card__tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.news-card__tag {
  font-size: 11px;
  color: var(--text-tertiary);
  opacity: 0.7;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px 0;
}

.empty__emoji {
  font-size: 40px;
  margin: 0;
}

.empty__text {
  margin-top: 12px;
  font-size: 14px;
  color: var(--text-tertiary);
}
</style>
