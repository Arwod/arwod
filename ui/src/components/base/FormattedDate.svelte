<script>
    import tooltip from "@/actions/tooltip";
    import CommonHelper from "@/utils/CommonHelper";

    export let date = "";

    $: localDateTime = date ? CommonHelper.formatToLocalDate(date) : null;
    $: localDateOnly = localDateTime ? localDateTime.substring(0, 10) : null;
    $: localTimeOnly = localDateTime ? localDateTime.substring(11, 19) : null;

    const tooltipData = {
        // generate the tooltip text as getter to speed up the initial load
        // in case the component is used with large number of items
        get text() {
            return date ? date.replace("Z", " UTC") : "";
        },
    };
</script>

{#if date}
    <div class="datetime" use:tooltip={tooltipData}>
        <div class="date">{localDateOnly}</div>
        <div class="time">{localTimeOnly} Local</div>
    </div>
{:else}
    <span class="txt txt-hint">N/A</span>
{/if}

<style>
    .datetime {
        display: inline-block;
        vertical-align: top;
        white-space: nowrap;
        line-height: var(--smLineHeight);
    }
    .time {
        font-size: var(--smFontSize);
        color: var(--txtHintColor);
    }
</style>
