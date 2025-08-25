<script>
    import { onMount } from "svelte";
    import ApiClient from "@/utils/ApiClient";
    import tooltip from "@/actions/tooltip";
    import { pageTitle } from "@/stores/app";
    import { addSuccessToast, addErrorToast } from "@/stores/toasts";
    import PageWrapper from "@/components/base/PageWrapper.svelte";
    import RefreshButton from "@/components/base/RefreshButton.svelte";
    import SettingsSidebar from "@/components/settings/SettingsSidebar.svelte";
    import OverlayPanel from "@/components/base/OverlayPanel.svelte";
    import Field from "@/components/base/Field.svelte";
    import CommonHelper from "@/utils/CommonHelper";

    $pageTitle = "Crons";

    let jobs = [];
    let isLoading = false;
    let isRunning = {};
    let showJobModal = false;
    let editingJob = null;
    let jobForm = {
        name: "",
        cron: "",
        status: "1",
        service: "",
        script: "",
        remark: "",
    };
    let isSubmitting = false;
    let codeEditorComponent;

    onMount(async () => {
        try {
            codeEditorComponent = (await import("@/components/base/CodeEditor.svelte")).default;
        } catch (err) {
            console.warn(err);
        }
    });

    loadJobs();

    async function loadJobs() {
        isLoading = true;

        try {
            jobs = await ApiClient.collection("_jobs").getFullList({
                sort: "created",
            });
            isLoading = false;
        } catch (err) {
            if (!err.isAbort) {
                ApiClient.error(err);
                isLoading = false;
            }
        }
    }

    async function cronRun(jobName) {
        isRunning[jobName] = true;

        try {
            await ApiClient.crons.run(jobName);
            addSuccessToast(`Successfully triggered ${jobName}.`);
            isRunning[jobName] = false;
        } catch (err) {
            if (!err.isAbort) {
                ApiClient.error(err);
                isRunning[jobName] = false;
            }
        }
    }

    function openCreateModal() {
        editingJob = null;
        jobForm = {
            name: "",
            cron: "",
            status: "1",
            service: "",
            script: "",
            remark: "",
        };
        showJobModal = true;
    }

    function openEditModal(job) {
        if (isSystemJob(job)) {
            addErrorToast("System jobs cannot be edited.");
            return;
        }
        editingJob = job;
        jobForm = {
            name: job.name,
            cron: job.cron,
            status: job.status,
            service: job.service || "",
            script: job.script || "",
            remark: job.remark || "",
        };
        showJobModal = true;
    }

    function closeModal() {
        showJobModal = false;
        editingJob = null;
        jobForm = {
            name: "",
            cron: "",
            status: "1",
            service: "",
            script: "",
            remark: "",
        };
    }

    async function saveJob() {
        if (isSubmitting) return;

        isSubmitting = true;

        try {
            if (editingJob) {
                // 更新现有任务
                await ApiClient.collection("_jobs").update(editingJob.id, jobForm);
                addSuccessToast("Job updated successfully.");
            } else {
                // 创建新任务
                await ApiClient.collection("_jobs").create(jobForm);
                addSuccessToast("Job created successfully.");
            }

            closeModal();
            await loadJobs();
        } catch (err) {
            ApiClient.error(err);
        } finally {
            isSubmitting = false;
        }
    }

    async function deleteJob(job) {
        if (job.service === "system") {
            addErrorToast("System jobs cannot be deleted.");
            return;
        }

        if (!confirm(`Are you sure you want to delete the job "${job.name}"?`)) {
            return;
        }

        try {
            await ApiClient.collection("_jobs").delete(job.id);
            addSuccessToast("Job deleted successfully.");
            await loadJobs();
        } catch (err) {
            ApiClient.error(err);
        }
    }

    async function toggleJobStatus(job) {
        if (isSystemJob(job)) {
            addErrorToast("System jobs status cannot be modified.");
            return;
        }
        try {
            const newStatus = job.status === "1" ? "0" : "1";
            await ApiClient.collection("_jobs").update(job.id, { status: newStatus });
            addSuccessToast(`Job ${newStatus === "1" ? "enabled" : "disabled"} successfully.`);
            await loadJobs();
        } catch (err) {
            ApiClient.error(err);
        }
    }

    function isSystemJob(job) {
        return job.service === "system";
    }

    function getStatusText(status) {
        return status === "1" ? "启用" : "禁用";
    }

    function getStatusClass(status) {
        return status === "1" ? "txt-success" : "txt-hint";
    }
</script>

<SettingsSidebar />

