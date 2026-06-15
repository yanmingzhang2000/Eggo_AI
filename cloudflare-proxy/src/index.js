/**
 * Eggo 国内金融 API 代理
 * Cloudflare Worker — 10万次/天免费
 *
 * 功能：
 *   1. 转发请求到国内金融 API（天天基金、东方财富等）
 *   2. 绕过境外 IP 封锁
 *   3. KV 缓存层（可选，降低源站压力）
 *   4. API Key 鉴权
 *   5. 域名白名单防滥用
 *   6. 限流保护
 */

// ─── 配置 ────────────────────────────────────────────────────────────────

// 允许转发的域名（防滥用）
const ALLOWED_HOSTS = new Set([
  'fundgz.1234567.com.cn',       // 天天基金实时估值
  'api.fund.eastmoney.com',      // 东方财富基金
  'fund.eastmoney.com',          // 东方财富基金
  'push2.eastmoney.com',         // 东方财富行情推送
  'push2his.eastmoney.com',      // 东方财富历史/K线数据
  'datacenter-web.eastmoney.com',// 东方财富数据中心
  'data.eastmoney.com',          // 东方财富数据
  'hq.sinajs.cn',                // 新浪财经行情
  'vip.stock.finance.sina.com.cn',// 新浪股票
  'money.finance.sina.com.cn',   // 新浪基金
  'api.doctorxiong.club',        // 第三方基金
  'api.yfin.dev',                // Yahoo Finance 代理（备用）
])

// 缓存 TTL（秒）— 不同接口不同策略
const CACHE_TTL = {
  'fundgz':       180,   // 天天基金估值：3分钟
  'rankhandler':  300,   // 排行榜：5分钟
  'hq':           30,    // 实时行情：30秒
  'index':        60,    // 指数：1分钟
  'default':      60,    // 默认：1分钟
}

// 限流：每 IP 每分钟最多 N 次请求
const RATE_LIMIT = 120

// ─── 主入口 ──────────────────────────────────────────────────────────────

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url)

    // OPTIONS 预检请求
    if (request.method === 'OPTIONS') {
      return handleCORS(env)
    }

    // 健康检查
    if (url.pathname === '/health') {
      return json({ status: 'ok', time: new Date().toISOString() })
    }

    // 路由分发
    try {
      switch (url.pathname) {
        case '/proxy':
          return handleProxy(request, env, ctx)
        case '/batch':
          return handleBatch(request, env, ctx)
        case '/stats':
          return handleStats(env)
        default:
          return json({ error: 'not found' }, 404)
      }
    } catch (err) {
      return json({ error: err.message }, 500)
    }
  }
}

// ─── 单个代理 ────────────────────────────────────────────────────────────

async function handleProxy(request, env, ctx) {
  // 鉴权
  const authError = checkAuth(request, env)
  if (authError) return authError

  const url = new URL(request.url)
  const targetURL = url.searchParams.get('url')
  if (!targetURL) {
    return json({ error: 'missing ?url= parameter' }, 400)
  }

  // 限流
  const rateError = await checkRateLimit(request, env)
  if (rateError) return rateError

  // 解析目标 URL 并校验域名
  let target
  try {
    target = new URL(targetURL)
  } catch {
    return json({ error: 'invalid url' }, 400)
  }

  if (!ALLOWED_HOSTS.has(target.hostname)) {
    return json({
      error: `host "${target.hostname}" not allowed`,
      allowed: [...ALLOWED_HOSTS]
    }, 403)
  }

  // 检查 KV 缓存
  const cacheKey = `proxy:${targetURL}`
  const ttl = getCacheTTL(targetURL)

  if (env.CACHE) {
    const cached = await env.CACHE.get(cacheKey, { type: 'text' })
    if (cached) {
      return new Response(cached, {
        headers: buildHeaders(env, {
          'X-Cache': 'HIT',
          'Content-Type': getContentType(targetURL),
        })
      })
    }
  }

  // 转发请求
  const proxyReq = new Request(targetURL, {
    method: request.method,
    headers: {
      'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
      'Referer': getReferer(target.hostname),
      'Accept': '*/*',
    },
  })

  const resp = await fetch(proxyReq)
  const body = await resp.text()

  // 写入缓存
  if (env.CACHE && resp.ok) {
    ctx.waitUntil(env.CACHE.put(cacheKey, body, { expirationTtl: ttl }))
  }

  return new Response(body, {
    status: resp.status,
    headers: buildHeaders(env, {
      'X-Cache': 'MISS',
      'Content-Type': getContentType(targetURL),
    })
  })
}

// ─── 批量代理（一次请求多个 URL）────────────────────────────────────────

