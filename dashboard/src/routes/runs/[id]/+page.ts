export const load = async ({ params, fetch }) => {
  const { id } = params;
  
  try {
    const res = await fetch(`/api/runs/${id}`);
    
    if (!res.ok) {
      return {
        status: res.status,
        error: new Error(`Failed to fetch run: ${res.statusText}`)
      };
    }
    
    const run = await res.json();
    return {
      run
    };
  } catch (e) {
    return {
      status: 500,
      error: e
    };
  }
};
