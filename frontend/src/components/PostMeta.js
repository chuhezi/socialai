import React, { useState } from "react";
import { Button, Dropdown, message } from "antd";
import { HeartOutlined, HeartFilled, MoreOutlined, EditOutlined, DeleteOutlined } from "@ant-design/icons";
import { request, errorText, session } from "../lib/api";
export default function PostMeta({ post, onChange, onEdit, onDelete }) {
  const [busy, setBusy] = useState(false);
  const username = session()?.username;
  const liked = post.likes?.includes(username);
  const toggleLike = async () => {
    setBusy(true);
    try { const { data } = await request(`/post/${encodeURIComponent(post.id)}/like`, { method: "PUT", data: { liked: !liked } }); onChange(data); }
    catch (error) { message.error(errorText(error, "Could not update your like.")); }
    finally { setBusy(false); }
  };
  return <div className="post-meta">
    <div className="post-byline"><span className="avatar-letter">{post.user?.slice(0, 1).toUpperCase()}</span><div><strong>@{post.user}</strong><time dateTime={post.createdAt ? new Date(post.createdAt).toISOString() : undefined}>{post.createdAt ? new Date(post.createdAt).toLocaleDateString("en-US", { month: "short", day: "numeric" }) : "Shared moment"}{post.updatedAt ? " · edited" : ""}</time></div>
      {post.user === username && <Dropdown trigger={["click"]} menu={{ items: [{ key: "edit", label: "Edit caption", icon: <EditOutlined aria-hidden="true" /> }, { key: "delete", label: "Delete post", danger: true, icon: <DeleteOutlined aria-hidden="true" /> }], onClick: ({ key }) => key === "edit" ? onEdit(post) : onDelete(post) }}><Button className="post-menu" type="text" aria-label={`Post options for ${post.message}`} icon={<MoreOutlined aria-hidden="true" />} /></Dropdown>}
    </div>
    <p className="post-caption">{post.message}</p>
    <Button type="text" className={`like-button ${liked ? "is-liked" : ""}`} aria-label={`${liked ? "Unlike" : "Like"} post: ${post.message}`} aria-pressed={Boolean(liked)} icon={liked ? <HeartFilled aria-hidden="true" /> : <HeartOutlined aria-hidden="true" />} loading={busy} onClick={toggleLike}>{post.likes?.length || 0}<span className="like-label">{post.likes?.length === 1 ? "like" : "likes"}</span></Button>
  </div>;
}
