<script>
    import { link } from "svelte-spa-router";
    import active from "svelte-spa-router/active";
    import PageSidebar from "@/components/base/PageSidebar.svelte";
    import tooltip from "@/actions/tooltip";

    let sidebarWidth = 220; // default sidebar width
</script>

<PageSidebar bind:width={sidebarWidth}>
    <nav class="sidebar-content">
        <a
            href="/scripts"
            class="sidebar-item"
            use:link
            use:active={{ path: "/scripts", className: "current-route" }}
        >
            <i class="ri-code-s-slash-line sidebar-item-icon" aria-hidden="true" />
            <span class="sidebar-item-txt">Scripts</span>
        </a>

        <a
            href="/scripts/logs"
            class="sidebar-item"
            use:link
            use:active={{ path: "/scripts/logs", className: "current-route" }}
        >
            <i class="ri-file-list-3-line sidebar-item-icon" aria-hidden="true" />
            <span class="sidebar-item-txt">Execution Logs</span>
        </a>

        <a
            href="/scripts/monitor"
            class="sidebar-item"
            use:link
            use:active={{ path: "/scripts/monitor", className: "current-route" }}
        >
            <i class="ri-dashboard-line sidebar-item-icon" aria-hidden="true" />
            <span class="sidebar-item-txt">Performance Monitor</span>
        </a>

        <hr class="sidebar-divider" />

        <div class="sidebar-title">
            <span class="txt">Quick Actions</span>
        </div>

        <button
            type="button"
            class="sidebar-item sidebar-item-btn"
            use:tooltip={{ text: "Create new script", position: "right" }}
            on:click={() => {
                // 触发新建脚本事件
                window.dispatchEvent(new CustomEvent('create-script'));
            }}
        >
            <i class="ri-add-line sidebar-item-icon" aria-hidden="true" />
            <span class="sidebar-item-txt">New Script</span>
        </button>

        <button
            type="button"
            class="sidebar-item sidebar-item-btn"
            use:tooltip={{ text: "Import script from file", position: "right" }}
            on:click={() => {
                // 触发导入脚本事件
                window.dispatchEvent(new CustomEvent('import-script'));
            }}
        >
            <i class="ri-upload-line sidebar-item-icon" aria-hidden="true" />
            <span class="sidebar-item-txt">Import Script</span>
        </button>

        <hr class="sidebar-divider" />

        <div class="sidebar-title">
            <span class="txt">Script Types</span>
        </div>

        <a
            href="/scripts?filter=trigger_type='manual'"
            class="sidebar-item"
            use:link
        >
            <i class="ri-play-circle-line sidebar-item-icon" aria-hidden="true" />
            <span class="sidebar-item-txt">Manual Scripts</span>
        </a>

        <a
            href="/scripts?filter=trigger_type='hook'"
            class="sidebar-item"
            use:link
        >
            <i class="ri-git-branch-line sidebar-item-icon" aria-hidden="true" />
            <span class="sidebar-item-txt">Hook Scripts</span>
        </a>

        <a
            href="/scripts?filter=trigger_type='cron'"
            class="sidebar-item"
            use:link
        >
            <i class="ri-time-line sidebar-item-icon" aria-hidden="true" />
            <span class="sidebar-item-txt">Scheduled Scripts</span>
        </a>

        <hr class="sidebar-divider" />

        <div class="sidebar-title">
            <span class="txt">Status</span>
        </div>

        <a
            href="/scripts?filter=status='active'"
            class="sidebar-item"
            use:link
        >
            <i class="ri-checkbox-circle-line sidebar-item-icon txt-success" aria-hidden="true" />
            <span class="sidebar-item-txt">Active Scripts</span>
        </a>

        <a
            href="/scripts?filter=status='inactive'"
            class="sidebar-item"
            use:link
        >
            <i class="ri-pause-circle-line sidebar-item-icon txt-warning" aria-hidden="true" />
            <span class="sidebar-item-txt">Inactive Scripts</span>
        </a>

        <a
            href="/scripts?filter=is_system=true"
            class="sidebar-item"
            use:link
        >
            <i class="ri-shield-check-line sidebar-item-icon txt-info" aria-hidden="true" />
            <span class="sidebar-item-txt">System Scripts</span>
        </a>
    </nav>
</PageSidebar>

<style>
    .sidebar-divider {
        margin: 10px 0;
        border: none;
        border-top: 1px solid var(--baseAlt2Color);
    }

    .sidebar-title {
        padding: 5px 15px;
        font-size: 11px;
        font-weight: 600;
        text-transform: uppercase;
        color: var(--txtHintColor);
        letter-spacing: 0.5px;
    }

    .sidebar-item-btn {
        background: none;
        border: none;
        width: 100%;
        text-align: left;
        cursor: pointer;
        transition: background-color 0.2s;
    }

    .sidebar-item-btn:hover {
        background-color: var(--baseAlt1Color);
    }
</style>