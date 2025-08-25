<script>
    import { createEventDispatcher } from "svelte";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import OverlayPanel from "@/components/base/OverlayPanel.svelte";
    import Field from "@/components/base/Field.svelte";
    import FormattedDate from "@/components/base/FormattedDate.svelte";
    import CodeEditor from "@/components/base/CodeEditor.svelte";
    import tooltip from "@/actions/tooltip";

    const dispatch = createEventDispatcher();

    let panel;
    let log = {};
    let isLoading = false;

    export function show(logData) {
        load(logData);
        return panel?.show();
    }

    export function hide() {
        return panel?.hide();
    }

    async function load(logData) {
        if (!logData?.id) {
            log = {};
            return;
        }

        isLoading = true;

        try {
            // 获取完整的日志详情，包括关联的脚本信息
            log = await ApiClient.collection("js_execution_logs").getOne(logData.id, {
                expand: "script_id"
            });
        } catch (err) {
            console.warn(err);
            log = logData; // 使用传入的数据作为备选
        }

        isLoading = false;
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

    function formatCpuUsage(percentage) {
        if (!percentage) return "-";
        return `${percentage.toFixed(2)}%`;
    }

    function copyToClipboard(text) {
        navigator.clipboard.writeText(text).then(() => {
            // 可以添加成功提示
        }).catch(err => {
            console.warn('Failed to copy text: ', err);
        });
    }
</script>

<OverlayPanel
    bind:this={panel}
    class="overlay-panel-lg log-details-panel"
    on:hide
    on:show
>
    <svelte:fragment slot="header">
        <h4>
            Execution Log Details
            {#if log.expand?.script_id}
                <span class="txt-hint">({log.expand.script_id.name})</span>
            {/if}
        </h4>
    </svelte:fragment>

    {#if isLoading}
        <div class="block txt-center p-lg">
            <span class="loader loader-sm" />
            <div class="txt-hint m-t-sm">Loading log details...</div>
        </div>
    {:else}
        <div class="grid">
            <!-- 基本信息 -->
            <div class="col-lg-12">
                <h6 class="m-b-sm">Basic Information</h6>
            </div>

            <div class="col-lg-6">
                <Field class="form-field" name="script_name" let:uniqueId>
                    <label for={uniqueId}>Script Name</label>
                    <div class="form-field-addon">
                        {#if log.expand?.script_id}
                            <span class="txt">{log.expand.script_id.name}</span>
                        {:else}
                            <span class="txt txt-hint">Unknown script</span>
                        {/if}
                    </div>
                </Field>
            </div>

            <div class="col-lg-6">
                <Field class="form-field" name="status" let:uniqueId>
                    <label for={uniqueId}>Status</label>
                    <div class="form-field-addon">
                        <span class="label label-sm label-{getStatusColor(log.status)}">
                            <i class="{getStatusIcon(log.status)}" aria-hidden="true" />
                            {log.status}
                        </span>
                    </div>
                </Field>
            </div>

            <div class="col-lg-4">
                <Field class="form-field" name="start_time" let:uniqueId>
                    <label for={uniqueId}>Start Time</label>
                    <div class="form-field-addon">
                        <FormattedDate date={log.start_time} />
                    </div>
                </Field>
            </div>

            <div class="col-lg-4">
                <Field class="form-field" name="end_time" let:uniqueId>
                    <label for={uniqueId}>End Time</label>
                    <div class="form-field-addon">
                        {#if log.end_time}
                            <FormattedDate date={log.end_time} />
                        {:else}
                            <span class="txt-hint">-</span>
                        {/if}
                    </div>
                </Field>
            </div>

            <div class="col-lg-4">
                <Field class="form-field" name="execution_time" let:uniqueId>
                    <label for={uniqueId}>Execution Time</label>
                    <div class="form-field-addon">
                        <span class="txt">{formatDuration(log.execution_time)}</span>
                    </div>
                </Field>
            </div>

            <!-- 性能指标 -->
            <div class="col-lg-12">
                <hr class="m-t-lg m-b-sm" />
                <h6 class="m-b-sm">Performance Metrics</h6>
            </div>

            <div class="col-lg-4">
                <Field class="form-field" name="memory_usage" let:uniqueId>
                    <label for={uniqueId}>Memory Usage</label>
                    <div class="form-field-addon">
                        <span class="txt">{formatMemory(log.memory_usage)}</span>
                    </div>
                </Field>
            </div>

            <div class="col-lg-4">
                <Field class="form-field" name="cpu_usage" let:uniqueId>
                    <label for={uniqueId}>CPU Usage</label>
                    <div class="form-field-addon">
                        <span class="txt">{formatCpuUsage(log.cpu_usage)}</span>
                    </div>
                </Field>
            </div>

            <div class="col-lg-4">
                <Field class="form-field" name="exit_code" let:uniqueId>
                    <label for={uniqueId}>Exit Code</label>
                    <div class="form-field-addon">
                        {#if log.exit_code !== null && log.exit_code !== undefined}
                            <span class="txt">{log.exit_code}</span>
                        {:else}
                            <span class="txt-hint">-</span>
                        {/if}
                    </div>
                </Field>
            </div>

            <!-- 输出和错误信息 -->
            {#if log.output || log.error_message}
                <div class="col-lg-12">
                    <hr class="m-t-lg m-b-sm" />
                    <h6 class="m-b-sm">Output & Errors</h6>
                </div>

                {#if log.output}
                    <div class="col-lg-12">
                        <Field class="form-field" name="output" let:uniqueId>
                            <label for={uniqueId}>
                                Output
                                <button
                                    type="button"
                                    class="btn btn-xs btn-circle btn-transparent"
                                    use:tooltip={"Copy to clipboard"}
                                    on:click={() => copyToClipboard(log.output)}
                                >
                                    <i class="ri-file-copy-line" aria-hidden="true" />
                                </button>
                            </label>
                            <CodeEditor
                                value={log.output}
                                language="text"
                                readonly={true}
                                maxHeight="200px"
                            />
                        </Field>
                    </div>
                {/if}

                {#if log.error_message}
                    <div class="col-lg-12">
                        <Field class="form-field" name="error_message" let:uniqueId>
                            <label for={uniqueId}>
                                Error Message
                                <button
                                    type="button"
                                    class="btn btn-xs btn-circle btn-transparent"
                                    use:tooltip={"Copy to clipboard"}
                                    on:click={() => copyToClipboard(log.error_message)}
                                >
                                    <i class="ri-file-copy-line" aria-hidden="true" />
                                </button>
                            </label>
                            <CodeEditor
                                value={log.error_message}
                                language="text"
                                readonly={true}
                                maxHeight="200px"
                            />
                        </Field>
                    </div>
                {/if}
            {/if}

            <!-- 堆栈跟踪 -->
            {#if log.stack_trace}
                <div class="col-lg-12">
                    <hr class="m-t-lg m-b-sm" />
                    <h6 class="m-b-sm">Stack Trace</h6>
                </div>

                <div class="col-lg-12">
                    <Field class="form-field" name="stack_trace" let:uniqueId>
                        <label for={uniqueId}>
                            Stack Trace
                            <button
                                type="button"
                                class="btn btn-xs btn-circle btn-transparent"
                                use:tooltip={"Copy to clipboard"}
                                on:click={() => copyToClipboard(log.stack_trace)}
                            >
                                <i class="ri-file-copy-line" aria-hidden="true" />
                            </button>
                        </label>
                        <CodeEditor
                            value={log.stack_trace}
                            language="text"
                            readonly={true}
                            maxHeight="300px"
                        />
                    </Field>
                </div>
            {/if}

            <!-- 元数据 -->
            <div class="col-lg-12">
                <hr class="m-t-lg m-b-sm" />
                <h6 class="m-b-sm">Metadata</h6>
            </div>

            <div class="col-lg-6">
                <Field class="form-field" name="created" let:uniqueId>
                    <label for={uniqueId}>Created</label>
                    <div class="form-field-addon">
                        <FormattedDate date={log.created} />
                    </div>
                </Field>
            </div>

            <div class="col-lg-6">
                <Field class="form-field" name="updated" let:uniqueId>
                    <label for={uniqueId}>Updated</label>
                    <div class="form-field-addon">
                        <FormattedDate date={log.updated} />
                    </div>
                </Field>
            </div>
        </div>
    {/if}

    <svelte:fragment slot="footer">
        <button type="button" class="btn btn-secondary" on:click={() => hide()}>
            <span class="txt">Close</span>
        </button>
        {#if log.expand?.script_id}
            <button
                type="button"
                class="btn btn-outline"
                on:click={() => {
                    // 跳转到脚本编辑页面
                    window.location.hash = `/scripts?filter=id='${log.script_id}'`;
                    hide();
                }}
            >
                <i class="ri-external-link-line" aria-hidden="true" />
                <span class="txt">View Script</span>
            </button>
        {/if}
    </svelte:fragment>
</OverlayPanel>

<style>
    .log-details-panel :global(.overlay-panel-content) {
        max-height: 90vh;
        overflow-y: auto;
    }

    .form-field-addon {
        padding: 8px 12px;
        background: var(--baseAlt1Color);
        border: 1px solid var(--baseAlt2Color);
        border-radius: 4px;
        min-height: 40px;
        display: flex;
        align-items: center;
    }
</style>