<script>
    import { createEventDispatcher } from "svelte";
    import { slide } from "svelte/transition";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { addSuccessToast, addErrorToast } from "@/stores/toasts";
    import { setErrors } from "@/stores/errors";
    import OverlayPanel from "@/components/base/OverlayPanel.svelte";
    import Field from "@/components/base/Field.svelte";
    import CodeEditor from "@/components/base/CodeEditor.svelte";
    import tooltip from "@/actions/tooltip";

    const dispatch = createEventDispatcher();

    let panel;
    let original = {};
    let script = {};
    let isLoading = false;
    let isSaving = false;
    let showAdvanced = false;

    export function show(scriptData) {
        load(scriptData);
        return panel?.show();
    }

    export function hide() {
        return panel?.hide();
    }

    function load(scriptData) {
        original = scriptData || {};
        script = {
            name: "",
            description: "",
            content: "",
            trigger_type: "manual",
            trigger_config: "{}",
            status: "active",
            timeout: 30,
            tags: [],
            version: "1.0.0",
            is_system: false,
            ...original,
        };

        // 确保 tags 是数组
        if (typeof script.tags === "string") {
            try {
                script.tags = JSON.parse(script.tags);
            } catch {
                script.tags = [];
            }
        }

        // 确保 trigger_config 是字符串
        if (typeof script.trigger_config === "object") {
            script.trigger_config = JSON.stringify(script.trigger_config, null, 2);
        }

        setErrors({});
    }

    async function save() {
        if (isSaving || isLoading) {
            return;
        }

        isSaving = true;
        setErrors({});

        try {
            // 验证 trigger_config JSON 格式
            try {
                JSON.parse(script.trigger_config || "{}");
            } catch {
                throw new Error("Invalid trigger configuration JSON format");
            }

            // 准备保存数据
            const data = {
                ...script,
                tags: Array.isArray(script.tags) ? script.tags : [],
            };

            let result;
            if (original.id) {
                result = await ApiClient.collection("js_scripts").update(original.id, data);
                addSuccessToast(`Successfully updated script "${result.name}".`);
            } else {
                result = await ApiClient.collection("js_scripts").create(data);
                addSuccessToast(`Successfully created script "${result.name}".`);
            }

            dispatch("save", result);
            hide();
        } catch (err) {
            if (!err?.isAbort) {
                console.warn(err);
                addErrorToast(err.data?.message || err.message || "Failed to save script.");
                setErrors(err.data?.data || {});
            }
        }

        isSaving = false;
    }

    function addTag() {
        const tag = prompt("Enter tag name:");
        if (tag && tag.trim() && !script.tags.includes(tag.trim())) {
            script.tags = [...script.tags, tag.trim()];
        }
    }

    function removeTag(index) {
        script.tags = script.tags.filter((_, i) => i !== index);
    }

    function getDefaultTriggerConfig(triggerType) {
        switch (triggerType) {
            case "hook":
                return JSON.stringify({
                    "events": ["records.before_create"],
                    "collections": ["*"]
                }, null, 2);
            case "cron":
                return JSON.stringify({
                    "schedule": "0 0 * * *",
                    "timezone": "UTC"
                }, null, 2);
            default:
                return "{}";
        }
    }

    $: if (script.trigger_type && (!script.trigger_config || script.trigger_config === "{}")) {
        script.trigger_config = getDefaultTriggerConfig(script.trigger_type);
    }
</script>

<OverlayPanel
    bind:this={panel}
    class="overlay-panel-lg script-panel"
    beforeHide={() => !isSaving}
    on:hide
    on:show
