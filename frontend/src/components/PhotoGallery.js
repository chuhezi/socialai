import React, { useState } from "react";
import { DeleteOutlined, EditOutlined, ExpandOutlined } from "@ant-design/icons";
import Lightbox from "yet-another-react-lightbox";
import Fullscreen from "yet-another-react-lightbox/plugins/fullscreen";
import Slideshow from "yet-another-react-lightbox/plugins/slideshow";
import Thumbnails from "yet-another-react-lightbox/plugins/thumbnails";
import Zoom from "yet-another-react-lightbox/plugins/zoom";
import PostMeta from "./PostMeta";
import { session } from "../lib/api";
export default function PhotoGallery({ posts, onChange, onEdit, onDelete }) {
  const [selectedID, setSelectedID] = useState(null);
  const index = posts.findIndex(post => post.id === selectedID);
  const selected = posts[index];
  const own = selected?.user === session()?.username;
  const slides = posts.map(post => ({ src: post.url, alt: post.message }));
  const action = callback => { const post = selected; setSelectedID(null); callback(post); };
  return <>
    <div className="post-grid" aria-label="Image posts">
      {posts.map(post => <article key={post.id} className="post-card">
        <button className="photo-trigger" onClick={() => setSelectedID(post.id)} aria-label={`View image: ${post.message}`}>
          <img src={post.url} alt={post.message} loading="lazy" decoding="async" />
          <span className="expand-hint"><ExpandOutlined aria-hidden="true" /> View</span>
        </button>
        <PostMeta post={post} onChange={onChange} onEdit={onEdit} onDelete={onDelete} />
      </article>)}
    </div>
    <Lightbox slides={slides} open={index >= 0} index={Math.max(0, index)} close={() => setSelectedID(null)} plugins={[Fullscreen, Slideshow, Thumbnails, Zoom]} on={{ view: ({ index: next }) => setSelectedID(posts[next]?.id || null) }} toolbar={{ buttons: [
      ...(own ? [<button className="yarl__button" key="edit" aria-label="Edit caption" title="Edit caption" onClick={() => action(onEdit)}><EditOutlined aria-hidden="true" /></button>, <button className="yarl__button" key="delete" aria-label="Delete post" title="Delete post" onClick={() => action(onDelete)}><DeleteOutlined aria-hidden="true" /></button>] : []), "close",
    ] }} />
  </>;
}
