<script>
    import { onMount, onDestroy } from "svelte";
    import { link } from "svelte-spa-router";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import Field from "@/components/base/Field.svelte";
    import ScriptsSidebar from "@/components/scripts/ScriptsSidebar.svelte";
    import { addErrorToast } from "@/stores/toasts";
    import tooltip from "@/actions/tooltip";

    let isLoading = false;
    let refreshInterval = null;
    let autoRefresh = true;
    let refreshRate = 5000; // 5秒刷新一次
    
    // 性能统计数据
    let stats = {
        totalExecutions: 0,
        successfulExecutions: 0,
        failedExecutions: 0,
        averageExecutionTime: 0,
        totalMemoryUsage: 0,
        averageMemoryUsage: 0,
        activeCronJobs: 0,
        runningScripts: 0
    };
    
    // 实时执行数据
    let recentExecutions = [];
    let performanceData = {
        executionTimes: [],
        memoryUsage: [],
        cpuUsage: [],
        timestamps: []
    };
    
    // 脚本状态统计
    let scriptStats = [];
    
    // 系统资源使用情况
    let systemResources = {
        totalMemory: 0,
        usedMemory: 0,
        cpuUsage: 0,
        activeConnections: 0
    };

    $: successRate = stats.totalExecutions > 0 ? ((stats.successfulExecutions / stats.totalExecutions) * 100).toFixed(1) : 0;
    $: failureRate = stats.totalExecutions > 0 ? ((stats.failedExecutions / stats.totalExecutions) * 100).toFixed(1) : 0;
    $: memoryUsagePercentage = systemResources.totalMemory > 0 ? ((systemResources.usedMemory / systemResources.totalMemory) * 100).toFixed(1) : 0;

    onMount(() => {
        loadMonitoringData();
        if (autoRefresh) {
            startAutoRefresh();
        }
    });

    onDestroy(() => {
        stopAutoRefresh();
    });

    function startAutoRefresh() {
        stopAutoRefresh();
        refreshInterval = setInterval(() => {
            loadMonitoringData();
        }, refreshRate);
    }

    function stopAutoRefresh() {
        if (refreshInterval) {
            clearInterval(refreshInterval);
            refreshInterval = null;
        }
    }

    function toggleAutoRefresh() {
        autoRefresh = !autoRefresh;
        if (autoRefresh) {
            startAutoRefresh();
        } else {
            stopAutoRefresh();
        }
    }

    async function loadMonitoringData() {
        if (isLoading) return;
        
        isLoading = true;
        
        try {
            // 并行加载所有监控数据
            const [statsResult, executionsResult, scriptsResult] = await Promise.all([
                loadStats(),
                loadRecentExecutions(),
                loadScriptStats()
            ]);
            
            // 模拟系统资源数据（实际项目中应该从后端API获取）
            systemResources = {
                totalMemory: 8192, // 8GB
                usedMemory: Math.random() * 4096, // 随机使用内存
                cpuUsage: Math.random() * 100,
                activeConnections: Math.floor(Math.random() * 50)
            };
            
        } catch (err) {
            console.warn(err);
            addErrorToast(err.data?.message || "Failed to load monitoring data.");
        }
        
        isLoading = false;
    }

    async function loadStats() {
        try {
            // 获取执行统计
            const logsResult = await ApiClient.collection("js_execution_logs").getList(1, 1, {
                filter: "",
                sort: "-created"
            });
            
            const totalExecutions = logsResult.totalItems;
            
            // 获取成功和失败的执行数
            const [successResult, failedResult] = await Promise.all([
                ApiClient.collection("js_execution_logs").getList(1, 1, {
                    filter: "status='success'"
                }),
                ApiClient.collection("js_execution_logs").getList(1, 1, {
                    filter: "status='error'"
                })
            ]);
            
            // 获取平均执行时间和内存使用
            const recentLogs = await ApiClient.collection("js_execution_logs").getList(1, 100, {
                filter: "",
                sort: "-created"
            });
            
            let totalTime = 0;
            let totalMemory = 0;
            let validLogs = 0;
            
            recentLogs.items.forEach(log => {
                if (log.execution_time) {
                    totalTime += log.execution_time;
                    validLogs++;
                }
                if (log.memory_usage) {
                    totalMemory += log.memory_usage;
                }
            });
            
            stats = {
                totalExecutions,
                successfulExecutions: successResult.totalItems,
                failedExecutions: failedResult.totalItems,
                averageExecutionTime: validLogs > 0 ? Math.round(totalTime / validLogs) : 0,
                totalMemoryUsage: totalMemory,
                averageMemoryUsage: recentLogs.items.length > 0 ? Math.round(totalMemory / recentLogs.items.length) : 0,
                activeCronJobs: 0, // 需要从脚本表获取
                runningScripts: 0 // 需要从执行状态获取
            };
            
        } catch (err) {
            console.warn("Failed to load stats:", err);
        }
    }

    async function loadRecentExecutions() {
        try {
            const result = await ApiClient.collection("js_execution_logs").getList(1, 20, {
                filter: "",
                sort: "-created",
                expand: "script_id"
            });
            
            recentExecutions = result.items;
            
            // 更新性能数据图表
            const now = new Date();
            performanceData.timestamps.push(now.toLocaleTimeString());
            
            // 计算最近的平均执行时间
            const recentAvgTime = result.items.length > 0 
                ? result.items.reduce((sum, log) => sum + (log.execution_time || 0), 0) / result.items.length
                : 0;
            performanceData.executionTimes.push(recentAvgTime);
            
            // 计算最近的平均内存使用
            const recentAvgMemory = result.items.length > 0
                ? result.items.reduce((sum, log) => sum + (log.memory_usage || 0), 0) / result.items.length
                : 0;
            performanceData.memoryUsage.push(recentAvgMemory / (1024 * 1024)); // 转换为MB
            
            // 模拟CPU使用率
            performanceData.cpuUsage.push(Math.random() * 100);
            
            // 保持最近30个数据点
            if (performanceData.timestamps.length > 30) {
                performanceData.timestamps.shift();
                performanceData.executionTimes.shift();
                performanceData.memoryUsage.shift();
                performanceData.cpuUsage.shift();
            }
            
        } catch (err) {
            console.warn("Failed to load recent executions:", err);
        }
    }

    async function loadScriptStats() {
        try {
            const scriptsResult = await ApiClient.collection("js_scripts").getList(1, 50, {
                filter: "",
                sort: "name"
            });
            
            // 为每个脚本统计执行情况
            const scriptStatsPromises = scriptsResult.items.map(async (script) => {
                const logsResult = await ApiClient.collection("js_execution_logs").getList(1, 1, {
                    filter: `script_id='${script.id}'`
                });
                
                const successResult = await ApiClient.collection("js_execution_logs").getList(1, 1, {
                    filter: `script_id='${script.id}' && status='success'`
                });
                
                return {
                    id: script.id,
                    name: script.name,
                    status: script.status,
                    trigger_type: script.trigger_type,
                    totalExecutions: logsResult.totalItems,
                    successfulExecutions: successResult.totalItems,
                    lastExecution: null // 可以从最近的日志获取
                };
            });
            
            scriptStats = await Promise.all(scriptStatsPromises);
            
        } catch (err) {
            console.warn("Failed to load script stats:", err);
        }
    }

    function getStatusColor(status) {
        switch (status) {
            case "success":
                return "success";
            case "error":
                return "danger";
            case "timeout":
                return "warning";
            case "running":
                return "info";
            default:
                return "";
        }
    }

    function getStatusIcon(status) {
        switch (status) {
            case "success":
                return "ri-check-line";
            case "error":
                return "ri-close-line";
            case "timeout":
                return "ri-time-line";
            case "running":
                return "ri-loader-line";
            default:
                return "ri-question-line";
        }
    }

    function formatDuration(ms) {
        if (!ms) return "-";
        if (ms < 1000) return `${ms}ms`;
        return `${(ms / 1000).toFixed(2)}s`;
    }

    function formatMemory(bytes) {
        if (!bytes) return "-";
        const mb = bytes / (1024 * 1024);
        return `${mb.toFixed(2)}MB`;
    }

    function formatBytes(bytes) {
        if (!bytes) return "0 B";
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
    }
