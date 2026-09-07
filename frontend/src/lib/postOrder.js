const captions = new Intl.Collator("en", { numeric: true, sensitivity: "base" });

const publishedAt = post => {
  const value = Number(post.createdAt);
  return Number.isFinite(value) && value > 0 ? value : 0;
};

// Dated posts stay newest first. Classroom posts without dates use their
// captions, so numbered series read 1, 2, 3 rather than random UUID order.
export function orderPosts(posts) {
  return [...posts].sort((a, b) => {
    const aTime = publishedAt(a);
    const bTime = publishedAt(b);
    if (aTime !== bTime) return bTime - aTime;
    if (!aTime) {
      const captionOrder = captions.compare(a.message || "", b.message || "");
      if (captionOrder) return captionOrder;
      const userOrder = captions.compare(a.user || "", b.user || "");
      if (userOrder) return userOrder;
    }
    return String(a.id).localeCompare(String(b.id));
  });
}
