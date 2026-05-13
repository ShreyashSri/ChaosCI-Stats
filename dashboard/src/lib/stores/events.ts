import { writable } from 'svelte/store';

export type ChaosEvent = {
  id: string;
  run_id: string;
  experiment_name: string;
  status: string;
  message: string;
  timestamp: string;
};

export const createEventStore = (runId: string) => {
  const { subscribe, update, set } = writable<ChaosEvent[]>([]);
  let eventSource: EventSource | null = null;

  const connect = () => {
    if (eventSource) return;

    eventSource = new EventSource(`/api/runs/${runId}/events`);

    eventSource.onmessage = (event) => {
      try {
        const parsed: ChaosEvent = JSON.parse(event.data);
        update((events) => [parsed, ...events]);
      } catch (err) {
        console.error('Failed to parse SSE event', err);
      }
    };

    eventSource.onerror = (err) => {
      console.error('SSE Error:', err);
      eventSource?.close();
    };
  };

  const disconnect = () => {
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }
  };

  return {
    subscribe,
    connect,
    disconnect,
    reset: () => set([])
  };
};