>
    <svelte:fragment slot="header">
        <h4>
            {original.id ? "Edit" : "New"} script
            {#if original.id}
                <span class="txt-hint">({original.name})</span>
            {/if}
        </h4>
    </svelte:fragment>

    <div class="grid">
        <div class="col-lg-6">
            <Field class="form-field required" name="name" let:uniqueId>
                <label for={uniqueId}>Name</label>
                <input
                    type="text"
                    id={uniqueId}
                    bind:value={script.name}
                    required
                    placeholder="Script name"
                />
            </Field>
        </div>

        <div class="col-lg-6">
            <Field class="form-field" name="version" let:uniqueId>
                <label for={uniqueId}>Version</label>
                <input
                    type="text"
                    id={uniqueId}
                    bind:value={script.version}
                    placeholder="1.0.0"
                />
            </Field>
        </div>

        <div class="col-lg-12">
            <Field class="form-field" name="description" let:uniqueId>
                <label for={uniqueId}>Description</label>
                <textarea
                    id={uniqueId}
                    bind:value={script.description}
                    placeholder="Script description"
                    rows="2"
                />
            </Field>
        </div>

        <div class="col-lg-6">
            <Field class="form-field required" name="trigger_type" let:uniqueId>
                <label for={uniqueId}>Trigger Type</label>
                <select id={uniqueId} bind:value={script.trigger_type} required>
                    <option value="manual">Manual</option>
                    <option value="hook">Hook</option>
                    <option value="cron">Scheduled (Cron)</option>
                </select>
            </Field>
        </div>

        <div class="col-lg-6">
            <Field class="form-field" name="status" let:uniqueId>
                <label for={uniqueId}>Status</label>
                <select id={uniqueId} bind:value={script.status}>
                    <option value="active">Active</option>
                    <option value="inactive">Inactive</option>
                </select>
            </Field>
        </div>

        <div class="col-lg-12">
            <Field class="form-field required" name="content" let:uniqueId>
                <label for={uniqueId}>Script Content</label>
                <CodeEditor
                    bind:value={script.content}
                    language="javascript"
                    placeholder="// Write your JavaScript code here\nconsole.log('Hello, PocketBase!');"
                />
            </Field>
        </div>

        <div class="col-lg-12">
            <button
                type="button"
                class="btn btn-sm btn-secondary btn-outline"
                on:click={() => showAdvanced = !showAdvanced}
            >
                <i class="ri-settings-3-line" aria-hidden="true" />
                <span class="txt">{showAdvanced ? "Hide" : "Show"} Advanced Settings</span>
            </button>
        </div>

        {#if showAdvanced}
            <div class="col-lg-12" transition:slide>
                <div class="grid">
                    <div class="col-lg-6">
                        <Field class="form-field" name="timeout" let:uniqueId>
                            <label for={uniqueId}>
                                Timeout (seconds)
                                <i
                                    class="ri-information-line txt-sm"
                                    use:tooltip={"Maximum execution time in seconds"}
                                    aria-hidden="true"
                                />
                            </label>
                            <input
                                type="number"
                                id={uniqueId}
                                bind:value={script.timeout}
                                min="1"
                                max="300"
                                placeholder="30"
                            />
                        </Field>
                    </div>

                    <div class="col-lg-6">
                        <Field class="form-field" name="is_system" let:uniqueId>
                            <label for={uniqueId}>
                                System Script
                                <i
                                    class="ri-information-line txt-sm"
                                    use:tooltip={"System scripts cannot be deleted"}
                                    aria-hidden="true"
                                />
                            </label>
                            <label class="form-field form-field-toggle">
                                <input
                                    type="checkbox"
                                    bind:checked={script.is_system}
                                />
                                <i class="ri-toggle-line" aria-hidden="true" />
                            </label>
                        </Field>
                    </div>

                    <div class="col-lg-12">
                        <Field class="form-field" name="trigger_config" let:uniqueId>
                            <label for={uniqueId}>
                                Trigger Configuration
                                <i
                                    class="ri-information-line txt-sm"
                                    use:tooltip={"JSON configuration for the trigger type"}
                                    aria-hidden="true"
                                />
                            </label>
                            <CodeEditor
                                bind:value={script.trigger_config}
                                language="json"
                                placeholder="{{}}"
                            />
                        </Field>
                    </div>

                    <div class="col-lg-12">
                        <Field class="form-field" name="tags" let:uniqueId>
                            <label for={uniqueId}>Tags</label>
                            <div class="tags-wrapper">
                                {#each script.tags as tag, i}
                                    <span class="label label-sm">
                                        {tag}
                                        <button
                                            type="button"
                                            class="btn btn-xs btn-circle btn-transparent"
                                            on:click={() => removeTag(i)}
                                        >
                                            <i class="ri-close-line" aria-hidden="true" />
                                        </button>
                                    </span>
                                {/each}
                                <button
                                    type="button"
                                    class="btn btn-xs btn-outline"
                                    on:click={addTag}
                                >
                                    <i class="ri-add-line" aria-hidden="true" />
                                    Add tag
                                </button>
                            </div>
                        </Field>
                    </div>
                </div>
            </div>
        {/if}
    </div>

    <svelte:fragment slot="footer">
        <button type="button" class="btn btn-secondary" disabled={isSaving} on:click={() => hide()}>
            <span class="txt">Cancel</span>
        </button>
        <button
            type="button"
            class="btn btn-expanded"
            class:btn-loading={isSaving}
            disabled={isSaving || !script.name?.trim() || !script.content?.trim()}
            on:click={() => save()}
        >
            <span class="txt">{original.id ? "Update" : "Create"} script</span>
        </button>
    </svelte:fragment>
</OverlayPanel>

<style>
    .tags-wrapper {
        display: flex;
        flex-wrap: wrap;
        gap: 5px;
        align-items: center;
    }

    .script-panel :global(.overlay-panel-content) {
        max-height: 90vh;
        overflow-y: auto;
    }
</style>