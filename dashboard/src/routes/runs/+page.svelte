<script lang="ts">
  import { ShieldCheck, ShieldAlert, RefreshCw, Layers } from 'lucide-svelte';

  let { data } = $props();
  let runs = $derived(data.runs || []);
</script>

<svelte:head>
  <title>ChaosCI Stats - All Runs</title>
</svelte:head>

<div class="space-y-6 max-w-5xl mx-auto">
  <div class="flex items-center gap-3 mb-8">
    <div class="bg-primary/10 p-3 rounded-xl text-primary">
      <Layers class="w-6 h-6" />
    </div>
    <h1 class="text-3xl font-bold text-gray-900 tracking-tight">Recent Runs</h1>
  </div>

  {#if runs.length === 0}
    <div class="bg-white/80 backdrop-blur-xl rounded-3xl p-12 shadow-sm border border-white text-center">
      <p class="text-gray-500">No runs found. Trigger a webhook to start a chaos experiment!</p>
    </div>
  {:else}
    <div class="bg-white/80 backdrop-blur-xl rounded-3xl shadow-sm border border-white overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm text-gray-600">
          <thead class="bg-gray-50/50 text-xs uppercase font-semibold text-gray-500 border-b border-gray-100">
            <tr>
              <th scope="col" class="px-6 py-4">Run ID</th>
              <th scope="col" class="px-6 py-4">Repository</th>
              <th scope="col" class="px-6 py-4">PR #</th>
              <th scope="col" class="px-6 py-4">Engine</th>
              <th scope="col" class="px-6 py-4">Status</th>
              <th scope="col" class="px-6 py-4">Date</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            {#each runs as run}
              <tr class="hover:bg-gray-50/50 transition-colors cursor-pointer" onclick={() => window.location.href = `/runs/${run.id}`}>
                <td class="px-6 py-4 font-mono font-medium text-gray-900">
                  <a href={`/runs/${run.id}`} class="hover:text-primary transition-colors">
                    {run.id}
                  </a>
                </td>
                <td class="px-6 py-4">{run.repo}</td>
                <td class="px-6 py-4">#{run.pr_number}</td>
                <td class="px-6 py-4 capitalize">{run.engine}</td>
                <td class="px-6 py-4">
                  <div class="flex items-center gap-2">
                    {#if run.status === 'success' || run.status === 'completed'}
                      <ShieldCheck class="w-4 h-4 text-green-500" />
                      <span class="font-medium text-green-700 capitalize">{run.status}</span>
                    {:else if run.status === 'failed'}
                      <ShieldAlert class="w-4 h-4 text-red-500" />
                      <span class="font-medium text-red-700 capitalize">{run.status}</span>
                    {:else}
                      <RefreshCw class="w-4 h-4 text-blue-500 animate-spin" />
                      <span class="font-medium text-blue-700 capitalize">{run.status || 'running'}</span>
                    {/if}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  {new Date(run.created_at).toLocaleDateString()}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</div>
