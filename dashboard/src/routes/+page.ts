export const load = async ({ fetch }) => {
  try {
    const res = await fetch('/api/stats');
    if (res.ok) {
      const stats = await res.json();
      return { stats };
    }
  } catch (e) {
    console.error('Failed to load stats', e);
  }
  return { stats: { total_runs: 0, unique_users: 0, runs: [] } };
};
