export const load = async ({ fetch }) => {
  try {
    const res = await fetch('/api/runs');
    if (res.ok) {
      const runs = await res.json();
      return { runs };
    }
  } catch (e) {
    console.error('Failed to load runs', e);
  }
  return { runs: [] };
};
