import axios from "axios";
import { BASE_URL, TOKEN_KEY } from "../constants";

export function session() {
  try {
    const token = localStorage.getItem(TOKEN_KEY);
    if (!token) return null;
    const payload = token.split(".")[1].replace(/-/g, "+").replace(/_/g, "/");
    const claims = JSON.parse(decodeURIComponent(Array.from(atob(payload), c => `%${c.charCodeAt(0).toString(16).padStart(2, "0")}`).join("")));
    return claims.username && claims.exp * 1000 > Date.now() ? claims : null;
  } catch { return null; }
}
export function expireSession() {
  localStorage.removeItem(TOKEN_KEY);
  window.dispatchEvent(new Event("socialai:session-expired"));
}
export async function request(path, { method = "GET", data, signal, auth = true } = {}) {
  try {
    return await axios({ method, url: `${BASE_URL}${path}`, data, signal, timeout: path === "/generate" ? 190000 : 90000,
      headers: auth ? { Authorization: `Bearer ${localStorage.getItem(TOKEN_KEY)}` } : {},
    });
  } catch (error) {
    if (auth && error.response?.status === 401) expireSession();
    throw error;
  }
}
export function errorText(error, fallback = "Something went wrong. Please try again.") {
  const data = error.response?.data;
  if (typeof data === "string" && data.length < 400 && !data.includes("<html")) return data.trim();
  return data?.error || fallback;
}
export async function publishMedia(file, caption) {
  const data = new FormData();
  data.append("message", caption.trim());
  data.append("media_file", file);
  const response = await request("/upload", { method: "POST", data });
  return response.data;
}
