import React, { useCallback, useEffect, useRef, useState } from "react";
import { Alert, Button, Empty, Input, Modal, Spin, Tabs, message } from "antd";
import { ReloadOutlined, DeleteOutlined } from "@ant-design/icons";
import SearchBar from "./SearchBar";
import PhotoGallery from "./PhotoGallery";
import CreatePostButton from "./CreatePostButton";
import PostMeta from "./PostMeta";
import { SEARCH_KEY } from "../constants";
import { errorText, request, session } from "../lib/api";
import { subscribeToPosts } from "../lib/liveFeed";
import { orderPosts } from "../lib/postOrder";
const PAGE_SIZE = 24;
export default function Collection() {
  const [posts, setPosts] = useState([]);
  const [activeTab, setActiveTab] = useState("image");
  const [option, setOption] = useState({ type: SEARCH_KEY.all, keyword: "" });
  const [limit, setLimit] = useState(PAGE_SIZE);
  const [refresh, setRefresh] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [live, setLive] = useState("connecting");
  const [editing, setEditing] = useState(null);
  const [caption, setCaption] = useState("");
  const [deleting, setDeleting] = useState(null);
  const [busy, setBusy] = useState(false);
  const mutationVersion = useRef(0);
  useEffect(() => {
    const controller = new AbortController();
    let retry;
    const query = new URLSearchParams({ type: activeTab, limit: String(limit) });
    if (option.type === SEARCH_KEY.user && option.keyword) query.set("user", option.keyword);
    else if (option.keyword) query.set("keywords", option.keyword);
    if (option.type === SEARCH_KEY.mine) { query.set("mine", "true"); query.set("user", session()?.username || ""); }
    setLoading(true); setError(""); setLive("connecting");
    const startVersion = mutationVersion.current;
    const connect = async () => {
      try {
        await subscribeToPosts(query, controller.signal, data => { if (!controller.signal.aborted) { setPosts(data); setError(""); } }, setLive);
      } catch (error) { if (controller.signal.aborted) return; if (error.retryable === false) { setLive("manual"); return; } }
      if (!controller.signal.aborted) { setLive("reconnecting"); retry = setTimeout(connect, 5000); }
    };
    request(`/search?${query}`, { signal: controller.signal }).then(({ data }) => {
      if (!controller.signal.aborted && startVersion === mutationVersion.current) setPosts(Array.isArray(data) ? data : []);
    }).catch(error => { if (!controller.signal.aborted) setError(errorText(error, "Could not load your collection.")); }).finally(() => {
      if (!controller.signal.aborted) { setLoading(false); connect(); }
    });
    return () => { controller.abort(); clearTimeout(retry); };
  }, [option, activeTab, limit, refresh]);
  const onChange = useCallback(post => { mutationVersion.current += 1; setPosts(items => items.map(item => item.id === post.id ? post : item)); setRefresh(n => n + 1); }, []);
  const onEdit = post => { setEditing(post); setCaption(post.message); };
  const saveEdit = async () => {
    if (!caption.trim()) { message.warning("Add a caption before saving."); return; }
    setBusy(true);
    try { const { data } = await request(`/post/${encodeURIComponent(editing.id)}`, { method: "PATCH", data: { message: caption.trim() } }); onChange(data); setEditing(null); message.success("Caption updated."); }
    catch (error) { message.error(errorText(error, "Could not save your caption.")); }
    finally { setBusy(false); }
  };
  const removePost = async () => {
    setBusy(true);
    try {
      await request(`/post/${encodeURIComponent(deleting.id)}`, { method: "DELETE" });
      mutationVersion.current += 1; setPosts(items => items.filter(post => post.id !== deleting.id)); setDeleting(null); setRefresh(n => n + 1); message.success("Post deleted.");
    } catch (error) { message.error(errorText(error, "Could not delete your post.")); }
    finally { setBusy(false); }
  };
  const onSearch = next => { setPosts([]); setLimit(PAGE_SIZE); setOption(next); };
  const showPost = type => { setActiveTab(type); setRefresh(n => n + 1); };
  const filtered = orderPosts(posts.filter(post => post.type === activeTab));
  const content = loading && !posts.length ? <div className="loading-state"><Spin size="large" /><p>Gathering your moments…</p></div> : !filtered.length ? <div className="empty-state"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={option.keyword ? "No posts match your search." : `No ${activeTab === "image" ? "images" : "videos"} here yet.`} /><p className="muted">{option.keyword ? "Try another caption or username." : "Share a moment to start the collection."}</p></div> : activeTab === "image" ? <PhotoGallery posts={filtered} onChange={onChange} onEdit={onEdit} onDelete={setDeleting} /> : <div className="post-grid video-grid">{filtered.map(post => <article className="post-card" key={post.id}><video src={post.url} controls playsInline preload="metadata" aria-label={post.message} /><PostMeta post={post} onChange={onChange} onEdit={onEdit} onDelete={setDeleting} /></article>)}</div>;
  return <main className="collection-page page-width">
    <div className="page-heading"><div><p className="eyebrow">THE SHARED GALLERY / 02</p><h1>Moments, <em>collected.</em></h1><p className="intro-copy">Explore what’s new, find a favorite, and share something of your own.</p></div><div className="heading-side"><span className={`live-status ${live === "live" ? "is-live" : ""}`} role="status"><i />{live === "live" ? "Live updates" : live === "connecting" ? "Connecting…" : live === "manual" ? "Refresh to update" : "Reconnecting…"}</span><Button type="text" icon={<ReloadOutlined aria-hidden="true" />} loading={loading} onClick={() => setRefresh(n => n + 1)}>Refresh</Button></div></div>
    <SearchBar handleSearch={onSearch} />
    {error && <Alert className="collection-error" message={error} type="error" showIcon action={<Button size="small" onClick={() => setRefresh(n => n + 1)}>Retry</Button>} />}
    <Tabs className="collection-tabs" activeKey={activeTab} onChange={key => { setActiveTab(key); setPosts([]); setLimit(PAGE_SIZE); }} tabBarExtraContent={<CreatePostButton onShowPost={showPost} />} items={[{ key: "image", label: "Images", children: activeTab === "image" ? content : null }, { key: "video", label: "Videos", children: activeTab === "video" ? content : null }]} />
    {posts.length >= limit && limit < 96 && <div className="load-more"><Button loading={loading} onClick={() => setLimit(n => n + PAGE_SIZE)}>Load more</Button></div>}
    {posts.length >= 96 && <p className="muted">Showing the latest 96 results. Search a caption or username to narrow your collection.</p>}
    <Modal title="Edit your caption" open={Boolean(editing)} onOk={saveEdit} okText="Save changes" confirmLoading={busy} onCancel={() => !busy && setEditing(null)} cancelButtonProps={{ disabled: busy }} closable={!busy} maskClosable={!busy}>
      <label className="field-label" htmlFor="edit-caption">Caption</label><Input.TextArea id="edit-caption" rows={4} value={caption} onChange={event => setCaption(event.target.value)} maxLength={2000} showCount />
    </Modal>
    <Modal className="delete-dialog" title={<span><DeleteOutlined aria-hidden="true" /> Delete this post?</span>} open={Boolean(deleting)} onOk={removePost} okText="Delete post" okButtonProps={{ danger: true }} confirmLoading={busy} onCancel={() => !busy && setDeleting(null)} cancelButtonProps={{ disabled: busy }} closable={!busy} maskClosable={!busy}>
      {deleting?.type === "image" && <img className="delete-preview" src={deleting.url} alt={deleting.message} />}<p>This removes your post and its media from the collection. This cannot be undone.</p><blockquote>{deleting?.message}</blockquote>
    </Modal>
  </main>;
}
