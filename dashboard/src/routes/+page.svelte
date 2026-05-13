<script lang="ts">
  import { goto } from '$app/navigation';
  import { Search } from 'lucide-svelte';
  import { browser } from '$app/environment';
  import RunChart from '$lib/components/RunChart.svelte';

  let { data } = $props();
  let runId = $state('');

  function handleSubmit(e: Event) {
    e.preventDefault();
    if (runId.trim()) goto(`/runs/${runId.trim()}`);
  }

  let stats       = $derived(data.stats ?? { total_runs: 0, unique_users: 0, ci_minutes: 0, success_rate: 0, runs: [] });
  let runs        = $derived((stats.runs ?? []) as any[]);

  function formatNumber(n: number): string {
    return n.toLocaleString();
  }
</script>

<svelte:head>
  <title>ChaosCI Stats – Home</title>
</svelte:head>

<div class="bg-white/80 backdrop-blur-sm rounded-2xl shadow-sm border border-gray-200/50 overflow-hidden">
  <div class="flex flex-col lg:flex-row">

    <div class="lg:w-[320px] flex-shrink-0 p-8 lg:border-r border-gray-100 space-y-5">
      <div>
        <h1 class="text-[22px] font-bold text-gray-900 leading-snug">ChaosCI Statistics</h1>
        <p class="text-[13px] text-gray-500 mt-2.5 leading-relaxed">
          Statistics about public <span class="underline decoration-dotted cursor-help">ChaosCI pipeline</span> runs to
          help detect overload, excessive usage, project-wide Chaos CI health and potential systemic regressions.
        </p>
      </div>

      <form onsubmit={handleSubmit}>
        <div class="flex items-center rounded-lg border border-gray-200 bg-gray-50/70
                    focus-within:border-gray-400 transition-all">
          <div class="pl-3 text-gray-400 flex-shrink-0">
            <Search class="w-4 h-4" />
          </div>
          <input
            type="text"
            bind:value={runId}
            placeholder="Jump to run ID…"
            class="flex-1 min-w-0 px-2.5 py-2.5 bg-transparent outline-none text-sm text-gray-800
                   placeholder:text-gray-300"
          />
          <button type="submit"
            class="flex-shrink-0 px-4 py-2.5 bg-[#1e1e3a] text-white text-xs font-semibold hover:bg-[#2d2d55]
                   transition-colors rounded-r-lg">
            Go
          </button>
        </div>
      </form>
    </div>

    <div class="flex-1 p-8 flex items-center">
      <div class="w-full grid grid-cols-2 md:grid-cols-4 gap-8">

        <div>
          <p class="text-[11px] font-semibold text-gray-400 uppercase tracking-wider leading-none">Active Users</p>
          <p class="text-[10px] text-gray-400 mt-1">Contributors</p>
          <p class="text-[32px] font-extrabold text-gray-900 tabular-nums leading-tight mt-2">
            {formatNumber(stats.unique_users)}
          </p>
        </div>

        <div>
          <p class="text-[11px] font-semibold text-gray-400 uppercase tracking-wider leading-none">Pipelines</p>
          <p class="text-[10px] text-gray-400 mt-1">Success rate: {stats.success_rate}%</p>
          <p class="text-[32px] font-extrabold text-gray-900 tabular-nums leading-tight mt-2">
            {formatNumber(stats.total_runs)}
          </p>
        </div>

        <div>
          <p class="text-[11px] font-semibold text-gray-400 uppercase tracking-wider leading-none">Runs</p>
          <p class="text-[10px] text-gray-400 mt-1">Total executed</p>
          <p class="text-[32px] font-extrabold text-gray-900 tabular-nums leading-tight mt-2">
            {formatNumber(stats.total_runs)}
          </p>
        </div>

        <div>
          <p class="text-[11px] font-semibold text-gray-400 uppercase tracking-wider leading-none">CI Minutes</p>
          <p class="text-[10px] text-gray-400 mt-1">Total consumed</p>
          <p class="text-[32px] font-extrabold text-gray-900 tabular-nums leading-tight mt-2">
            {formatNumber(stats.ci_minutes ?? 0)}
          </p>
        </div>

      </div>
    </div>
  </div>
</div>

<div class="mt-8">
  <div class="flex items-center justify-between mb-3 px-1">
    <div>
      <h2 class="font-semibold text-gray-800 text-[15px]">Pipeline Stats</h2>
      <p class="text-xs text-gray-500">Pipeline activity over time</p>
    </div>
    <div class="flex items-center gap-5 text-[11px] text-gray-500 font-medium">
      <span class="flex items-center gap-1.5"><span class="w-3 h-[3px] rounded-full bg-rose-400 inline-block"></span>Failed</span>
      <span class="flex items-center gap-1.5"><span class="w-3 h-[3px] rounded-full bg-sky-400 inline-block"></span>Passed</span>
      <span class="flex items-center gap-1.5"><span class="w-3 h-[3px] rounded-full bg-gray-300 inline-block"></span>Total</span>
    </div>
  </div>

  <div class="bg-white/80 backdrop-blur-sm rounded-2xl px-8 py-7 shadow-sm border border-gray-200/50">
    <div class="h-80 w-full">
      {#if browser}
        <RunChart {runs} />
      {:else}
        <div class="h-full flex items-center justify-center text-gray-300 text-sm">Loading chart…</div>
      {/if}
    </div>
  </div>
</div>