</script>

<PageWrapper>
    <ScriptsSidebar />

    <main class="page-content" tabindex="-1">
        <header class="page-header">
            <nav class="breadcrumbs">
                <div class="breadcrumb-item">
                    <a href="/scripts" use:link>Scripts</a>
                </div>
                <div class="breadcrumb-item">Performance Monitor</div>
            </nav>
        </header>

        <div class="page-header-wrapper m-b-sm">
            <header class="page-header">
                <h1 class="page-title">Performance Monitor</h1>
                <div class="btns-group">
                    <button
                        type="button"
                        class="btn btn-sm btn-outline"
                        class:btn-success={autoRefresh}
                        on:click={toggleAutoRefresh}
                        use:tooltip={autoRefresh ? "Disable auto refresh" : "Enable auto refresh"}
                    >
                        <i class="{autoRefresh ? 'ri-pause-line' : 'ri-play-line'}" aria-hidden="true" />
                        <span class="txt">{autoRefresh ? 'Auto' : 'Manual'}</span>
                    </button>
                    <RefreshButton on:refresh={() => loadMonitoringData()} />
                </div>
            </header>
        </div>

        <!-- 统计卡片 -->
        <div class="grid m-b-lg">
            <div class="col-lg-3 col-md-6">
                <div class="card">
                    <div class="card-content">
                        <div class="flex flex-gap-sm">
                            <div class="flex-fill">
                                <div class="txt-hint txt-sm">Total Executions</div>
                                <div class="txt-xl txt-bold">{stats.totalExecutions}</div>
                            </div>
                            <div class="flex-shrink-0">
                                <i class="ri-play-circle-line txt-2xl txt-info" aria-hidden="true" />
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <div class="col-lg-3 col-md-6">
                <div class="card">
                    <div class="card-content">
                        <div class="flex flex-gap-sm">
                            <div class="flex-fill">
                                <div class="txt-hint txt-sm">Success Rate</div>
                                <div class="txt-xl txt-bold txt-success">{successRate}%</div>
                            </div>
                            <div class="flex-shrink-0">
                                <i class="ri-check-circle-line txt-2xl txt-success" aria-hidden="true" />
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <div class="col-lg-3 col-md-6">
                <div class="card">
                    <div class="card-content">
                        <div class="flex flex-gap-sm">
                            <div class="flex-fill">
                                <div class="txt-hint txt-sm">Avg Execution Time</div>
                                <div class="txt-xl txt-bold">{formatDuration(stats.averageExecutionTime)}</div>
                            </div>
                            <div class="flex-shrink-0">
                                <i class="ri-timer-line txt-2xl txt-warning" aria-hidden="true" />
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <div class="col-lg-3 col-md-6">
                <div class="card">
                    <div class="card-content">
                        <div class="flex flex-gap-sm">
                            <div class="flex-fill">
                                <div class="txt-hint txt-sm">Memory Usage</div>
                                <div class="txt-xl txt-bold">{memoryUsagePercentage}%</div>
                            </div>
                            <div class="flex-shrink-0">
                                <i class="ri-database-line txt-2xl txt-primary" aria-hidden="true" />
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- 性能图表区域 -->
        <div class="grid m-b-lg">
            <div class="col-lg-8">
                <div class="card">
                    <div class="card-header">
                        <h5>Performance Trends</h5>
                    </div>
                    <div class="card-content">
                        <div class="performance-chart">
                            <!-- 简单的性能趋势显示 -->
                            <div class="chart-legend">
                                <div class="legend-item">
                                    <span class="legend-color legend-execution-time"></span>
                                    <span class="txt-sm">Execution Time (ms)</span>
                                </div>
                                <div class="legend-item">
                                    <span class="legend-color legend-memory"></span>
                                    <span class="txt-sm">Memory Usage (MB)</span>
                                </div>
                                <div class="legend-item">
                                    <span class="legend-color legend-cpu"></span>
                                    <span class="txt-sm">CPU Usage (%)</span>
                                </div>
                            </div>
                            
                            <div class="chart-data">
                                {#if performanceData.timestamps.length > 0}
                                    <div class="data-points">
                                        {#each performanceData.timestamps as timestamp, i}
                                            <div class="data-point">
                                                <div class="timestamp">{timestamp}</div>
                                                <div class="values">
                                                    <div class="value execution-time">{performanceData.executionTimes[i]?.toFixed(0) || 0}ms</div>
                                                    <div class="value memory">{performanceData.memoryUsage[i]?.toFixed(1) || 0}MB</div>
                                                    <div class="value cpu">{performanceData.cpuUsage[i]?.toFixed(1) || 0}%</div>
                                                </div>
                                            </div>
                                        {/each}
                                    </div>
                                {:else}
                                    <div class="txt-center txt-hint p-lg">
                                        <i class="ri-line-chart-line txt-2xl" aria-hidden="true" />
                                        <div class="m-t-sm">No performance data available</div>
                                    </div>
                                {/if}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <div class="col-lg-4">
                <div class="card">
                    <div class="card-header">
                        <h5>System Resources</h5>
                    </div>
                    <div class="card-content">
                        <div class="resource-item">
                            <div class="resource-label">Memory Usage</div>
                            <div class="resource-value">
                                {formatBytes(systemResources.usedMemory * 1024 * 1024)} / {formatBytes(systemResources.totalMemory * 1024 * 1024)}
                            </div>
                            <div class="progress-bar">
                                <div class="progress-fill" style="width: {memoryUsagePercentage}%"></div>
                            </div>
                        </div>

                        <div class="resource-item">
                            <div class="resource-label">CPU Usage</div>
                            <div class="resource-value">{systemResources.cpuUsage.toFixed(1)}%</div>
                            <div class="progress-bar">
                                <div class="progress-fill" style="width: {systemResources.cpuUsage}%"></div>
                            </div>
                        </div>

                        <div class="resource-item">
                            <div class="resource-label">Active Connections</div>
                            <div class="resource-value">{systemResources.activeConnections}</div>
                        </div>

                        <div class="resource-item">
                            <div class="resource-label">Running Scripts</div>
                            <div class="resource-value">{stats.runningScripts}</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- 脚本状态统计 -->
        <div class="grid m-b-lg">
            <div class="col-lg-6">
                <div class="card">
                    <div class="card-header">
                        <h5>Script Statistics</h5>
                    </div>
                    <div class="card-content">
                        <div class="table-wrapper">
                            <table class="table table-sm">
                                <thead>
                                    <tr>
                                        <th>Script Name</th>
                                        <th>Status</th>
                                        <th>Type</th>
                                        <th>Executions</th>
                                        <th>Success Rate</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {#each scriptStats as script}
                                        <tr>
                                            <td>
                                                <a href="/scripts?filter=id='{script.id}'" use:link class="txt">
                                                    {script.name}
                                                </a>
                                            </td>
                                            <td>
                                                <span class="label label-sm label-{script.status === 'active' ? 'success' : 'secondary'}">
                                                    {script.status}
                                                </span>
                                            </td>
                                            <td>
                                                <span class="txt-sm txt-hint">{script.trigger_type}</span>
                                            </td>
                                            <td>
                                                <span class="txt">{script.totalExecutions}</span>
                                            </td>
                                            <td>
                                                {#if script.totalExecutions > 0}
                                                    <span class="txt">{((script.successfulExecutions / script.totalExecutions) * 100).toFixed(1)}%</span>
                                                {:else}
                                                    <span class="txt-hint">-</span>
                                                {/if}
                                            </td>
                                        </tr>
                                    {/each}
                                    
                                    {#if scriptStats.length === 0}
                                        <tr>
                                            <td colspan="5" class="txt-center txt-hint p-sm">
                                                No scripts found
                                            </td>
                                        </tr>
                                    {/if}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>

            <div class="col-lg-6">
                <div class="card">
                    <div class="card-header">
                        <h5>Recent Executions</h5>
                    </div>
                    <div class="card-content">
                        <div class="execution-list">
                            {#each recentExecutions.slice(0, 10) as execution}
                                <div class="execution-item">
                                    <div class="execution-info">
                                        <div class="execution-script">
                                            {#if execution.expand?.script_id}
                                                <span class="txt">{execution.expand.script_id.name}</span>
                                            {:else}
                                                <span class="txt txt-hint">Unknown script</span>
                                            {/if}
                                        </div>
                                        <div class="execution-meta">
                                            <span class="txt-sm txt-hint">{formatDuration(execution.execution_time)}</span>
                                            <span class="txt-sm txt-hint">•</span>
                                            <span class="txt-sm txt-hint">{formatMemory(execution.memory_usage)}</span>
                                        </div>
                                    </div>
                                    <div class="execution-status">
                                        <span class="label label-sm label-{getStatusColor(execution.status)}">
                                            <i class="{getStatusIcon(execution.status)}" aria-hidden="true" />
                                            {execution.status}
                                        </span>
                                    </div>
                                </div>
                            {/each}
                            
                            {#if recentExecutions.length === 0}
                                <div class="txt-center txt-hint p-sm">
                                    <i class="ri-history-line txt-xl" aria-hidden="true" />
                                    <div class="m-t-sm">No recent executions</div>
                                </div>
                            {/if}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </main>
</PageWrapper>

<style>
    .performance-chart {
        min-height: 300px;
    }
    
    .chart-legend {
        display: flex;
        gap: 1rem;
        margin-bottom: 1rem;
        flex-wrap: wrap;
    }
    
    .legend-item {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }
    
    .legend-color {
        width: 12px;
        height: 12px;
        border-radius: 2px;
    }
    
    .legend-execution-time {
        background-color: var(--primaryColor);
    }
    
    .legend-memory {
        background-color: var(--successColor);
    }
    
    .legend-cpu {
        background-color: var(--warningColor);
    }
    
    .chart-data {
        overflow-x: auto;
    }
    
    .data-points {
        display: flex;
        gap: 1rem;
        min-width: max-content;
    }
    
    .data-point {
        text-align: center;
        min-width: 80px;
    }
    
    .timestamp {
        font-size: 0.75rem;
        color: var(--txtHintColor);
        margin-bottom: 0.5rem;
    }
    
    .values {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
    }
    
    .value {
        font-size: 0.75rem;
        padding: 2px 4px;
        border-radius: 2px;
    }
    
    .value.execution-time {
        background-color: var(--primaryAlt1Color);
        color: var(--primaryColor);
    }
    
    .value.memory {
        background-color: var(--successAlt1Color);
        color: var(--successColor);
    }
    
    .value.cpu {
        background-color: var(--warningAlt1Color);
        color: var(--warningColor);
    }
    
    .resource-item {
        margin-bottom: 1rem;
    }
    
    .resource-item:last-child {
        margin-bottom: 0;
    }
    
    .resource-label {
        font-size: 0.875rem;
        color: var(--txtHintColor);
        margin-bottom: 0.25rem;
    }
    
    .resource-value {
        font-weight: 600;
        margin-bottom: 0.5rem;
    }
    
    .progress-bar {
        height: 6px;
        background-color: var(--baseAlt2Color);
        border-radius: 3px;
        overflow: hidden;
    }
    
    .progress-fill {
        height: 100%;
        background-color: var(--primaryColor);
        transition: width 0.3s ease;
    }
    
    .execution-list {
        max-height: 400px;
        overflow-y: auto;
    }
    
    .execution-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0.75rem 0;
        border-bottom: 1px solid var(--baseAlt2Color);
    }
    
    .execution-item:last-child {
        border-bottom: none;
    }
    
    .execution-info {
        flex: 1;
    }
    
    .execution-script {
        margin-bottom: 0.25rem;
    }
    
    .execution-meta {
        display: flex;
        gap: 0.5rem;
        align-items: center;
    }
    
    .execution-status {
        flex-shrink: 0;
    }
</style>