<PageWrapper>
    <header class="page-header">
        <nav class="breadcrumbs">
            <div class="breadcrumb-item">Settings</div>
            <div class="breadcrumb-item">{$pageTitle}</div>
        </nav>
    </header>

    <div class="wrapper" style="max-width: none; width: 100%;">
        <div class="panel" autocomplete="off" style="max-width: none; width: 100%;">
            <div class="flex m-b-sm flex-gap-10">
                <span class="txt-xl">Cron Jobs Management</span>
                <div class="flex flex-gap-5">
                    <button type="button" class="btn btn-sm btn-outline" on:click={openCreateModal}>
                        <i class="ri-add-line"></i>
                        <span class="txt">New Job</span>
                    </button>
                    <RefreshButton class="btn-sm" tooltip={"Refresh"} on:refresh={loadJobs} />
                </div>
            </div>

            <div class="table-wrapper">
                <table class="table">
                    <thead>
                        <tr>
                            <th>Name</th>
                            <th>Cron Expression</th>
                            <th>Status</th>
                            <th>Remark</th>
                            <th class="col-type-text col-sm">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#if isLoading}
                            <tr>
                                <td colspan="5" class="txt-center p-lg">
                                    <span class="loader"></span>
                                </td>
                            </tr>
                        {:else}
                            {#each jobs as job (job.id)}
                                <tr>
                                    <td>
                                        <div class="flex flex-nowrap">
                                            <span class="txt">{job.name}</span>
                                            {#if isSystemJob(job)}
                                                <i
                                                    class="ri-settings-3-line txt-hint m-l-5"
                                                    use:tooltip={"System Job"}
                                                ></i>
                                            {/if}
                                        </div>
                                    </td>
                                    <td>
                                        <span class="txt-mono">{job.cron}</span>
                                    </td>
                                    <td>
                                        <div class="form-field form-field-toggle">
                                            <input
                                                type="checkbox"
                                                id="status-{job.id}"
                                                checked={job.status === "1"}
                                                disabled={isSystemJob(job)}
                                                on:change={() => !isSystemJob(job) && toggleJobStatus(job)}
                                            />
                                            <label for="status-{job.id}"></label>
                                        </div>
                                    </td>
                                    <td>
                                        <span class="txt-hint">{job.remark || "-"}</span>
                                    </td>
                                    <td>
                                        <div class="flex flex-nowrap flex-gap-5">
                                            <!-- Run button -->
                                            <button
                                                type="button"
                                                class="btn btn-sm btn-circle btn-hint btn-transparent"
                                                class:btn-loading={isRunning[job.name]}
                                                disabled={isRunning[job.name] || job.status === "0"}
                                                aria-label="Run"
                                                use:tooltip={"Run"}
                                                on:click|preventDefault={() => cronRun(job.name)}
                                            >
                                                <i class="ri-play-large-line"></i>
                                            </button>

                                            <!-- Edit button (only for non-system jobs) -->
                                            {#if !isSystemJob(job)}
                                                <button
                                                    type="button"
                                                    class="btn btn-sm btn-circle btn-hint btn-transparent"
                                                    aria-label="Edit"
                                                    use:tooltip={"Edit"}
                                                    on:click|preventDefault={() => openEditModal(job)}
                                                >
                                                    <i class="ri-pencil-line"></i>
                                                </button>
                                            {/if}

                                            <!-- Delete button (only for non-system jobs) -->
                                            {#if !isSystemJob(job)}
                                                <button
                                                    type="button"
                                                    class="btn btn-sm btn-circle btn-hint btn-transparent"
                                                    aria-label="Delete"
                                                    use:tooltip={"Delete"}
                                                    on:click|preventDefault={() => deleteJob(job)}
                                                >
                                                    <i class="ri-delete-bin-line"></i>
                                                </button>
                                            {/if}
                                        </div>
                                    </td>
                                </tr>
                            {:else}
                                <tr>
                                    <td colspan="5" class="txt-center txt-hint p-lg">
                                        No cron jobs found.
                                    </td>
                                </tr>
                            {/each}
                        {/if}
                    </tbody>
                </table>
            </div>

            <p class="txt-hint m-t-xs">
                System cron jobs are automatically registered and cannot be edited, disabled, or deleted.
                Custom jobs can be created, edited, and deleted as needed.
            </p>
        </div>
    </div>
</PageWrapper>

<!-- Job Modal -->
<OverlayPanel class="overlay-panel-lg" bind:active={showJobModal} on:hide={closeModal}>
    <svelte:fragment slot="header">
        <h4>{editingJob ? "Edit Job" : "Create Job"}</h4>
    </svelte:fragment>

    <form on:submit|preventDefault={saveJob}>
        <Field class="form-field required" name="name" let:uniqueId>
            <label for={uniqueId}>Name</label>
            <input
                type="text"
                id={uniqueId}
                required
                disabled={editingJob && isSystemJob(editingJob)}
                bind:value={jobForm.name}
            />
        </Field>

        <Field class="form-field required" name="cron" let:uniqueId>
            <label for={uniqueId}>Cron Expression</label>
            <input type="text" id={uniqueId} required placeholder="0 0 * * *" bind:value={jobForm.cron} />
            <div class="help-block">Use standard cron format: minute hour day month weekday</div>
        </Field>

        <Field class="form-field" name="status" let:uniqueId>
            <label for={uniqueId}>Status</label>
            <select id={uniqueId} bind:value={jobForm.status}>
                <option value="1">启用</option>
                <option value="0">禁用</option>
            </select>
        </Field>

        <!-- <Field class="form-field" name="service" let:uniqueId>
            <label for={uniqueId}>Service</label>
            <input
                type="text"
                id={uniqueId}
                disabled={editingJob && isSystemJob(editingJob)}
                bind:value={jobForm.service}
            />
            <div class="help-block">Service identifier (optional)</div>
        </Field> -->

        <Field class="form-field" name="script" let:uniqueId>
            <label for={uniqueId}>Script</label>
            {#if codeEditorComponent}
                <svelte:component
                    this={codeEditorComponent}
                    id={uniqueId}
                    bind:value={jobForm.script}
                    language="javascript"
                    placeholder="// Enter your JavaScript code here..."
                    minHeight={120}
                    maxHeight={500}
                    disabled={editingJob && isSystemJob(editingJob)}
                />
            {:else}
                <textarea
                    id={uniqueId}
                    rows="4"
                    bind:value={jobForm.script}
                    disabled={editingJob && isSystemJob(editingJob)}
                ></textarea>
            {/if}
            <div class="help-block">JavaScript code to execute (optional)</div>
        </Field>

        <Field class="form-field" name="remark" let:uniqueId>
            <label for={uniqueId}>Remark</label>
            <input type="text" id={uniqueId} bind:value={jobForm.remark} />
            <div class="help-block">Description or notes about this job (optional)</div>
        </Field>
    </form>

    <svelte:fragment slot="footer">
        <button type="button" class="btn btn-transparent" on:click={closeModal}>
            <span class="txt">Cancel</span>
        </button>
        <button
            type="button"
            class="btn btn-expanded"
            class:btn-loading={isSubmitting}
            disabled={isSubmitting || !jobForm.name || !jobForm.cron}
            on:click={saveJob}
        >
            <span class="txt">{editingJob ? "Update" : "Create"}</span>
        </button>
    </svelte:fragment>
</OverlayPanel>

<style>
    .wrapper {
        max-width: none !important;
        width: 100% !important;
        padding: 0 20px !important;
    }

    .panel {
        max-width: none !important;
        width: 100% !important;
        margin: 0 !important;
    }

    .table-wrapper {
        overflow-x: auto;
        width: 100%;
        margin: 0;
    }

    .table {
        width: 100%;
        table-layout: fixed;
        margin: 0;
    }

    .table th,
    .table td {
        text-align: left;
        vertical-align: middle;
        padding: 12px 15px;
        border-bottom: 1px solid var(--baseAlt2Color);
        word-wrap: break-word;
    }

    .table th {
        background: var(--baseAlt1Color);
        font-weight: 600;
        font-size: 0.85rem;
        color: var(--txtHintColor);
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .table tbody tr:hover {
        background: var(--baseAlt1Color);
    }

    .table th:nth-child(1) {
        width: 25%;
    } /* Name */
    .table th:nth-child(2) {
        width: 25%;
    } /* Cron Expression */
    .table th:nth-child(3) {
        width: 12%;
    } /* Status */
    .table th:nth-child(4) {
        width: 23%;
    } /* Remark */
    .table th:nth-child(5) {
        width: 15%;
    } /* Actions */

    /* Switch组件在表格中的样式优化 */
    .table .form-field-toggle {
        margin: 0;
        display: flex;
        justify-content: center;
        align-items: center;
    }

    .table .form-field-toggle input[type="checkbox"] ~ label {
        margin: 0;
        min-height: auto;
    }

    .col-sm {
        width: 120px;
    }
</style>
