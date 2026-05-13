<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  interface Run {
    id: string;
    status: string;
    created_at: string;
  }

  let { runs = [] as Run[] } = $props();
  let canvas: HTMLCanvasElement;
  let chart: any = null;

  onMount(async () => {
    const { Chart, LineElement, LinearScale, CategoryScale, PointElement, Filler, Tooltip } = await import('chart.js');
    Chart.register(LineElement, LinearScale, CategoryScale, PointElement, Filler, Tooltip);

    const counts: Record<string, { total: number; passed: number; failed: number }> = {};
    for (const run of runs) {
      const d = new Date(run.created_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
      if (!counts[d]) counts[d] = { total: 0, passed: 0, failed: 0 };
      counts[d].total++;
      if (run.status === 'success' || run.status === 'completed') counts[d].passed++;
      if (run.status === 'failed') counts[d].failed++;
    }

    const labels = Object.keys(counts);
    if (labels.length === 0) labels.push(new Date().toLocaleDateString('en-US', { month: 'short', day: 'numeric' }));

    const total  = labels.map(l => counts[l]?.total  ?? 0);
    const passed = labels.map(l => counts[l]?.passed ?? 0);
    const failed = labels.map(l => counts[l]?.failed ?? 0);

    chart = new Chart(canvas, {
      type: 'line',
      data: {
        labels,
        datasets: [
          {
            label: 'Total',
            data: total,
            borderColor: 'rgba(148,163,184,0.9)',
            backgroundColor: 'rgba(148,163,184,0.15)',
            fill: true,
            tension: 0.4,
            borderWidth: 2,
            pointRadius: 2,
          },
          {
            label: 'Passed',
            data: passed,
            borderColor: 'rgba(14,165,233,0.9)',
            backgroundColor: 'rgba(14,165,233,0.12)',
            fill: true,
            tension: 0.4,
            borderWidth: 2,
            pointRadius: 2,
          },
          {
            label: 'Failed',
            data: failed,
            borderColor: 'rgba(251,113,133,0.9)',
            backgroundColor: 'rgba(251,113,133,0.1)',
            fill: true,
            tension: 0.4,
            borderWidth: 2,
            pointRadius: 2,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { mode: 'index', intersect: false },
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              label: (ctx) => ` ${ctx.dataset.label}: ${ctx.parsed.y}`
            }
          }
        },
        scales: {
          x: {
            grid: { color: 'rgba(147,197,253,0.2)' },
            ticks: { font: { size: 11 }, color: '#94a3b8' },
          },
          y: {
            beginAtZero: true,
            ticks: { precision: 0, font: { size: 11 }, color: '#94a3b8' },
            grid: { color: 'rgba(147,197,253,0.2)' },
          }
        }
      }
    });
  });

  onDestroy(() => {
    chart?.destroy();
  });
</script>

<canvas bind:this={canvas} class="w-full h-full"></canvas>
