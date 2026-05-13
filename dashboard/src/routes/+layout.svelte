<script lang="ts">
  import '../app.css';
  import { page } from '$app/stores';
  let { children } = $props();
</script>

<div class="min-h-screen flex flex-col">
  <header class="w-full max-w-6xl mx-auto px-8 py-5 flex items-center justify-between">
    <a href="/" class="flex items-center gap-2.5 select-none">
      <svg viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg" class="w-7 h-7 flex-shrink-0">
        <polygon points="20,2 36,11 36,29 20,38 4,29 4,11" fill="rgba(30,41,99,0.12)"/>
        <polygon points="20,2 36,11 36,29 20,38 4,29 4,11" stroke="#1e2963" stroke-width="2.2" fill="none"/>
        <path d="M22 7 L14 22 H19.5 L17 33 L27 18 H21 Z" fill="#1e2963"/>
      </svg>
      <span class="font-semibold text-[15px] text-gray-800 tracking-tight">ChaosCI Stats</span>
    </a>

    <nav class="flex items-center gap-1">
      <a href="/"
        class="px-5 py-1.5 rounded-full text-sm font-medium transition-all
          {$page.url.pathname === '/'
            ? 'bg-[#1e1e3a] text-white'
            : 'text-gray-500 hover:text-gray-800'}">
        Home
      </a>
      <a href="/runs"
        class="px-5 py-1.5 rounded-full text-sm font-medium transition-all
          {$page.url.pathname.startsWith('/runs')
            ? 'bg-[#1e1e3a] text-white'
            : 'text-gray-500 hover:text-gray-800'}">
        Pipelines
      </a>
    </nav>
  </header>

  <main class="flex-1 w-full max-w-6xl mx-auto px-8 pb-16">
    {@render children()}
  </main>
</div>
