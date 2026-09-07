import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import Collection from "./Collection";
import { request } from "../lib/api";
import { subscribeToPosts } from "../lib/liveFeed";
jest.mock("../lib/api", () => ({ request: jest.fn(), errorText: (e, text) => text, session: () => ({ username: "hazel" }) }));
jest.mock("../lib/liveFeed", () => ({ subscribeToPosts: jest.fn() }));
jest.mock("./CreatePostButton", () => () => <button>New post</button>);
jest.mock("./PhotoGallery", () => ({ posts, onEdit, onDelete }) => <div>{posts.map(post => <article key={post.id}>{post.message}<button onClick={() => onEdit(post)}>Edit {post.id}</button><button onClick={() => onDelete(post)}>Delete {post.id}</button></article>)}</div>);
const post = { id: "1", user: "hazel", message: "Mint coast", type: "image", url: "/photo.png" };
beforeEach(() => { jest.clearAllMocks(); subscribeToPosts.mockImplementation(() => new Promise(() => {})); request.mockResolvedValue({ data: [post] }); });
test("deletion is confirmed and a successful delete is removed from the parent feed", async () => {
  render(<Collection />); await screen.findByText("Mint coast"); fireEvent.click(screen.getByRole("button", { name: "Delete 1" }));
  expect(request).toHaveBeenCalledTimes(1); expect(screen.getByRole("dialog")).toHaveTextContent("Delete this post?");
  fireEvent.click(screen.getByRole("button", { name: "Cancel" })); expect(request).toHaveBeenCalledTimes(1);
  fireEvent.click(screen.getByRole("button", { name: "Delete 1" })); request.mockResolvedValueOnce({ status: 204 }).mockResolvedValue({ data: [] });
  fireEvent.click(screen.getByRole("button", { name: "Delete post" }));
  await waitFor(() => expect(request).toHaveBeenCalledWith("/post/1", { method: "DELETE" }));
  await waitFor(() => expect(screen.queryByRole("button", { name: "Delete 1" })).not.toBeInTheDocument());
});
test("server events update the visible feed and special characters are encoded in search", async () => {
  render(<Collection />); await screen.findByText("Mint coast");
  await act(async () => subscribeToPosts.mock.calls[0][2]([{ ...post, message: "Edited in another window" }]));
  expect(screen.getByText("Edited in another window")).toBeInTheDocument();
  fireEvent.change(screen.getByRole("textbox", { name: "Search posts" }), { target: { value: "sea & sky" } });
  fireEvent.click(screen.getByRole("button", { name: "Search" }));
  await waitFor(() => expect(request).toHaveBeenCalledWith(expect.stringContaining("keywords=sea+%26+sky"), expect.anything()));
});
test("my posts sends the authenticated scope to both current and classroom backends", async () => {
  render(<Collection />); await screen.findByText("Mint coast");
  fireEvent.click(screen.getByRole("radio", { name: "My posts" }));
  await waitFor(() => expect(request).toHaveBeenCalledWith(expect.stringContaining("mine=true&user=hazel"), expect.anything()));
});

test("numbered legacy posts keep their display order after a live update", async () => {
  const items = [6, 2, 1].map(n => ({ ...post, id: String(n), message: `hazel painting${n}` }));
  request.mockResolvedValue({ data: items });
  render(<Collection />);
  await screen.findByText("hazel painting6");
  const captions = () => screen.getAllByRole("article").map(article => article.textContent.match(/hazel painting\d+/)[0]);
  expect(captions()).toEqual(["hazel painting1", "hazel painting2", "hazel painting6"]);
  await act(async () => subscribeToPosts.mock.calls[0][2]([...items].reverse()));
  expect(captions()).toEqual(["hazel painting1", "hazel painting2", "hazel painting6"]);
});
