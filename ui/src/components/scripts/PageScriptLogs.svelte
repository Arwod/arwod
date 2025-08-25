<script>
    import { onMount } from "svelte";
    import { link } from "svelte-spa-router";
    import active from "svelte-spa-router/active";
    import CommonHelper from "@/utils/CommonHelper";
    import ApiClient from "@/utils/ApiClient";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import Searchbar from "@/components/base/Searchbar.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import FormattedDate from "@/components/base/FormattedDate.svelte";
    import Field from "@/components/base/Field.svelte";
    import ScriptsSidebar from "@/components/scripts/ScriptsSidebar.svelte";
    import ScriptLogDetailsPanel from "@/components/scripts/ScriptLogDetailsPanel.svelte";
    import { addSuccessToast, addErrorToast } from "@/stores/toasts";
    import { setErrors } from "@/stores/errors";
    import tooltip from "@/actions/tooltip";

    const pageSize = 30;

    let logs = [];
    let currentPage = 1;
    let totalPages = 1;
    let totalItems = 0;
    let isLoading = false;
    let filter = "";
    let sort = "-created";
    let detailsPanel;
    let bulkSelected = {};
    let lastSelectedId = "";

    $: canLoadMore = totalPages > currentPage;
    $: areAllSelected = logs.length > 0 && logs.every((log) => bulkSelected[log.id]);
    $: hasSelected = Object.keys(bulkSelected).length > 0;
    $: selectedTotal = Object.keys(bulkSelected).length;

    export let queryParams = {};

    $: if (typeof queryParams?.filter !== "undefined") {
        filter = queryParams.filter;
    }

    $: if (typeof queryParams?.sort !== "undefined") {
        sort = queryParams.sort;
    }

    onMount(() => {
        return load();
    });

    async function load(reset = true) {
        if (isLoading) {
            return;
        }

        isLoading = true;

        try {
            const page = reset ? 1 : currentPage + 1;
            const result = await ApiClient.collection("js_execution_logs").getList(page, pageSize, {
                filter: filter,
                sort: sort,
                expand: "script_id",
            });

            if (reset) {
                logs = result.items || [];
                currentPage = 1;
            } else {
                logs = logs.concat(result.items || []);
                currentPage++;
            }

            totalPages = result.totalPages;
            totalItems = result.totalItems;
        } catch (err) {
            if (!err?.isAbort) {
                console.warn(err);
                addErrorToast(err.data?.message || `Failed to load execution logs.`);
            }
        }

        isLoading = false;
    }

    function toggleSelectAll() {
        if (areAllSelected) {
            deselectAll();
        } else {
            selectAll();
        }
    }

    function selectAll() {
        for (const log of logs) {
            bulkSelected[log.id] = log;
        }
        bulkSelected = bulkSelected;
    }

    function deselectAll() {
        bulkSelected = {};
    }

    function toggleSelect(log) {
        if (!bulkSelected[log.id]) {
            bulkSelected[log.id] = log;
        } else {
            delete bulkSelected[log.id];
        }
        bulkSelected = bulkSelected;
        lastSelectedId = log.id;
    }

    function selectTo(log) {
        if (!lastSelectedId) {
            return toggleSelect(log);
        }

        const lastIndex = logs.findIndex((l) => l.id === lastSelectedId);
        const currentIndex = logs.findIndex((l) => l.id === log.id);

        if (lastIndex === -1 || currentIndex === -1) {
            return toggleSelect(log);
        }

        const startIndex = Math.min(lastIndex, currentIndex);
        const endIndex = Math.max(lastIndex, currentIndex);

        for (let i = startIndex; i <= endIndex; i++) {
            bulkSelected[logs[i].id] = logs[i];
        }
        bulkSelected = bulkSelected;
    }

    async function deleteSelected() {
        if (!hasSelected || !window.confirm(`Do you really want to delete the selected ${selectedTotal} log(s)?`)) {
            return;
        }

        let promises = [];
        for (const logId of Object.keys(bulkSelected)) {
            promises.push(ApiClient.collection("js_execution_logs").delete(logId));
        }

        isLoading = true;

        try {
            await Promise.all(promises);
            addSuccessToast(`Successfully deleted ${selectedTotal} log(s).`);
            deselectAll();
            return load();
        } catch (err) {
            if (!err?.isAbort) {
                console.warn(err);
                addErrorToast(err.data?.message || `Failed to delete the selected logs.`);
            }
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
</script>

<PageWrapper>
    <ScriptsSidebar />

    <main class="page-content" tabindex="-1">
        <header class="page-header">
            <nav class="breadcrumbs">
                <div class="breadcrumb-item">
                    <a href="/scripts" use:link>Scripts</a>
                </div>
                <div class="breadcrumb-item">Execution Logs</div>
            </nav>
        </header>

        <div class="page-header-wrapper m-b-sm">
            <header class="page-header">
                <h1 class="page-title">Script Execution Logs</h1>
                <div class="btns-group">
                    <RefreshButton on:refresh={() => load()} />
                    {#if hasSelected}
                        <button
                            type="button"
                            class="btn btn-sm btn-danger btn-outline"
                            on:click={() => deleteSelected()}
                        >
                            <i class="ri-delete-bin-7-line" aria-hidden="true" />
                            <span class="txt">Delete selected</span>
                        </button>
                    {/if}
                </div>
            </header>

            <div class="searchbar-wrapper">
                <Searchbar
                    value={filter}
                    placeholder={`Search logs...`}
                    autocompleteCollection="js_execution_logs"
                    on:submit={(e) => {
                        filter = e.detail;
                        load();
                    }}
                />
            </div>
        </div>

        <div class="table-wrapper">
            <table class="table" class:table-loading={isLoading}>
                <thead>
                    <tr>
                        <th class="bulk-select-col min-width">
                            <!-- svelte-ignore a11y-missing-attribute -->
                            <label class="form-field">
                                <input
                                    type="checkbox"
                                    checked={areAllSelected}
                                    on:change={() => toggleSelectAll()}
                                />
                                <i class="ri-checkbox-blank-line" aria-hidden="true" />
                            </label>
                        </th>
                        <th class="col-type-text col-field-script">
                            <div class="col-header-content">
                                <i class="ri-code-s-slash-line" aria-hidden="true" />
                                <span class="txt">Script</span>
                            </div>
                        </th>
                        <th class="col-type-select col-field-status">
                            <div class="col-header-content">
                                <i class="ri-checkbox-circle-line" aria-hidden="true" />
                                <span class="txt">Status</span>
                            </div>
                        </th>
                        <th class="col-type-number col-field-execution_time">
                            <div class="col-header-content">
                                <i class="ri-timer-line" aria-hidden="true" />
                                <span class="txt">Duration</span>
                            </div>
                        </th>
                        <th class="col-type-number col-field-memory_usage">
                            <div class="col-header-content">
                                <i class="ri-database-line" aria-hidden="true" />
                                <span class="txt">Memory</span>
                            </div>
                        </th>
                        <th class="col-type-date col-field-start_time">
                            <div class="col-header-content">
                                <i class="ri-calendar-event-line" aria-hidden="true" />
                                <span class="txt">Started</span>
                            </div>
                        </th>
                        <th class="col-type-action min-width">
                            <div class="col-header-content">
                                <i class="ri-more-line" aria-hidden="true" />
                                <span class="txt">Actions</span>
                            </div>
                        </th>
                    </tr>
                </thead>
                <tbody>
                    {#each logs as log (log.id)}
                        <tr tabindex="0" class="row-handle">
                            <td class="bulk-select-col min-width">
                                <!-- svelte-ignore a11y-missing-attribute -->
                                <label class="form-field">
                                    <input
                                        type="checkbox"
                                        checked={!!bulkSelected[log.id]}
                                        on:change={(e) => {
                                            if (e.shiftKey) {
                                                selectTo(log);
                                            } else {
                                                toggleSelect(log);
                                            }
                                        }}
                                    />
                                    <i class="ri-checkbox-blank-line" aria-hidden="true" />
                                </label>
                            </td>
                            <td class="col-type-text col-field-script">
                                <div class="flex">
                                    {#if log.expand?.script_id}
                                        <span class="txt">{log.expand.script_id.name}</span>
                                    {:else}
                                        <span class="txt txt-hint">Unknown script</span>
                                    {/if}
                                </div>
                            </td>
                            <td class="col-type-select col-field-status">
                                <span class="label label-sm label-{getStatusColor(log.status)}">
                                    <i class="{getStatusIcon(log.status)}" aria-hidden="true" />
                                    {log.status}
                                </span>
                            </td>
                            <td class="col-type-number col-field-execution_time">
                                <span class="txt">{formatDuration(log.execution_time)}</span>
                            </td>
                            <td class="col-type-number col-field-memory_usage">
                                <span class="txt">{formatMemory(log.memory_usage)}</span>
                            </td>
                            <td class="col-type-date col-field-start_time">
                                <FormattedDate date={log.start_time} />
                            </td>
                            <td class="col-type-action min-width">
                                <div class="flex gap-5">
                                    <button
                                        type="button"
                                        class="btn btn-sm btn-circle"
                                        use:tooltip={"View details"}
                                        on:click={() => detailsPanel?.show(log)}
                                    >
                                        <i class="ri-eye-line" aria-hidden="true" />
                                    </button>
                                </div>
                            </td>
                        </tr>
                    {/each}

                    {#if !logs.length && !isLoading}
                        <tr>
                            <td colspan="7" class="txt-center txt-hint p-xs">
                                <h6>No execution logs found.</h6>
                                {#if filter?.length}
                                    <button
                                        type="button"
                                        class="btn btn-hint btn-expanded m-t-sm"
                                        on:click={() => {
                                            filter = "";
                                            load();
                                        }}
                                    >
                                        <span class="txt">Clear filters</span>
                                    </button>
                                {/if}
                            </td>
                        </tr>
                    {/if}
                </tbody>
            </table>
        </div>

        {#if hasSelected}
            <div class="bulkbar">
                <div class="txt">
                    Selected <strong>{selectedTotal}</strong>
                    {selectedTotal === 1 ? "log" : "logs"}
                </div>
                <button type="button" class="btn btn-sm btn-danger" on:click={() => deleteSelected()}>
                    <i class="ri-delete-bin-7-line" aria-hidden="true" />
                    <span class="txt">Delete selected</span>
                </button>
            </div>
        {/if}

        {#if canLoadMore}
            <div class="block txt-center m-t-sm">
                <button
                    type="button"
                    class="btn btn-lg btn-secondary btn-expanded"
                    class:btn-loading={isLoading}
                    class:btn-disabled={isLoading}
                    on:click={() => load(false)}
                >
                    <span class="txt">Load more ({logs.length}/{totalItems})</span>
                </button>
            </div>
        {/if}
    </main>
</PageWrapper>

<ScriptLogDetailsPanel bind:this={detailsPanel} />