async function handleBatch(request, env, ctx) {
  const authError = checkAuth(request, env)
  if (authError) return authError

  const { urls } = await request.json()
  if (!Array.isArray(urls) || urls.length === 0) {
    return json({ error: 'missing urls array' }, 400)
  }

  if (urls.length > 20) {
    return json({ error: 'max 20 urls per batch' }, 400)
  }

  // 逐个转发，带缓存
  const results = await Promise.allSettled(
    urls.map(async (targetURL) => {
      try {
        const target = new URL(targetURL)
        if (!ALLOWED_HOSTS.has(target.hostname)) {
          return { url: targetURL, error: 'host not allowed' }
        }

        // 检查缓存
        const cacheKey = `proxy:${targetURL}`
        if (env.CACHE) {
          const cached = await env.CACHE.get(cacheKey, { type: 'text' })
          if (cached) {
            return { url: targetURL, data: cached, cached: true }
          }
        }

        const proxyReq = new Request(targetURL, {
          headers: {
            'User-Agent': 'Mozilla/5.0',
            'Referer': getReferer(target.hostname),
          }
        })
        const resp = await fetch(proxyReq)
        const body = await resp.text()

        // 写缓存
        const ttl = getCacheTTL(targetURL)
        if (env.CACHE && resp.ok) {
          ctx.waitUntil(env.CACHE.put(cacheKey, body, { expirationTtl: ttl }))
        }

        return { url: targetURL, data: body }
      } catch (e) {
        return { url: targetURL, error: e.message }
      }
    })
  )

  return json({
    results: results.map(r => r.value || { error: r.reason?.message }),
    time: new Date().toISOString()
  })
}

// ─── 统计信息 ────────────────────────────────────────────────────────────

async function handleStats(env) {
  let kvStats = null
  if (env.CACHE) {
    try {
      const list = await env.CACHE.list()
      kvStats = { keys: list.keys.length }
    } catch {
      kvStats = { error: 'kv not available' }
    }
  }
  return json({
    allowedHosts: [...ALLOWED_HOSTS],
    cacheTTL: CACHE_TTL,
    rateLimit: RATE_LIMIT,
    kv: kvStats,
  })
}

// ─── 工具函数 ────────────────────────────────────────────────────────────

function checkAuth(request, env) {
  const apiKey = env.API_KEY
  if (!apiKey) return null // 未配置则跳过鉴权

  const authHeader = request.headers.get('Authorization')
  const url = new URL(request.url)
  const tokenParam = url.searchParams.get('key')

  if (authHeader === `Bearer ${apiKey}` || tokenParam === apiKey) {
    return null
  }

  return json({ error: 'unauthorized', hint: 'set Authorization: Bearer <key>' }, 401)
}

// 简单的内存限流（每个 Worker 实例独立）
const rateLimitMap = new Map()

async function checkRateLimit(request, env) {
  const ip = request.headers.get('CF-Connecting-IP') || 'unknown'
  const now = Date.now()
  const windowMs = 60 * 1000

  const record = rateLimitMap.get(ip)
  if (!record || now - record.start > windowMs) {
    rateLimitMap.set(ip, { start: now, count: 1 })
    return null
  }

  record.count++
  if (record.count > RATE_LIMIT) {
    return json({
      error: 'rate limit exceeded',
      retryAfter: Math.ceil((record.start + windowMs - now) / 1000),
    }, 429)
  }

  return null
}

function getCacheTTL(url) {
  if (url.includes('fundgz')) return CACHE_TTL.fundgz
  if (url.includes('rankhandler')) return CACHE_TTL.rankhandler
  if (url.includes('hq.sinajs')) return CACHE_TTL.hq
  if (url.includes('push2')) return CACHE_TTL.index
  return CACHE_TTL.default
}

function getReferer(hostname) {
  if (hostname.includes('eastmoney')) return 'https://fund.eastmoney.com/'
  if (hostname.includes('1234567')) return 'https://fund.eastmoney.com/'
  if (hostname.includes('sina')) return 'https://finance.sina.com.cn/'
  return 'https://www.google.com'
}

function getContentType(url) {
  if (url.includes('.js')) return 'application/javascript'
  if (url.includes('.json')) return 'application/json'
  return 'text/plain'
}

function buildHeaders(env, extra = {}) {
  return {
    'Access-Control-Allow-Origin': env.ALLOWED_ORIGINS?.split(',')[0] || '*',
    'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
    'Access-Control-Allow-Headers': 'Content-Type, Authorization',
    'Access-Control-Max-Age': '86400',
    ...extra,
  }
}

function handleCORS(env) {
  return new Response(null, {
    status: 204,
    headers: buildHeaders(env),
  })
}

function json(data, status = 200) {
  return new Response(JSON.stringify(data), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}
