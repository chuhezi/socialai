import { orderPosts } from "./postOrder";

test("undated numbered captions use natural order, without mutating the response", () => {
  const posts = [6, 10, 2, 1].map(n => ({ id: String(20 - n), message: `hazel painting${n}` }));
  expect(orderPosts(posts).map(post => post.message)).toEqual([
    "hazel painting1", "hazel painting2", "hazel painting6", "hazel painting10",
  ]);
  expect(posts[0].message).toBe("hazel painting6");
});

test("new posts stay first and edits or reactions do not change their chronology", () => {
  const posts = [
    { id: "legacy", message: "painting1" },
    { id: "older", createdAt: 100, updatedAt: 900, message: "A revised caption", likes: ["123"] },
    { id: "newer", createdAt: 200, message: "Z coast" },
  ];
  expect(orderPosts(posts).map(post => post.id)).toEqual(["newer", "older", "legacy"]);
  expect(orderPosts([...posts].reverse()).map(post => post.id)).toEqual(["newer", "older", "legacy"]);
});
