import axios from "axios";
import { session, request } from "./api";
jest.mock("axios", () => ({ __esModule: true, default: jest.fn() }));
const token = claims => `header.${btoa(JSON.stringify(claims))}.signature`;
afterEach(() => { localStorage.clear(); jest.clearAllMocks(); });
test("missing, malformed and expired tokens cannot open protected pages", () => {
  expect(session()).toBeNull(); localStorage.setItem("token", "invalid"); expect(session()).toBeNull();
  localStorage.setItem("token", token({ username: "hazel", exp: Date.now()/1000-10 })); expect(session()).toBeNull();
  localStorage.setItem("token", token({ username: "hazel", exp: Date.now()/1000+100 })); expect(session().username).toBe("hazel");
});
test("a protected 401 ends the local session, while failed login does not", async () => {
  localStorage.setItem("token", "token");
  axios.mockRejectedValue({ response: { status: 401 } });
  await expect(request("/signin", { auth: false })).rejects.toBeDefined(); expect(localStorage.getItem("token")).toBe("token");
  const listener = jest.fn(); window.addEventListener("socialai:session-expired", listener);
  await expect(request("/search")).rejects.toBeDefined(); expect(localStorage.getItem("token")).toBeNull(); expect(listener).toHaveBeenCalledTimes(1);
  window.removeEventListener("socialai:session-expired", listener);
});
