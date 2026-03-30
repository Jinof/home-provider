<template>
  <div class="container">
    <div class="tab-nav">
      <button class="tab-btn" :class="{ active: tab === 'guide' }" @click="switchTab('guide')">
        {{ $t('tabs.guide') }}
      </button>
      <button class="tab-btn" :class="{ active: tab === 'tags' }" @click="switchTab('tags')">
        {{ $t('tabs.tags') }}
      </button>
      <button
        class="tab-btn"
        :class="{ active: tab === 'providers' }"
        @click="switchTab('providers')"
      >
        {{ $t('tabs.providers') }}
      </button>
      <button class="tab-btn" :class="{ active: tab === 'keys' }" @click="switchTab('keys')">
        {{ $t('tabs.keys') }}
      </button>
      <button class="tab-btn" :class="{ active: tab === 'stats' }" @click="switchTab('stats')">
        {{ $t('tabs.stats') }}
      </button>
      <button class="tab-btn" :class="{ active: tab === 'logs' }" @click="switchTab('logs')">
        {{ $t('logs.title') }}
      </button>
      <select v-model="locale" @change="saveLocale" class="lang-switch">
        <option value="en">{{ $t('lang.en') }}</option>
        <option value="zh">{{ $t('lang.zh') }}</option>
      </select>
    </div>

    <!-- Usage Guide -->
    <div v-show="tab === 'guide'" class="card">
      <h2>{{ $t('guide.api_usage') }}</h2>
      <p style="color: #666; margin-bottom: 20px">{{ $t('guide.intro') }}</p>

      <!-- Static Architecture Diagram -->
      <div class="diagram-section" style="margin-bottom: 25px">
        <h3 style="margin-bottom: 15px">{{ $t('stats.diagram_title') }}</h3>
        <div class="diagram-container">
          <svg viewBox="0 0 800 130" class="flow-diagram">
            <defs>
              <marker
                id="arrow"
                markerWidth="10"
                markerHeight="7"
                refX="9"
                refY="3.5"
                orient="auto"
              >
                <polygon points="0 0, 10 3.5, 0 7" fill="#4a4a6a" />
              </marker>
            </defs>

            <!-- Client Box -->
            <g transform="translate(50, 30)">
              <rect width="90" height="70" rx="8" class="box-fill-client" />
              <text x="45" y="28" text-anchor="middle" class="box-title-client">Client</text>
              <text x="45" y="44" text-anchor="middle" class="box-subtitle-client">
                Claude Code
              </text>
              <text x="45" y="58" text-anchor="middle" class="box-subtitle-client">Your App</text>
            </g>

            <!-- Arrow to API Key -->
            <line x1="140" y1="65" x2="185" y2="65" class="static-arrow" marker-end="url(#arrow)" />
            <text x="162" y="58" text-anchor="middle" class="flow-label">Key</text>

            <!-- API Key Box -->
            <g transform="translate(185, 35)">
              <rect width="90" height="60" rx="8" class="box-fill-static" />
              <text x="45" y="22" text-anchor="middle" class="box-title">API Key</text>
              <text x="45" y="42" text-anchor="middle" class="box-subtitle-sm">sk-xxx</text>
            </g>

            <!-- Arrow to Tag -->
            <line x1="275" y1="65" x2="320" y2="65" class="static-arrow" marker-end="url(#arrow)" />
            <text x="297" y="58" text-anchor="middle" class="flow-label">Bearer</text>

            <!-- Tag Box -->
            <g transform="translate(320, 35)">
              <rect width="90" height="60" rx="8" class="box-fill-static" />
              <text x="45" y="25" text-anchor="middle" class="box-title">Tag</text>
              <text x="45" y="45" text-anchor="middle" class="box-subtitle-sm">latest</text>
            </g>

            <!-- Arrow to Provider -->
            <line x1="410" y1="65" x2="455" y2="65" class="static-arrow" marker-end="url(#arrow)" />
            <text x="432" y="58" text-anchor="middle" class="flow-label">model</text>

            <!-- Provider Box -->
            <g transform="translate(455, 35)">
              <rect width="110" height="60" rx="8" class="box-fill-static" />
              <text x="55" y="20" text-anchor="middle" class="box-title">Provider</text>
              <text x="55" y="38" text-anchor="middle" class="box-subtitle-sm">MiniMax</text>
              <text x="55" y="52" text-anchor="middle" class="box-subtitle-sm">Kimi</text>
            </g>

            <!-- Arrow to AI -->
            <line x1="565" y1="65" x2="610" y2="65" class="static-arrow" marker-end="url(#arrow)" />
            <text x="587" y="58" text-anchor="middle" class="flow-label">AI</text>

            <!-- AI Output -->
            <g transform="translate(610, 35)">
              <rect width="70" height="60" rx="8" class="box-fill-ai" />
              <text x="35" y="35" text-anchor="middle" class="box-title-ai">Response</text>
            </g>
          </svg>
        </div>
        <div class="diagram-legend">
          <span class="legend-item">{{ $t('stats.diagram_step1') }}</span>
          <span class="legend-item">{{ $t('stats.diagram_step2') }}</span>
          <span class="legend-item">{{ $t('stats.diagram_step3') }}</span>
        </div>
      </div>

      <!-- Curl Example -->
      <h3 style="font-size: 1rem; margin-bottom: 10px">cURL Example</h3>
      <pre
        style="
          background: #2d2d2d;
          color: #f8f8f2;
          padding: 15px;
          border-radius: 4px;
          overflow-x: auto;
          font-size: 0.85rem;
          margin-bottom: 20px;
          font-family: 'SF Mono', Monaco, monospace;
        "
      >
curl -X POST http://127.0.0.1:18427/v1/messages \
  -H "Authorization: Bearer hpk_xxxxxxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "latest",
    "messages": [{"role": "user", "content": "Hello"}],
    "max_tokens": 1024
  }'</pre
      >

      <h3 style="font-size: 1rem; margin-bottom: 10px">{{ $t('guide.claude_config') }}</h3>
      <pre
        style="
          background: #f4f4f4;
          padding: 12px;
          border-radius: 4px;
          overflow-x: auto;
          font-size: 0.85rem;
          margin-bottom: 20px;
        "
      >
# {{ $t('guide.claude_config_hint') }}
ANTHROPIC_BASE_URL=http://127.0.0.1:18427
ANTHROPIC_AUTH_TOKEN=hpk_xxxxxxxxxxxxxxxx
ANTHROPIC_MODEL=latest</pre
      >

      <h3 style="font-size: 1rem; margin-bottom: 10px">{{ $t('guide.base_url') }}</h3>
      <p style="margin-bottom: 5px; color: #666">
        <code>http://127.0.0.1:18427</code>
      </p>
      <p style="margin-bottom: 20px; color: #666">{{ $t('guide.port_info') }}</p>

      <h3 style="font-size: 1rem; margin-bottom: 10px">{{ $t('guide.auth') }}</h3>
      <pre
        style="
          background: #f4f4f4;
          padding: 12px;
          border-radius: 4px;
          overflow-x: auto;
          font-size: 0.85rem;
          margin-bottom: 20px;
        "
      >
