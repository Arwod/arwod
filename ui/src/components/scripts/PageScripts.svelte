<script>
    import { onMount } from "svelte";
    import { link } from "svelte-spa-router";
    import active from "svelte-spa-router/active";
    import CommonHelper from "@/utils/CommonHelper";
    import ApiClient from "@/utils/ApiClient";

    import Searchbar from "@/components/base/Searchbar.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import FormattedDate from "@/components/base/FormattedDate.svelte";
    import Field from "@/components/base/Field.svelte";
    import ScriptUpsertPanel from "@/components/scripts/ScriptUpsertPanel.svelte";
    import ScriptsSidebar from "@/components/scripts/ScriptsSidebar.svelte";
    import { addSuccessToast, addErrorToast } from "@/stores/toasts";
    import { setErrors } from "@/stores/errors";
    import tooltip from "@/actions/tooltip";

    const pageSize = 30;

    let scripts = [];
    let currentPage = 1;
    let totalPages = 1;
    let totalItems = 0;
    let isLoading = false;
    let filter = "";
    let sort = "-created";
    let upsertPanel;
    let bulkSelected = {};
    let lastSelectedId = "";

    $: canLoadMore = totalPages > currentPage;
    $: areAllSelected = scripts.length > 0 && scripts.every((script) => bulkSelected[script.id]);
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
            const result = await ApiClient.collection("js_scripts").getList(page, pageSize, {
                filter: filter,
                sort: sort,
                expand: "",
            });

            if (reset) {
                scripts = result.items || [];
                currentPage = 1;
            } else {
                scripts = scripts.concat(result.items || []);
                currentPage++;
            }

            totalPages = result.totalPages;
            totalItems = result.totalItems;
        } catch (err) {
            if (!err?.isAbort) {
                console.warn(err);
                addErrorToast(err.data?.message || `Failed to load scripts.`);
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
        for (const script of scripts) {
            bulkSelected[script.id] = script;
        }
        bulkSelected = bulkSelected;
    }

    function deselectAll() {
        bulkSelected = {};
    }

    function toggleSelect(script) {
        if (!bulkSelected[script.id]) {
            bulkSelected[script.id] = script;
        } else {
            delete bulkSelected[script.id];
        }
        bulkSelected = bulkSelected;
        lastSelectedId = script.id;
    }

    function selectTo(script) {
        if (!lastSelectedId) {
            return toggleSelect(script);
        }

        const lastIndex = scripts.findIndex((s) => s.id === lastSelectedId);
        const currentIndex = scripts.findIndex((s) => s.id === script.id);

        if (lastIndex === -1 || currentIndex === -1) {
            return toggleSelect(script);
        }

        const startIndex = Math.min(lastIndex, currentIndex);
        const endIndex = Math.max(lastIndex, currentIndex);

        for (let i = startIndex; i <= endIndex; i++) {
            bulkSelected[scripts[i].id] = scripts[i];
        }
        bulkSelected = bulkSelected;
    }

    async function deleteSelected() {
        if (!hasSelected || !window.confirm(`Do you really want to delete the selected ${selectedTotal} script(s)?`)) {
            return;
        }

        let promises = [];
        for (const scriptId of Object.keys(bulkSelected)) {
            promises.push(ApiClient.collection("js_scripts").delete(scriptId));
        }

        isLoading = true;

        try {
            await Promise.all(promises);
            addSuccessToast(`Successfully deleted ${selectedTotal} script(s).`);
            deselectAll();
            return load();
        } catch (err) {
            if (!err?.isAbort) {
                console.warn(err);
                addErrorToast(err.data?.message || `Failed to delete the selected scripts.`);
            }
        }

        isLoading = false;
    }

    async function executeScript(script) {
        if (!window.confirm(`Do you want to execute the script "${script.name}"?`)) {
            return;
        }

        try {
            // 这里需要调用脚本执行API
            const result = await ApiClient.send("/api/scripts/execute", {
                method: "POST",
                body: {
                    script_id: script.id
                }
            });
            
            addSuccessToast(`Script "${script.name}" executed successfully.`);
        } catch (err) {
            console.warn(err);
            addErrorToast(err.data?.message || `Failed to execute script "${script.name}".`);
        }
    }

    function getStatusColor(status) {
        switch (status) {
            case "active":
                return "success";
            case "inactive":
                return "warning";
            default:
                return "";
        }
    }

    function getTriggerTypeIcon(triggerType) {
        switch (triggerType) {
            case "manual":
                return "ri-play-circle-line";
            case "hook":
                return "ri-git-branch-line";
            case "cron":
                return "ri-time-line";
            default:
                return "ri-code-line";
        }
    }
</script>

<div class="page-wrapper scripts-layout">
    <ScriptsSidebar />

    <main class="page-content" tabindex="-1">
        <header class="page-header">
            <nav class="breadcrumbs">
                <div class="breadcrumb-item">Scripts</div>
            </nav>
        </header>

        <div class="page-header-wrapper m-b-sm">
            <header class="page-header">
                <h1 class="page-title">JavaScript Scripts</h1>
                <div class="btns-group">
                    <RefreshButton on:refresh={() => load()} />
                    <button
                        type="button"
                        class="btn btn-expanded"
                        on:click={() => upsertPanel?.show()}
                    >
                        <i class="ri-add-line" aria-hidden="true" />
                        <span class="txt">New script</span>
                    </button>
                </div>
            </header>

            <div class="searchbar-wrapper">
                <Searchbar
                    value={filter}
                    placeholder={`Search scripts...`}
                    autocompleteCollection="js_scripts"
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
                        <th class="col-type-text col-field-name">
                            <div class="col-header-content">
                                <i class="ri-text" aria-hidden="true" />
                                <span class="txt">Name</span>
                            </div>
                        </th>
                        <th class="col-type-text col-field-description">
                            <div class="col-header-content">
                                <i class="ri-file-text-line" aria-hidden="true" />
                                <span class="txt">Description</span>
                            </div>
                        </th>
                        <th class="col-type-select col-field-trigger_type">
                            <div class="col-header-content">
                                <i class="ri-settings-3-line" aria-hidden="true" />
                                <span class="txt">Trigger</span>
                            </div>
                        </th>
                        <th class="col-type-select col-field-status">
                            <div class="col-header-content">
                                <i class="ri-checkbox-circle-line" aria-hidden="true" />
                                <span class="txt">Status</span>
                            </div>
                        </th>
                        <th class="col-type-date col-field-created">
                            <div class="col-header-content">
                                <i class="ri-calendar-event-line" aria-hidden="true" />
                                <span class="txt">Created</span>
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
                    {#each scripts as script (script.id)}
                        <tr tabindex="0" class="row-handle">
                            <td class="bulk-select-col min-width">
                                <!-- svelte-ignore a11y-missing-attribute -->
                                <label class="form-field">
                                    <input
                                        type="checkbox"
                                        checked={!!bulkSelected[script.id]}
                                        on:change={(e) => {
                                            if (e.shiftKey) {
                                                selectTo(script);
                                            } else {
                                                toggleSelect(script);
                                            }
                                        }}
                                    />
                                    <i class="ri-checkbox-blank-line" aria-hidden="true" />
                                </label>
                            </td>
                            <td class="col-type-text col-field-name">
                                <div class="flex">
                                    <i class="{getTriggerTypeIcon(script.trigger_type)} txt-sm m-r-5" aria-hidden="true" />
                                    <span class="txt">{script.name}</span>
                                    {#if script.is_system}
                                        <i class="ri-shield-check-line txt-xs txt-hint m-l-5" use:tooltip={"System script"} aria-hidden="true" />
                                    {/if}
                                </div>
                            </td>
                            <td class="col-type-text col-field-description">
                                <span class="txt txt-ellipsis">{script.description || "-"}</span>
                            </td>
                            <td class="col-type-select col-field-trigger_type">
                                <span class="label label-sm">
                                    <i class="{getTriggerTypeIcon(script.trigger_type)}" aria-hidden="true" />
                                    {script.trigger_type}
                                </span>
                            </td>
                            <td class="col-type-select col-field-status">
                                <span class="label label-sm label-{getStatusColor(script.status)}">
                                    {script.status}
                                </span>
                            </td>
                            <td class="col-type-date col-field-created">
                                <FormattedDate date={script.created} />
                            </td>
                            <td class="col-type-action min-width">
                                <div class="flex gap-5">
                                    {#if script.trigger_type === "manual"}
                                        <button
                                            type="button"
                                            class="btn btn-sm btn-circle btn-secondary"
                                            use:tooltip={"Execute script"}
                                            on:click={() => executeScript(script)}
                                        >
                                            <i class="ri-play-line" aria-hidden="true" />
                                        </button>
                                    {/if}
                                    <button
                                        type="button"
                                        class="btn btn-sm btn-circle"
                                        use:tooltip={"Edit script"}
                                        on:click={() => upsertPanel?.show(script)}
                                    >
                                        <i class="ri-pencil-line" aria-hidden="true" />
                                    </button>
                                </div>
                            </td>
                        </tr>
                    {/each}

                    {#if !scripts.length && !isLoading}
                        <tr>
                            <td colspan="7" class="txt-center txt-hint p-xs">
                                <h6>No scripts found.</h6>
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
                    {selectedTotal === 1 ? "script" : "scripts"}
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
                    <span class="txt">Load more ({scripts.length}/{totalItems})</span>
                </button>
            </div>
        {/if}
    </main>
</div>

<ScriptUpsertPanel bind:this={upsertPanel} on:save={() => load()} />