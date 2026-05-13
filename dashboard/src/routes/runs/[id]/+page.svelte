<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { ShieldCheck, ShieldAlert, Clock, Activity, TerminalSquare, RefreshCw } from 'lucide-svelte';
  import type { ChaosEvent } from '$lib/stores/events';

  let { data } = $props();
  
  let events: ChaosEvent[] = $state([]);
  let eventSource: EventSource | null = null;

  onMount(() => {
    if (data.run && data.run.id) {
      eventSource = new EventSource(`/api/runs/${data.run.id}/events`);
      eventSource.onmessage = (event) => {
        try {
          const parsed = JSON.parse(event.data);
          events = [parsed, ...events];
        } catch (err) {
          console.error(err);
        }
      };
      eventSource.onerror = () => {
        eventSource?.close();
      };
    }
  });

  onDestroy(() => {
    if (eventSource) {
      eventSource.close();
    }
  });

  let run = $derived(data.run);
  let error = $derived(data.error as any);
</script>

<svelte:head>
  <title>Run Details {run?.id ? `- ${run.id}` : ''}</title>
</svelte:head>

{#if error}
  <div class="flex items-center justify-center min-h-[50vh]">
    <div class="bg-white rounded-2xl p-8 shadow-sm border border-red-100 max-w-md w-full text-center space-y-4">
      <div class="flex justify-center text-red-500 mb-2">
        <ShieldAlert class="w-12 h-12" />
      </div>
      <h2 class="text-2xl font-bold text-gray-900">Run Not Found</h2>
      <p class="text-gray-500">{error.message || 'The requested run could not be found or an error occurred.'}</p>
      <a href="/" class="inline-block mt-4 px-6 py-2 bg-gray-100 hover:bg-gray-200 text-gray-800 rounded-full font-medium transition-colors">
        Back to Home
      </a>
    </div>
  </div>
{:else if run}
  <div class="space-y-6">
    <div class="bg-white rounded-3xl p-8 shadow-sm border border-gray-100 flex flex-col md:flex-row justify-between items-start md:items-center gap-6">
      <div class="space-y-2">
        <div class="flex items-center gap-3">
          <h1 class="text-2xl font-bold text-gray-900 tracking-tight">Run Details</h1>
          <span class="px-3 py-1 bg-gray-100 text-gray-700 text-xs font-semibold rounded-full font-mono border border-gray-200">
            {run.id}
          </span>
        </div>
        <p class="text-sm text-gray-500 flex items-center gap-2">
          <Clock class="w-4 h-4" />
          Started at {new Date(run.created_at).toLocaleString()}
        </p>
      </div>

      <div class="flex items-center gap-6">
        <div class="text-center">
          <p class="text-sm text-gray-500 mb-1">Status</p>
          <div class="flex items-center gap-2">
            {#if run.status === 'success' || run.status === 'completed'}
              <ShieldCheck class="w-5 h-5 text-green-500" />
              <span class="font-semibold text-gray-900 capitalize">{run.status}</span>
            {:else if run.status === 'failed'}
              <ShieldAlert class="w-5 h-5 text-red-500" />
              <span class="font-semibold text-gray-900 capitalize">{run.status}</span>
            {:else}
              <RefreshCw class="w-5 h-5 text-blue-500 animate-spin" />
              <span class="font-semibold text-gray-900 capitalize">{run.status || 'Running'}</span>
            {/if}
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">

      <div class="lg:col-span-1 space-y-6">
        <div class="bg-white rounded-3xl p-6 shadow-sm border border-gray-100">
          <div class="flex items-center gap-2 mb-4 pb-4 border-b border-gray-100">
            <Activity class="w-5 h-5 text-primary" />
            <h3 class="font-semibold text-gray-900">Recent Activity</h3>
          </div>
          
          {#if events.length === 0}
            <p class="text-sm text-gray-500 italic">Waiting for events...</p>
          {:else}
            <div class="space-y-4">
              {#each events.slice(0, 5) as event}
                <div class="flex items-start gap-3">
                  <div class="mt-1">
                    {#if event.status === 'success' || event.status === 'completed'}
                      <div class="w-2 h-2 rounded-full bg-green-500"></div>
                    {:else if event.status === 'failed'}
                      <div class="w-2 h-2 rounded-full bg-red-500"></div>
                    {:else}
                      <div class="w-2 h-2 rounded-full bg-blue-500 animate-pulse"></div>
                    {/if}
                  </div>
                  <div>
                    <p class="text-sm font-medium text-gray-900">{event.experiment_name}</p>
                    <p class="text-xs text-gray-500">{new Date(event.timestamp).toLocaleTimeString()}</p>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <div class="lg:col-span-2">
        <div class="bg-white rounded-3xl overflow-hidden shadow-sm border border-gray-100 h-[600px] flex flex-col">
          <div class="bg-gray-50 border-b border-gray-100 px-6 py-4 flex items-center justify-between">
            <div class="flex items-center gap-2 text-gray-700">
              <TerminalSquare class="w-5 h-5" />
              <h3 class="font-semibold">Live Event Stream</h3>
            </div>
            <div class="flex items-center gap-2">
              <span class="flex h-2 w-2 relative">
                <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                <span class="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
              </span>
              <span class="text-xs font-medium text-gray-500 uppercase tracking-wider">Live</span>
            </div>
          </div>
          
          <div class="flex-1 overflow-y-auto p-6 bg-gray-900 font-mono text-sm space-y-3">
            {#if events.length === 0}
              <div class="text-gray-500 italic">Connecting to event stream...</div>
            {/if}
            
            {#each events as event}
              <div class="border-l-2 pl-3 {event.status === 'failed' ? 'border-red-500' : event.status === 'success' ? 'border-green-500' : 'border-blue-500'}">
                <div class="flex items-center gap-2 text-gray-500 text-xs mb-1">
                  <span>{new Date(event.timestamp).toISOString()}</span>
                  <span class="uppercase font-semibold {event.status === 'failed' ? 'text-red-400' : event.status === 'success' ? 'text-green-400' : 'text-blue-400'}">
                    [{event.status}]
                  </span>
                  <span class="text-gray-400">{event.experiment_name}</span>
                </div>
                <div class="text-gray-300">
                  {event.message}
                </div>
              </div>
            {/each}
          </div>
        </div>
      </div>
      
    </div>
  </div>
{/if}
