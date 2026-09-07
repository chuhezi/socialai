import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import axios from "axios";
import Landing from "./Landing";
jest.mock("axios", () => ({ __esModule: true, default: jest.fn() }));
jest.mock("./CreatePostButton", () => () => <button>Upload a photo or video</button>);
const image = "data:image/png;base64,dGVzdA==";
const setup = () => render(<MemoryRouter><Landing /></MemoryRouter>);
beforeEach(() => { jest.clearAllMocks(); localStorage.setItem("token", "test-token"); });
async function generate() {
  fireEvent.change(screen.getByLabelText("Your description"), { target: { value: "a watercolor fox" } });
  fireEvent.click(screen.getByRole("button", { name: "Generate image" }));
  return screen.findByAltText("a watercolor fox");
}
test("generates through the protected Go endpoint and displays the result", async () => {
  axios.mockResolvedValueOnce({ status: 200, data: { image, prompt: "a watercolor fox" } }); setup();
  expect(await generate()).toHaveAttribute("src", image);
  expect(axios).toHaveBeenCalledWith(expect.objectContaining({ method: "POST", url: expect.stringMatching(/\/generate$/), data: { prompt: "a watercolor fox" }, headers: { Authorization: "Bearer test-token" } }));
});
test("publishes a generated image with its editable caption and accepts HTTP 201", async () => {
  axios.mockResolvedValueOnce({ status: 200, data: { image, prompt: "a watercolor fox" } }).mockResolvedValueOnce({ status: 201, data: { id: "created" } });
  global.fetch = jest.fn().mockResolvedValue({ blob: async () => new Blob(["test"], { type: "image/png" }) }); setup(); await generate();
  fireEvent.change(screen.getByLabelText("Post caption"), { target: { value: "Fox by the sea" } });
  fireEvent.click(screen.getByRole("button", { name: "Publish to collection" }));
  await waitFor(() => expect(screen.getByRole("button", { name: "Published" })).toBeDisabled());
  const options = axios.mock.calls[1][0];
  expect(options.url).toMatch(/\/upload$/); expect(options.data.get("message")).toBe("Fox by the sea");
  expect(options.data.get("media_file").type).toBe("image/png");
});
test("empty prompts never spend credits and failed generation can be retried", async () => {
  setup(); fireEvent.click(screen.getByRole("button", { name: "Generate image" })); expect(axios).not.toHaveBeenCalled();
  axios.mockRejectedValueOnce({ response: { status: 429, data: "Image API credit or rate limit reached" } });
  fireEvent.change(screen.getByLabelText("Your description"), { target: { value: "a fox" } }); fireEvent.click(screen.getByRole("button", { name: "Generate image" }));
  await waitFor(() => expect(screen.getByRole("button", { name: "Generate image" })).toBeEnabled());
  expect(await screen.findByText("Image API credit or rate limit reached")).toBeInTheDocument();
});
