import { BASE_URL, TOKEN_KEY } from "../constants";
import { expireSession } from "./api";

// fetch streaming keeps the bearer token out of URLs and access logs.
export async function subscribeToPosts(query, signal, onPosts, onStatus) {
  const response = await fetch(`${BASE_URL}/events?${query}`, {
    headers: { Authorization: `Bearer ${localStorage.getItem(TOKEN_KEY)}` }, signal,
  });
  if (response.status === 401) { expireSession(); return; }
  if (response.status === 404) { const error = new Error("Live updates unavailable"); error.retryable = false; throw error; }
  if (!response.ok || !response.body || !response.headers.get("content-type")?.includes("text/event-stream")) throw new Error("Live updates unavailable");
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  try {
    while (!signal.aborted) {
      const { value, done } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      let separator;
      while ((separator = buffer.indexOf("\n\n")) >= 0) {
        const event = buffer.slice(0, separator);
        buffer = buffer.slice(separator + 2);
        if (event.includes("event: session-expired")) { expireSession(); return; }
        if (event.includes("event: unavailable")) throw new Error("Live updates unavailable");
        if (event.startsWith(":")) onStatus("live");
        if (event.includes("event: posts")) {
          const data = event.split("\n").filter(line => line.startsWith("data: ")).map(line => line.slice(6)).join("\n");
          const posts = JSON.parse(data);
          if (Array.isArray(posts)) { onPosts(posts); onStatus("live"); }
        }
      }
    }
  } finally { await reader.cancel().catch(() => {}); reader.releaseLock(); }
}