Authorization: Bearer &lt;your-api-key&gt;</pre
      >

      <h3 style="font-size: 1rem; margin-bottom: 10px">{{ $t('guide.chat_completions') }}</h3>
      <pre
        style="
          background: #f4f4f4;
          padding: 12px;
          border-radius: 4px;
          overflow-x: auto;
          font-size: 0.85rem;
          margin-bottom: 20px;
        "
      >
POST /v1/chat/completions
Authorization: Bearer &lt;your-api-key&gt;
Content-Type: application/json

{
  "model": "latest",
  "messages": [{"role": "user", "content": "hello"}]
}</pre
      >

      <h3 style="font-size: 1rem; margin-bottom: 10px">{{ $t('guide.messages') }}</h3>
      <pre
        style="
          background: #f4f4f4;
          padding: 12px;
          border-radius: 4px;
          overflow-x: auto;
          font-size: 0.85rem;
          margin-bottom: 20px;
        "
      >
POST /v1/messages
Authorization: Bearer &lt;your-api-key&gt;
anthropic-version: 2023-06-01
Content-Type: application/json

{
  "model": "latest",
  "messages": [{"role": "user", "content": "hello"}],
  "max_tokens": 1024
}</pre
      >

      <h3 style="font-size: 1rem; margin-bottom: 10px">{{ $t('guide.model_routing') }}</h3>
      <p style="margin-bottom: 10px; color: #666">
        {{ $t('guide.model_routing_hint') }}
      </p>
      <table style="margin-bottom: 20px; font-size: 0.9rem">
        <thead>
          <tr>
            <th style="text-align: left; padding: 6px 12px; background: #f8f8f8">
              {{ $t('guide.table_tag') }}
            </th>
            <th style="text-align: left; padding: 6px 12px; background: #f8f8f8">
              {{ $t('guide.table_provider') }}
            </th>
            <th style="text-align: left; padding: 6px 12px; background: #f8f8f8">
              {{ $t('guide.table_model') }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="!tags || tags.length === 0">
            <td colspan="3" style="padding: 12px; text-align: center; color: #888">
              {{ $t('guide.table_empty') }}
            </td>
          </tr>
          <tr v-for="tag in tags || []" :key="tag.id">
            <td style="padding: 6px 12px">
              <code>{{ tag.name }}</code>
            </td>
            <td style="padding: 6px 12px">{{ getProviderName(tag.provider_id) }}</td>
            <td style="padding: 6px 12px">
              <code>{{ getProviderModel(tag.provider_id) }}</code>
            </td>
          </tr>
        </tbody>
      </table>

      <h3 style="font-size: 1rem; margin-bottom: 10px">{{ $t('guide.list_models') }}</h3>
      <pre
        style="
          background: #f4f4f4;
          padding: 12px;
          border-radius: 4px;
          overflow-x: auto;
          font-size: 0.85rem;
        "
      >
GET /v1/models
Authorization: Bearer &lt;your-api-key&gt;</pre
      >
    </div>

    <!-- Providers -->
    <div v-show="tab === 'providers'">
      <div class="card">
        <div class="usage-guide">
          <h3>{{ $t('providers.guide_title') }}</h3>
          <p>{{ $t('providers.guide_desc') }}</p>
        </div>
        <div
          style="
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
          "
        >
          <h2>{{ $t('providers.title') }}</h2>
          <button class="btn btn-primary" @click="openProviderModal()">
            {{ $t('providers.add') }}
          </button>
        </div>
        <div v-if="error" class="error-msg">{{ error }}</div>
        <div v-if="!providers || providers.length === 0" class="empty-state">
          {{ $t('providers.empty') }}
        </div>
        <table v-if="providers && providers.length > 0">
          <thead>
            <tr>
              <th>{{ $t('providers.name') }}</th>
              <th>{{ $t('providers.endpoint') }}</th>
              <th>{{ $t('providers.model') }}</th>
              <th>{{ $t('providers.status') }}</th>
              <th>{{ $t('providers.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in providers || []" :key="p.id">
              <td>{{ p.name }}</td>
              <td style="max-width: 200px; overflow: hidden; text-overflow: ellipsis">
                {{ p.api_endpoint }}
              </td>
              <td style="max-width: 150px; overflow: hidden; text-overflow: ellipsis">
                {{ getModels(p.models).join(', ') }}
              </td>
              <td>
                <span class="badge" :class="p.is_active ? 'badge-active' : 'badge-inactive'">
                  {{ p.is_active ? $t('providers.active') : $t('providers.inactive') }}
                </span>
              </td>
              <td>
                <button
                  class="btn btn-sm"
                  style="background: #eee; margin-right: 4px"
                  @click="openProviderModal(p)"
                >
                  {{ $t('providers.edit') }}
                </button>
                <button class="btn btn-danger btn-sm" @click="deleteProvider(p.id)">
                  {{ $t('providers.delete') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- API Keys -->
    <div v-show="tab === 'keys'">
      <div class="card">
        <div class="usage-guide">
          <h3>{{ $t('keys.guide_title') }}</h3>
          <p>{{ $t('keys.guide_desc') }}</p>
        </div>
        <div
          style="
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
          "
        >
          <h2>{{ $t('keys.title') }}</h2>
          <button class="btn btn-primary" @click="openKeyModal()">{{ $t('keys.create') }}</button>
        </div>
        <div v-if="error" class="error-msg">{{ error }}</div>
        <div
          v-if="newKey"
          style="background: #d4edda; padding: 15px; border-radius: 4px; margin-bottom: 15px"
        >
          <strong>{{ $t('keys.new_key') }}</strong>
          <div class="api-key-display">{{ newKey }}</div>
          <p style="margin-top: 10px; font-size: 0.85rem; color: #155724">
            {{ $t('keys.new_key_hint') }}
          </p>
        </div>
        <div v-if="!apiKeys || apiKeys.length === 0" class="empty-state">
          {{ $t('keys.empty') }}
        </div>
        <table v-if="apiKeys && apiKeys.length > 0">
          <thead>
            <tr>
              <th>{{ $t('keys.name') }}</th>
              <th>{{ $t('keys.prefix') }}</th>
              <th>{{ $t('keys.expires') }}</th>
              <th>{{ $t('keys.status') }}</th>
              <th>{{ $t('keys.created') }}</th>
              <th>{{ $t('keys.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="k in apiKeys || []" :key="k.id">
              <td>{{ k.name }}</td>
              <td>
                <code>{{ k.key_prefix }}</code>
              </td>
              <td>{{ k.expires_at ? formatDate(k.expires_at) : $t('keys.never') }}</td>
              <td>
                <span class="badge" :class="k.is_active ? 'badge-active' : 'badge-inactive'">
                  {{ k.is_active ? $t('keys.active') : $t('keys.revoked') }}
                </span>
              </td>
              <td>{{ formatDate(k.created_at) }}</td>
              <td>
                <button class="btn btn-danger btn-sm" @click="deleteKey(k.id)">
                  {{ $t('keys.delete') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Usage Stats -->
    <div v-show="tab === 'stats'">
      <div class="card">
        <div class="usage-guide">
          <h3>{{ $t('stats.usage_guide_title') }}</h3>
          <p>{{ $t('stats.usage_guide_desc') }}</p>
        </div>

        <div
          style="
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
          "
        >
          <h2>{{ $t('stats.title') }}</h2>
          <div class="form-group" style="margin: 0">
            <select v-model="usageDays" @change="loadUsage()">
              <option value="7">{{ $t('stats.last_7') }}</option>
              <option value="30">{{ $t('stats.last_30') }}</option>
              <option value="90">{{ $t('stats.last_90') }}</option>
            </select>
          </div>
        </div>

        <!-- Global Stats -->
        <h3 style="margin-bottom: 10px">{{ $t('stats.global') }}</h3>
        <div v-if="usageStats && usageStats.global" class="stats-grid">
          <div class="stat-item">
            <div class="stat-value">{{ formatNumber(usageStats.global.total_requests) }}</div>
            <div class="stat-label">{{ $t('stats.total_requests') }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ formatNumber(usageStats.global.total_input_tokens) }}</div>
            <div class="stat-label">{{ $t('stats.input_tokens') }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ formatNumber(usageStats.global.total_output_tokens) }}</div>
            <div class="stat-label">{{ $t('stats.output_tokens') }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ usageStats.global.avg_latency_ms }}ms</div>
            <div class="stat-label">{{ $t('stats.avg_latency') }}</div>
          </div>
        </div>
        <div v-else class="empty-state">{{ $t('stats.no_data') }}</div>

        <!-- Time Series -->
        <h3 style="margin: 20px 0 10px">{{ $t('stats.time_series') }}</h3>
        <div
          v-if="usageStats && usageStats.time_series && usageStats.time_series.length > 0"
          class="charts-grid"
        >
          <div
            v-for="metric in ['requests', 'input_tokens', 'output_tokens', 'avg_latency_ms']"
            :key="metric"
            class="chart-card"
          >
            <div class="chart-title">
              {{ $t('stats.' + (metric === 'avg_latency_ms' ? 'avg_latency' : metric)) }}
            </div>
            <svg
              :viewBox="`0 0 ${chartWidth} ${chartHeight}`"
              class="line-chart"
              @mousemove="onChartHover($event, metric)"
              @mouseleave="chartTooltip.visible = false"
            >
              <defs>
                <linearGradient :id="`chartGradient-${metric}`" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#4a4a6a" stop-opacity="0.3" />
                  <stop offset="100%" stop-color="#4a4a6a" stop-opacity="0" />
                </linearGradient>
              </defs>
              <polyline
                :points="getChartPoints(metric)"
                fill="none"
                stroke="#4a4a6a"
                stroke-width="2"
                stroke-linejoin="round"
                stroke-linecap="round"
              />
              <polygon
                :points="getChartPoints(metric) + ` ${chartWidth},${chartHeight} 0,${chartHeight}`"
                :fill="`url(#chartGradient-${metric})`"
              />
              <g v-for="(pt, i) in usageStats.time_series" :key="pt.timestamp">
                <circle
                  :cx="getChartX(i)"
                  :cy="getChartY(getMetricValue(pt, metric), metric)"
                  r="4"
                  fill="#4a4a6a"
                  class="chart-dot"
                />
              </g>
              <line
                v-if="chartTooltip.visible && chartTooltip.metric === metric"
                :x1="chartTooltip.x"
                :y1="chartPadding.top"
                :x2="chartTooltip.x"
                :y2="chartHeight - chartPadding.bottom"
                stroke="#4a4a6a"
                stroke-width="1"
                stroke-dasharray="4"
                opacity="0.5"
              />
              <circle
                v-if="chartTooltip.visible && chartTooltip.metric === metric"
                :cx="chartTooltip.x"
                :cy="chartTooltip.y"
                r="6"
                fill="#4a4a6a"
                stroke="#fff"
                stroke-width="2"
              />
            </svg>
            <div class="chart-labels chart-end-label">
              <span></span>
              <span class="chart-total">Total: {{ formatNumber(getChartTotal(metric)) }}</span>
            </div>
            <div
              v-if="chartTooltip.visible && chartTooltip.metric === metric"
              class="chart-tooltip"
              :style="{ left: chartTooltip.x + 'px', top: chartTooltip.y - 50 + 'px' }"
            >
              <div class="tooltip-time">{{ chartTooltip.time }}</div>
              <div class="tooltip-value">{{ formatNumber(chartTooltip.value) }}</div>
            </div>
          </div>
        </div>
        <div v-else class="empty-state">{{ $t('stats.no_data') }}</div>

        <!-- By Key Stats -->
        <h3 style="margin: 20px 0 10px">{{ $t('stats.by_key') }}</h3>
        <div v-if="usageStats && usageStats.by_key && usageStats.by_key.length > 0">
          <table>
            <thead>
              <tr>
                <th>{{ $t('stats.key_name') }}</th>
                <th>{{ $t('stats.key_prefix') }}</th>
                <th>{{ $t('stats.total_requests') }}</th>
                <th>{{ $t('stats.input_tokens') }}</th>
                <th>{{ $t('stats.output_tokens') }}</th>
                <th>{{ $t('stats.avg_latency') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="ks in usageStats.by_key" :key="ks.key_id">
                <td>{{ ks.key_name || '-' }}</td>
                <td>
                  <code>{{ ks.key_prefix }}</code>
                </td>
                <td>{{ formatNumber(ks.total_requests) }}</td>
                <td>{{ formatNumber(ks.input_tokens) }}</td>
                <td>{{ formatNumber(ks.output_tokens) }}</td>
                <td>{{ ks.avg_latency_ms }}ms</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="empty-state">{{ $t('stats.no_data') }}</div>
      </div>
    </div>

    <!-- Tags -->
    <div v-show="tab === 'tags'">
      <div class="card">
        <div class="usage-guide">
          <h3>{{ $t('tags.guide_title') }}</h3>
          <p>{{ $t('tags.guide_desc') }}</p>
        </div>
        <div
          style="
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
          "
        >
          <h2>{{ $t('tags.title') }}</h2>
          <button class="btn btn-primary" @click="openTagModal()">{{ $t('tags.add') }}</button>
        </div>
        <div v-if="tagError" class="error-msg">{{ tagError }}</div>
        <div v-if="!tags || tags.length === 0" class="empty-state">{{ $t('tags.empty') }}</div>
        <table v-if="tags && tags.length > 0">
          <thead>
            <tr>
              <th>{{ $t('tags.name') }}</th>
              <th>{{ $t('tags.provider') }}</th>
              <th>{{ $t('tags.created') }}</th>
              <th>{{ $t('tags.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in tags || []" :key="t.id">
              <td>
                <code
                  style="
                    font-size: 0.9rem;
                    padding: 2px 6px;
                    background: #f4f4f4;
                    border-radius: 3px;
                  "
                  >{{ t.name }}</code
                >
              </td>
              <td>{{ getProviderName(t.provider_id) }}</td>
              <td>{{ formatDate(t.created_at) }}</td>
              <td>
                <button
                  class="btn btn-sm"
                  style="background: #eee; margin-right: 4px"
                  @click="openTagModal(t)"
                >
                  {{ $t('tags.edit') }}
                </button>
                <button
                  v-if="t.name !== 'default'"
                  class="btn btn-danger btn-sm"
                  @click="deleteTag(t.id)"
                >
                  {{ $t('tags.delete') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Logs -->
    <div v-show="tab === 'logs'">
      <div class="card">
        <div class="usage-guide">
          <h3>{{ $t('logs.guide_title') }}</h3>
          <p>{{ $t('logs.guide_desc') }}</p>
        </div>
        <div
          style="
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
          "
        >
          <h2>{{ $t('logs.title') }}</h2>
          <div style="display: flex; gap: 10px; align-items: center">
            <span style="font-size: 0.85rem; color: #666">{{ $t('logs.auto_refresh') }}</span>
            <button class="btn" style="background: #eee" @click="loadLogs()">
              {{ $t('logs.refresh') }}
            </button>
          </div>
        </div>
        <div
          style="
            display: flex;
            gap: 10px;
            align-items: center;
            margin-bottom: 15px;
            flex-wrap: wrap;
          "
        >
          <select
            v-model="logFilters.level"
            style="padding: 6px; border: 1px solid #ddd; border-radius: 4px; min-width: 100px"
          >
            <option value="">All</option>
            <option value="INFO">INFO</option>
            <option value="WARN">WARN</option>
            <option value="ERROR">ERROR</option>
          </select>
          <input
            v-model="logFilters.path"
            :placeholder="$t('logs.filters.path')"
            style="padding: 6px; border: 1px solid #ddd; border-radius: 4px; width: 120px"
          />
          <input
            v-model="logFilters.model"
            :placeholder="$t('logs.filters.model')"
            style="padding: 6px; border: 1px solid #ddd; border-radius: 4px; width: 100px"
          />
          <input
            v-model="logFilters.key_prefix"
            :placeholder="$t('logs.filters.key_prefix')"
            style="padding: 6px; border: 1px solid #ddd; border-radius: 4px; width: 100px"
          />
          <input
            v-model="logFilters.status"
            :placeholder="$t('logs.filters.status')"
            type="number"
            style="padding: 6px; border: 1px solid #ddd; border-radius: 4px; width: 80px"
          />
          <button class="btn" @click="loadLogs()">{{ $t('logs.filters.search') }}</button>
          <span style="margin-left: auto; font-size: 0.85rem; color: #666"
            >total: {{ logsTotal }}</span
          >
        </div>
        <div v-if="!logs || logs.length === 0" class="empty-state">{{ $t('logs.empty') }}</div>
        <div
          v-else
          ref="logsContainer"
          style="max-height: 500px; overflow-y: auto; border: 1px solid #eee; border-radius: 4px"
          @scroll="onLogsScroll"
        >
          <table>
            <thead>
              <tr>
                <th>{{ $t('logs.columns.time') }}</th>
                <th>{{ $t('logs.columns.level') }}</th>
                <th>{{ $t('logs.columns.method') }}</th>
                <th>{{ $t('logs.columns.path') }}</th>
                <th>{{ $t('logs.columns.status') }}</th>
                <th>{{ $t('logs.columns.latency') }}</th>
                <th>{{ $t('logs.columns.key_prefix') }}</th>
                <th>{{ $t('logs.columns.model') }}</th>
                <th>Tag</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(log, idx) in logs || []" :key="idx">
                <td style="font-size: 0.85rem; color: #666">{{ formatLogTime(log.time) }}</td>
                <td>
                  <span
                    class="badge"
                    :style="{
                      background:
                        log.level === 'INFO'
                          ? '#d4edda'
                          : log.level === 'WARN'
                            ? '#fff3cd'
                            : '#f8d7da',
                      color:
                        log.level === 'INFO'
                          ? '#155724'
                          : log.level === 'WARN'
                            ? '#856404'
                            : '#721c24',
                    }"
                  >
                    {{ log.level }}
                  </span>
                </td>
                <td>
                  <span
                    :style="{
                      color:
                        log.method === 'GET'
                          ? '#155724'
                          : log.method === 'POST'
                            ? '#0c5460'
                            : log.method === 'DELETE'
                              ? '#721c24'
                              : '#333',
                    }"
                    style="font-weight: 500"
                  >
                    {{ log.method }}
                  </span>
                </td>
                <td style="max-width: 200px; overflow: hidden; text-overflow: ellipsis">
                  {{ log.path }}
                </td>
                <td>
                  <span
                    :style="{
                      color:
                        log.status >= 500 ? '#721c24' : log.status >= 400 ? '#856404' : '#155724',
                      fontWeight: '500',
                    }"
                  >
                    {{ log.status }}
                  </span>
                </td>
                <td style="color: #666">{{ formatLatency(log.latency) }}</td>
                <td>
                  <code style="font-size: 0.8rem">{{ log.key_prefix }}</code>
                </td>
                <td style="color: #666; font-size: 0.85rem">{{ log.model || '-' }}</td>
                <td>
                  <code
                    v-if="log.tag"
                    style="
                      font-size: 0.75rem;
                      padding: 2px 5px;
                      background: #e8f4ea;
                      border-radius: 3px;
                      color: #155724;
                    "
                    >{{ log.tag }}</code
                  >
                  <span v-else style="color: #999">-</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div
          style="
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-top: 15px;
          "
        >
          <button
            class="btn"
            style="background: #eee"
            :disabled="logsOffset <= 0"
            @click="loadLogs(logsOffset - logsLimit)"
          >
            {{ $t('logs.pagination.prev') }}
          </button>
          <span style="font-size: 0.85rem; color: #666">
            offset: {{ logsOffset }}-{{ Math.min(logsOffset + logsLimit - 1, logsTotal - 1) }}
            {{ $t('logs.pagination.of') }} {{ logsTotal }}
          </span>
          <button
            class="btn"
            style="background: #eee"
            :disabled="logsOffset + logsLimit >= logsTotal"
            @click="loadLogs(logsOffset + logsLimit)"
          >
            {{ $t('logs.pagination.next') }}
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- Provider Modal -->
  <Transition>
    <div v-if="showProviderModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ editingProvider ? $t('providers.modal_edit') : $t('providers.modal_add') }}</h3>
          <button class="close-btn" @click="showProviderModal = false">&times;</button>
        </div>
        <div v-if="!editingProvider" style="margin-bottom: 15px">
          <div style="display: flex; gap: 10px; flex-wrap: wrap; align-items: flex-end">
            <div class="form-group" style="margin-bottom: 0; flex: 1; min-width: 150px">
              <label style="font-weight: 500; color: #555; display: block; margin-bottom: 5px"
                >{{ $t('providers.quick_preset') }}:</label
              >
              <select
                v-model="selectedPreset"
                @change="onPresetChange()"
                style="
                  width: 100%;
                  padding: 8px;
                  border: 1px solid #ddd;
                  border-radius: 4px;
                  font-size: 0.9rem;
                "
              >
                <option value="">{{ $t('providers.presets.select') }}</option>
                <option value="glm">{{ $t('providers.presets.glm') }}</option>
                <option value="minimax">{{ $t('providers.presets.minimax') }}</option>
                <option value="kimi">{{ $t('providers.presets.kimi') }}</option>
                <option value="mimo">{{ $t('providers.presets.mimo') }}</option>
              </select>
            </div>
            <div
              v-if="selectedPreset !== ''"
              class="form-group"
              style="margin-bottom: 0; flex: 1; min-width: 200px"
            >
              <label style="font-weight: 500; color: #555; display: block; margin-bottom: 5px"
                >{{ $t('providers.api_type') }}:</label
              >
              <select
                v-model="selectedApiType"
                @change="onApiTypeChange()"
                style="
                  width: 100%;
                  padding: 8px;
                  border: 1px solid #ddd;
                  border-radius: 4px;
                  font-size: 0.9rem;
                "
              >
                <option value="openai">{{ $t('providers.api_type_openai') }}</option>
                <option value="anthropic">{{ $t('providers.api_type_anthropic') }}</option>
              </select>
            </div>
          </div>
        </div>
        <form @submit.prevent="saveProvider()">
          <div class="form-group">
            <label>{{ $t('providers.form.name') }}</label>
            <input
              type="text"
              v-model="providerForm.name"
              required
              placeholder="e.g., Production OpenAI"
            />
          </div>
          <div class="form-group">
            <label>{{ $t('providers.form.endpoint') }}</label>
            <input
              type="text"
              v-model="providerForm.api_endpoint"
              required
              placeholder="https://api.openai.com"
            />
          </div>
          <div class="form-group">
            <label>{{ $t('providers.form.key') }}</label>
            <input
              type="password"
              v-model="providerForm.api_key"
              :required="!editingProvider"
              placeholder="sk-..."
            />
          </div>
          <div class="form-group">
            <label>{{ $t('providers.form.model') }}</label>
            <input type="text" v-model="providerForm.models" placeholder="e.g., gpt-4" />
          </div>
          <div style="display: flex; gap: 10px; justify-content: flex-end">
            <button
              type="button"
              class="btn"
              @click="showProviderModal = false"
              style="background: #eee"
            >
              {{ $t('providers.form.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary">{{ $t('providers.form.save') }}</button>
          </div>
        </form>
      </div>
    </div>
  </Transition>

  <!-- Key Modal -->
  <Transition>
    <div v-if="showKeyModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ $t('keys.modal_create') }}</h3>
          <button class="close-btn" @click="showKeyModal = false">&times;</button>
        </div>
        <form @submit.prevent="createKey()">
          <div class="form-group">
            <label>{{ $t('keys.form.name') }}</label>
            <input type="text" v-model="keyForm.name" required placeholder="e.g., Production App" />
          </div>
          <div class="form-group">
            <label>{{ $t('keys.form.limit') }}</label>
            <input type="number" v-model="keyForm.request_limit" placeholder="100" />
          </div>
          <div class="form-group">
            <label>{{ $t('keys.form.expires') }}</label>
            <input type="datetime-local" v-model="keyForm.expires_at" />
          </div>
          <div style="display: flex; gap: 10px; justify-content: flex-end">
            <button
              type="button"
              class="btn"
              @click="showKeyModal = false"
              style="background: #eee"
            >
              {{ $t('keys.form.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary">{{ $t('keys.form.create') }}</button>
          </div>
        </form>
      </div>
    </div>
  </Transition>

  <!-- Tag Modal -->
  <Transition>
    <div v-if="showTagModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ editingTag ? $t('tags.modal_edit') : $t('tags.modal_add') }}</h3>
          <button class="close-btn" @click="showTagModal = false">&times;</button>
        </div>
        <form @submit.prevent="saveTag()">
          <div class="form-group">
            <label>{{ $t('tags.form.name') }}</label>
            <input
              type="text"
              v-model="tagForm.name"
              required
              placeholder="e.g., latest"
              pattern="^[a-z0-9]+(-[a-z0-9]+)*$"
              title="lowercase letters, numbers, and hyphens only"
              :disabled="!!editingTag"
            />
            <p style="font-size: 0.8rem; color: #666; margin-top: 5px">
              {{ $t('tags.form.name_hint') }}
            </p>
          </div>
          <div class="form-group">
            <label>{{ $t('tags.form.provider') }}</label>
            <select v-model="tagForm.provider_id" required>
              <option value="">{{ $t('tags.form.select_provider') }}</option>
              <option v-for="p in providers || []" :key="p.id" :value="p.id">{{ p.name }}</option>
            </select>
          </div>
          <div style="display: flex; gap: 10px; justify-content: flex-end">
            <button
              type="button"
              class="btn"
              @click="showTagModal = false"
              style="background: #eee"
            >
              {{ $t('tags.form.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary">{{ $t('tags.form.save') }}</button>
          </div>
        </form>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue';
import { useRoute, useRouter } from 'vue-router';

const route = useRoute();
const router = useRouter();
import { useI18n } from 'vue-i18n';

const { t, locale } = useI18n();

function saveLocale() {
  localStorage.setItem('lang', locale.value);
}

const tab = ref('guide');
const providers = ref<any[]>([]);

function switchTab(newTab: string, isInit = false) {
  if (tab.value === 'logs' && newTab !== 'logs') {
    stopLogsAutoRefresh();
  }
  tab.value = newTab;
  if (!isInit) {
    router.push('/' + newTab);
  }
  if (newTab === 'stats') loadUsage();
  if (newTab === 'providers') loadProviders();
  if (newTab === 'keys') loadKeys();
  if (newTab === 'tags') {
    loadTags();
    loadProviders();
  }
  if (newTab === 'logs') startLogsAutoRefresh();
  if (newTab === 'guide') {
    loadTags();
    loadProviders();
  }
}
const apiKeys = ref<any[]>([]);
const tags = ref<any[]>([]);
const usageStats = ref<any>(null);
const usageDays = ref('7');
const error = ref<string | null>(null);
const tagError = ref<string | null>(null);
const newKey = ref<string | null>(null);
const showProviderModal = ref(false);
const showKeyModal = ref(false);
const showTagModal = ref(false);
const editingProvider = ref<any>(null);
const editingTag = ref<any>(null);
const selectedPreset = ref('');
const selectedApiType = ref('openai');
const logs = ref<any[]>([]);
const logsContainer = ref<HTMLElement | null>(null);
const logsOffset = ref(0);
const logsLimit = ref(20);
const logsTotal = ref(0);
const logsAutoScroll = ref(true);
const logFilters = reactive({
  level: '',
  path: '',
  model: '',
  key_prefix: '',
  status: '',
});
let logsInterval: ReturnType<typeof setInterval> | null = null;

const providerForm = reactive({
  name: '',
  api_endpoint: '',
  api_key: '',
  models: '',
});

const keyForm = reactive({
  name: '',
  request_limit: null as number | null,
  expires_at: '',
});

const tagForm = reactive({
  name: '',
  provider_id: '',
});

const presets: Record<string, any> = {
  glm: {
    name: 'GLM (z.ai)',
    openai_endpoint: 'https://open.bigmodel.cn/api/coding/paas/v4',
    anthropic_endpoint: 'https://open.bigmodel.cn/api/anthropic',
    models: 'glm-4.7-flash',
  },
  minimax: {
    name: 'MiniMax',
    openai_endpoint: 'https://api.minimaxi.com/v1',
    anthropic_endpoint: 'https://api.minimaxi.com/anthropic',
    models: 'MiniMax-M2.7-highspeed',
  },
  kimi: {
    name: 'Kimi',
    openai_endpoint: 'https://api.moonshot.ai/v1',
    anthropic_endpoint: 'https://api.kimi.com/coding',
    models: 'kimi-k2.5',
  },
  mimo: {
    name: 'MiMo (Xiaomi)',
    openai_endpoint: 'https://api.xiaomimimo.com/v1',
    anthropic_endpoint: 'https://api.xiaomimimo.com/anthropic',
    models: 'mimo-v2-pro',
  },
};

onMounted(() => {
  const path = route.path.replace('/', '') || 'guide';
  tab.value = path;
  switchTab(path, true);
});

onUnmounted(() => {
  stopLogsAutoRefresh();
});

function startLogsAutoRefresh() {
  stopLogsAutoRefresh();
  loadLogs(logsOffset.value);
  logsInterval = setInterval(() => {
    loadLogs(logsOffset.value);
  }, 5000);
}

function stopLogsAutoRefresh() {
  if (logsInterval) {
    clearInterval(logsInterval);
    logsInterval = null;
  }
}

async function loadLogs(offset = logsOffset.value) {
  try {
    const params = new URLSearchParams();
    params.set('offset', String(offset));
    params.set('limit', String(logsLimit.value));
    if (logFilters.level) params.set('level', logFilters.level);
    if (logFilters.path) params.set('path', logFilters.path);
    if (logFilters.model) params.set('model', logFilters.model);
    if (logFilters.key_prefix) params.set('key_prefix', logFilters.key_prefix);
    if (logFilters.status) params.set('status', logFilters.status);
    const res = await fetch(`/admin/logs?${params}`);
    if (!res.ok) throw new Error('Failed to load logs');
    const data = await res.json();
    logs.value = data?.logs || [];
    logsTotal.value = data?.total || 0;
    logsOffset.value = data?.offset || offset;
    logsLimit.value = data?.limit || 20;
    nextTick(() => {
      if (logsContainer.value && logsAutoScroll.value) {
        logsContainer.value.scrollTop = 0;
      }
    });
  } catch (e: any) {
    console.error(e);
    logs.value = [];
  }
}

function formatLogTime(timeStr: string): string {
  try {
    const d = new Date(timeStr);
    return (
      d.toLocaleTimeString('en-US', { hour12: false }) +
      '.' +
      String(d.getMilliseconds()).padStart(3, '0')
    );
  } catch {
    return timeStr;
  }
}

function formatLatency(latency: any): string {
  if (typeof latency === 'number') {
    const ms = Math.round(latency / 1000000);
    return ms + 'ms';
  }
  if (typeof latency === 'string') {
    const seconds = parseFloat(latency);
    if (!isNaN(seconds)) {
      return Math.round(seconds * 1000) + 'ms';
    }
  }
  return String(latency);
}

function onLogsScroll() {
  if (!logsContainer.value) return;
  const { scrollTop, scrollHeight, clientHeight } = logsContainer.value;
  logsAutoScroll.value = scrollTop < 50;
}

async function loadProviders() {
  try {
    const res = await fetch('/admin/providers', {
      headers: { 'Content-Type': 'application/json' },
    });
    if (res.status >= 500) throw new Error(t('error.failed_providers'));
    providers.value = (await res.json()) || [];
  } catch (e: any) {
    error.value = t('error.failed_providers');
    providers.value = [];
  }
}

async function loadKeys() {
  try {
    const res = await fetch('/admin/keys', {
      headers: { 'Content-Type': 'application/json' },
    });
    if (res.status >= 500) throw new Error(t('error.failed_keys'));
    apiKeys.value = (await res.json()) || [];
  } catch (e: any) {
    error.value = t('error.failed_keys');
    apiKeys.value = [];
  }
}

async function loadUsage() {
  try {
    const res = await fetch(`/admin/usage?days=${usageDays.value}`, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.ok) throw new Error(t('error.failed_usage'));
    usageStats.value = await res.json();
  } catch (e: any) {
    error.value = e.message;
  }
}

function openProviderModal(provider: any = null) {
  editingProvider.value = provider;
  selectedPreset.value = '';
  selectedApiType.value = 'openai';
  if (provider) {
    providerForm.name = provider.name;
    providerForm.api_endpoint = provider.api_endpoint;
    providerForm.api_key = '';
    providerForm.models = getSingleModel(provider.models);
  } else {
    providerForm.name = '';
    providerForm.api_endpoint = '';
    providerForm.api_key = '';
    providerForm.models = '';
  }
  showProviderModal.value = true;
}

async function saveProvider() {
  try {
    const body: any = {
      name: providerForm.name,
      api_endpoint: providerForm.api_endpoint,
      models: providerForm.models,
    };
    if (providerForm.api_key) {
      body.api_key = providerForm.api_key;
    }

    const url = editingProvider.value
      ? `/admin/providers/${editingProvider.value.id}`
      : '/admin/providers';
    const method = editingProvider.value ? 'PUT' : 'POST';

    const res = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    });

    if (!res.ok) throw new Error(t('error.failed_save'));
    showProviderModal.value = false;
    loadProviders();
  } catch (e: any) {
    error.value = e.message;
  }
}

async function deleteProvider(id: string) {
  if (!confirm(t('providers.delete_confirm'))) return;
  try {
    await fetch(`/admin/providers/${id}`, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
    });
    loadProviders();
  } catch (e: any) {
    error.value = e.message;
  }
}

function openKeyModal() {
  keyForm.name = '';
  keyForm.request_limit = null;
  keyForm.expires_at = '';
  newKey.value = null;
  showKeyModal.value = true;
}

async function createKey() {
  try {
    const body: any = { name: keyForm.name };
    if (keyForm.request_limit) {
      body.request_limit = parseInt(keyForm.request_limit as any);
    }
    if (keyForm.expires_at) {
      body.expires_at = new Date(keyForm.expires_at).toISOString();
    }

    const res = await fetch('/admin/keys', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    });

    if (!res.ok) throw new Error(t('error.failed_save_key'));
    const data = await res.json();
    newKey.value = data.api_key;
    showKeyModal.value = false;
    loadKeys();
  } catch (e: any) {
    error.value = e.message;
  }
}

async function deleteKey(id: string) {
  if (!confirm(t('keys.delete_confirm'))) return;
  try {
    await fetch(`/admin/keys/${id}`, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
    });
    loadKeys();
  } catch (e: any) {
    error.value = e.message;
  }
}

async function loadTags() {
  try {
    const res = await fetch('/admin/tags', {
      headers: { 'Content-Type': 'application/json' },
    });
    if (res.status >= 500) throw new Error(t('error.failed_tags'));
    tags.value = (await res.json()) || [];
  } catch (e: any) {
    tagError.value = t('error.failed_tags');
    tags.value = [];
  }
}

function getProviderName(providerId: string): string {
  const provider = providers.value.find((p) => p.id === providerId);
  return provider ? provider.name : providerId;
}

function getProviderModel(providerId: string): string {
  const provider = providers.value.find((p) => p.id === providerId);
  if (!provider) return providerId;
  try {
    const models = JSON.parse(provider.models);
    if (Array.isArray(models) && models.length > 0) return models[0];
    if (typeof models === 'string') return models;
  } catch {}
  return provider.models || '';
}

function openTagModal(tag: any = null) {
  editingTag.value = tag;
  if (tag) {
    tagForm.name = tag.name;
    tagForm.provider_id = tag.provider_id;
  } else {
    tagForm.name = '';
    tagForm.provider_id = '';
  }
  showTagModal.value = true;
}

async function saveTag() {
  try {
    const body = {
      name: tagForm.name,
      provider_id: tagForm.provider_id,
    };

    const url = editingTag.value ? `/admin/tags/${editingTag.value.id}` : '/admin/tags';
    const method = editingTag.value ? 'PUT' : 'POST';

    const res = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    });

    if (!res.ok) {
      if (res.status === 409) {
        throw new Error(t('error.tag_duplicate'));
      }
      if (res.status === 400) {
        throw new Error(t('error.tag_invalid_name'));
      }
      throw new Error(t('error.failed_save_tag'));
    }

    showTagModal.value = false;
    loadTags();
  } catch (e: any) {
    tagError.value = e.message;
  }
}

async function deleteTag(id: string) {
  if (!confirm(t('tags.delete_confirm'))) return;
  try {
    await fetch(`/admin/tags/${id}`, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
    });
    loadTags();
  } catch (e: any) {
    tagError.value = e.message;
  }
}

function getModels(modelsJson: string): string[] {
  try {
    const parsed = JSON.parse(modelsJson);
    if (Array.isArray(parsed)) return parsed;
    if (typeof parsed === 'string') return [parsed];
    return [];
  } catch {
    return [];
  }
}

function getSingleModel(modelsJson: string): string {
  try {
    const parsed = JSON.parse(modelsJson);
    if (Array.isArray(parsed)) return parsed[0] || '';
    if (typeof parsed === 'string') return parsed;
    return '';
  } catch {
    return '';
  }
}

function fillPreset(preset: string, apiType: string) {
  const p = presets[preset];
  if (p) {
    providerForm.name = p.name;
    providerForm.api_endpoint = apiType === 'anthropic' ? p.anthropic_endpoint : p.openai_endpoint;
    providerForm.api_key = '';
    providerForm.models = p.models;
  }
}

function onPresetChange() {
  if (selectedPreset.value) {
    fillPreset(selectedPreset.value, selectedApiType.value);
  }
}

function onApiTypeChange() {
  if (selectedPreset.value) {
    fillPreset(selectedPreset.value, selectedApiType.value);
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

function formatNumber(n: number): string {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
  return n.toString();
}

const chartWidth = 600;
const chartHeight = 150;
const chartPadding = { top: 10, right: 10, bottom: 5, left: 10 };

function getChartX(index: number): number {
  const ts = usageStats.value?.time_series || [];
  if (ts.length <= 1) return chartWidth / 2;
  return (
    chartPadding.left +
    (index / (ts.length - 1)) * (chartWidth - chartPadding.left - chartPadding.right)
  );
}

function getMetricValue(pt: any, metric: string): number {
  switch (metric) {
    case 'requests':
      return pt.cum_requests || 0;
    case 'input_tokens':
      return pt.cum_input_tokens || 0;
    case 'output_tokens':
      return pt.cum_output_tokens || 0;
    case 'avg_latency_ms':
      return pt.requests || 0;
    default:
      return pt[metric] || 0;
  }
}

function getChartY(value: number, metric: string): number {
  const ts = usageStats.value?.time_series || [];
  const values = ts.map((p: any) => getMetricValue(p, metric));
  const max = Math.max(...values, 1);
  const chartH = chartHeight - chartPadding.top - chartPadding.bottom;
  return chartPadding.top + chartH - (value / max) * chartH;
}

function getChartPoints(metric: string): string {
  const ts = usageStats.value?.time_series || [];
  return ts
    .map((p: any, i: number) => `${getChartX(i)},${getChartY(getMetricValue(p, metric), metric)}`)
    .join(' ');
}

function getChartTotal(metric: string): number {
  const ts = usageStats.value?.time_series || [];
  if (ts.length === 0) return 0;
  const last = ts[ts.length - 1];
  switch (metric) {
    case 'requests':
      return last.cum_requests || 0;
    case 'input_tokens':
      return last.cum_input_tokens || 0;
    case 'output_tokens':
      return last.cum_output_tokens || 0;
    case 'avg_latency_ms':
      return ts.reduce((sum: number, p: any) => sum + (p.requests || 0), 0);
    default:
      return 0;
  }
}

const chartTooltip = reactive({
  visible: false,
  metric: '',
  x: 0,
  y: 0,
  time: '',
  value: 0,
});

function onChartHover(event: MouseEvent, metric: string) {
  const svg = event.currentTarget as SVGElement;
  const rect = svg.getBoundingClientRect();
  const scaleX = chartWidth / rect.width;
  const svgX = (event.clientX - rect.left) * scaleX;

  const ts = usageStats.value?.time_series || [];
  if (ts.length === 0) return;

  let closestIdx = 0;
  let closestDist = Infinity;
  for (let i = 0; i < ts.length; i++) {
    const x = getChartX(i);
    const dist = Math.abs(x - svgX);
    if (dist < closestDist) {
      closestDist = dist;
      closestIdx = i;
    }
  }

  chartTooltip.visible = true;
  chartTooltip.metric = metric;
  chartTooltip.x = getChartX(closestIdx);
  chartTooltip.y = getChartY(getMetricValue(ts[closestIdx], metric), metric);
  chartTooltip.time = formatChartTime(ts[closestIdx].timestamp);
  chartTooltip.value = getMetricValue(ts[closestIdx], metric);
}

function formatChartTime(ts: string): string {
  return new Date(ts).toLocaleString('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}
</script>

<style>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #f5f5f5;
  color: #333;
}
.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}
header {
  background: #1a1a2e;
  color: #fff;
  padding: 20px 0;
  margin-bottom: 30px;
}
header h1 {
  text-align: center;
  font-size: 1.8rem;
}
.card {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 20px;
  margin-bottom: 20px;
}
.card h2 {
  margin-bottom: 15px;
  font-size: 1.2rem;
  color: #1a1a2e;
  border-bottom: 2px solid #4a4a6a;
  padding-bottom: 8px;
}
table {
  width: 100%;
  border-collapse: collapse;
}
th,
td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #eee;
}
th {
  background: #f8f8f8;
  font-weight: 600;
  color: #555;
}
.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s;
}
.btn-primary {
  background: #4a4a6a;
  color: #fff;
}
.btn-primary:hover {
  background: #3a3a5a;
}
.btn-danger {
  background: #dc3545;
  color: #fff;
}
.btn-danger:hover {
  background: #c82333;
}
.btn-sm {
  padding: 4px 10px;
  font-size: 0.8rem;
}
.form-group {
  margin-bottom: 15px;
}
.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: 500;
  color: #555;
}
.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 0.95rem;
}
.form-group textarea {
  min-height: 80px;
  resize: vertical;
}
.form-row {
  display: flex;
  gap: 15px;
}
.form-row .form-group {
  flex: 1;
}
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.modal-content {
  background: #fff;
  border-radius: 8px;
  padding: 25px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.modal-header h3 {
  font-size: 1.2rem;
}
.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #888;
}
.badge {
  display: inline-block;
  padding: 3px 8px;
  border-radius: 3px;
  font-size: 0.75rem;
  font-weight: 500;
}
.badge-active {
  background: #d4edda;
  color: #155724;
}
.badge-inactive {
  background: #f8d7da;
  color: #721c24;
}
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 15px;
}
.stat-item {
  background: #f8f8f8;
  padding: 15px;
  border-radius: 6px;
  text-align: center;
}
.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: #4a4a6a;
}
.stat-label {
  font-size: 0.85rem;
  color: #666;
  margin-top: 5px;
}
.tab-nav {
  display: flex;
  gap: 5px;
  margin-bottom: 20px;
  border-bottom: 2px solid #eee;
  align-items: center;
}
.tab-btn {
  padding: 10px 20px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.95rem;
  color: #666;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  transition: all 0.2s;
}
.tab-btn:hover {
  color: #4a4a6a;
}
.tab-btn.active {
  color: #4a4a6a;
  border-bottom-color: #4a4a6a;
  font-weight: 500;
}
.empty-state {
  text-align: center;
  padding: 40px;
  color: #888;
}
.api-key-display {
  background: #f8f8f8;
  padding: 15px;
  border-radius: 4px;
  font-family: monospace;
  word-break: break-all;
  margin-top: 10px;
}
.error-msg {
  background: #f8d7da;
  color: #721c24;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
}

/* Language Switcher */
.lang-switch {
  margin-left: auto;
  padding: 6px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 0.85rem;
  background: #fff;
  cursor: pointer;
}

/* Modal transition */
.v-enter-active,
.v-leave-active {
  transition: opacity 0.2s ease;
}
.v-enter-from,
.v-leave-to {
  opacity: 0;
}

/* Time Series Charts */
.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
}
.chart-card {
  border: 1px solid #eee;
  border-radius: 4px;
  padding: 12px;
  background: #fafafa;
  position: relative;
}
.chart-title {
  font-weight: 600;
  font-size: 0.9rem;
  margin-bottom: 8px;
  color: #333;
}
.line-chart {
  width: 100%;
  height: 120px;
  overflow: visible;
}
.chart-dot {
  transition: r 0.2s;
}
.chart-dot:hover {
  r: 6;
}
.chart-labels {
  display: flex;
  justify-content: space-between;
  font-size: 0.7rem;
  color: #888;
  margin-top: 5px;
  padding: 0 5px;
}
.chart-end-label {
  justify-content: flex-end;
}
.chart-total {
  font-weight: 600;
  color: #4a4a6a;
}
.chart-tooltip {
  position: absolute;
  background: #333;
  color: #fff;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 0.75rem;
  pointer-events: none;
  z-index: 100;
  transform: translateX(-50%);
  white-space: nowrap;
}
.tooltip-time {
  color: #aaa;
  margin-bottom: 2px;
}
.tooltip-value {
  font-weight: 600;
  font-size: 0.9rem;
}
.usage-guide {
  background: #f8f9fa;
  border-left: 4px solid #4a4a6a;
  padding: 15px 20px;
  margin-bottom: 20px;
  border-radius: 4px;
}
.usage-guide h3 {
  margin: 0 0 8px 0;
  font-size: 1rem;
  color: #333;
}
.usage-guide p {
  margin: 0;
  color: #666;
  font-size: 0.9rem;
  line-height: 1.5;
}

/* Architecture Diagram */
.diagram-section {
  margin-bottom: 25px;
  padding: 15px;
  background: #fafbfc;
  border-radius: 8px;
  border: 1px solid #e8e8e8;
  text-align: center;
}
.diagram-container {
  background: #fff;
  border-radius: 6px;
  padding: 10px 0;
  margin-bottom: 10px;
}
.flow-diagram {
  display: block;
  width: auto;
  max-width: 100%;
  height: auto;
  max-height: 200px;
  margin: 0 auto;
}
.box-fill {
  fill: #fff;
  stroke: #4a4a6a;
  stroke-width: 2;
  transition: all 0.3s ease;
}
.box-fill:hover {
  fill: #f0f0f8;
  stroke-width: 3;
}
.box-title {
  font-size: 12px;
  font-weight: 600;
  fill: #333;
}
.box-subtitle {
  font-size: 11px;
  fill: #666;
}
.box-subtitle-sm {
  font-size: 10px;
  fill: #888;
}
.box-fill-client {
  fill: none;
  stroke: #4a4a6a;
  stroke-width: 2;
}
.box-title-client {
  font-size: 12px;
  font-weight: 600;
  fill: #333;
}
.box-subtitle-client {
  font-size: 10px;
  fill: #666;
}
.box-fill-static {
  fill: none;
  stroke: #4a4a6a;
  stroke-width: 2;
}
.box-fill-ai {
  fill: none;
  stroke: #4a4a6a;
  stroke-width: 2;
}
.box-title-ai {
  font-size: 11px;
  font-weight: 600;
  fill: #333;
}
.static-arrow {
  stroke: #4a4a6a;
  stroke-width: 2;
}
.flow-line {
  stroke: #4a4a6a;
  stroke-width: 2;
  stroke-dasharray: 8 4;
  animation: dash 1.5s linear infinite;
}
.flow-label {
  font-size: 10px;
  fill: #888;
}
.flow-direction {
  font-size: 11px;
  fill: #aaa;
  font-style: italic;
}
@keyframes dash {
  to {
    stroke-dashoffset: -24;
  }
}
.flow-dot {
  fill: #4a4a6a;
  opacity: 0;
}
.dot-1 {
  animation: flowDot1 2s ease-in-out infinite;
}
.dot-2 {
  animation: flowDot2 2s ease-in-out infinite 0.4s;
}
.dot-3 {
  animation: flowDot3 2s ease-in-out infinite 0.8s;
}
@keyframes flowDot1 {
  0% {
    opacity: 0;
    transform: translate(30px, 100px);
  }
  20% {
    opacity: 1;
  }
  50% {
    opacity: 1;
    transform: translate(150px, 100px);
  }
  70% {
    opacity: 0;
  }
  100% {
    opacity: 0;
    transform: translate(150px, 100px);
  }
}
@keyframes flowDot2 {
  0% {
    opacity: 0;
    transform: translate(170px, 100px);
  }
  20% {
    opacity: 1;
  }
  50% {
    opacity: 1;
    transform: translate(310px, 100px);
  }
  70% {
    opacity: 0;
  }
  100% {
    opacity: 0;
    transform: translate(310px, 100px);
  }
}
@keyframes flowDot3 {
  0% {
    opacity: 0;
    transform: translate(330px, 100px);
  }
  20% {
    opacity: 1;
  }
  50% {
    opacity: 1;
    transform: translate(450px, 100px);
  }
  70% {
    opacity: 0;
  }
  100% {
    opacity: 0;
    transform: translate(450px, 100px);
  }
}
.diagram-legend {
  display: flex;
  justify-content: center;
  gap: 20px;
  font-size: 0.8rem;
  color: #666;
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
}
.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #4a4a6a;
}
</style>
