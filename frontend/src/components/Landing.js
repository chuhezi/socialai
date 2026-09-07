import React, { useRef, useState } from "react";
import { Button, Input, message } from "antd";
import { ArrowRightOutlined, PictureOutlined, DownloadOutlined, CheckOutlined } from "@ant-design/icons";
import { Link, useNavigate } from "react-router-dom";
import { errorText, publishMedia, request } from "../lib/api";
import CreatePostButton from "./CreatePostButton";
const ideas = ["A quiet seaside town in watercolor", "A mint-green volleyball court at sunset", "A tiny café tucked inside a cloud"];
export default function Landing() {
  const [prompt, setPrompt] = useState("");
  const [result, setResult] = useState(null);
  const [caption, setCaption] = useState("");
  const [generating, setGenerating] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [published, setPublished] = useState(false);
  const inFlight = useRef(false);
  const navigate = useNavigate();
  const generate = async event => {
    event.preventDefault();
    if (!prompt.trim()) { message.warning("Describe the image you have in mind."); return; }
    if (inFlight.current) return;
    inFlight.current = true; setGenerating(true);
    try {
      const { data } = await request("/generate", { method: "POST", data: { prompt: prompt.trim() } });
      if (!data.image?.startsWith("data:image/")) throw new Error("No image returned");
      setResult(data); setCaption((data.prompt || prompt.trim()).slice(0, 2000)); setPublished(false);
    } catch (error) { message.error(errorText(error, "Could not generate an image. Please try again."), 6); }
    finally { setGenerating(false); inFlight.current = false; }
  };
  const publish = async () => {
    if (!caption.trim()) { message.warning("Add a caption before publishing."); return; }
    setPublishing(true);
    try {
      const blob = await (await fetch(result.image)).blob();
      await publishMedia(new File([blob], "socialai-generated.png", { type: "image/png" }), caption);
      setPublished(true); message.success("Your image is now in the collection.");
    } catch (error) { message.error(errorText(error, "Could not publish your image.")); }
    finally { setPublishing(false); }
  };
  return <main className="create-page page-width">
    <div className="studio-heading"><p className="eyebrow">THE CREATIVE STUDIO / 01</p><h1>A little imagination.<br /><em>A new perspective.</em></h1><p className="intro-copy">Turn an idea into an image. Make it yours. Share it with the world.</p></div>
    <section className="studio-panel" aria-label="Image studio">
      <div className="prompt-panel"><span className="section-number">01 / IMAGINE</span><h2>What’s on your mind?</h2><p className="muted">Describe a place, a feeling, or something entirely new.</p>
        <form onSubmit={generate}><label className="field-label" htmlFor="image-prompt">Your description</label><Input.TextArea id="image-prompt" placeholder="A sunlit seaside café, mint-green chairs, soft watercolor…" value={prompt} onChange={event => setPrompt(event.target.value)} maxLength={4000} rows={5} disabled={generating} />
          <Button type="primary" htmlType="submit" block loading={generating} disabled={publishing} icon={<ArrowRightOutlined aria-hidden="true" />}>{generating ? "Creating your image…" : "Generate image"}</Button>
        </form><div className="prompt-ideas"><span className="field-label">A spark to get you started</span>{ideas.map(idea => <button key={idea} disabled={generating} onClick={() => setPrompt(idea)}>{idea}<span>↗</span></button>)}</div>
        <div className="upload-alternative"><p>Already have something to share?</p><CreatePostButton type="default" label="Upload a photo or video" onShowPost={() => navigate("/collection")} /></div>
      </div>
      <div className="result-panel"><span className="section-number">02 / MAKE IT YOURS</span>
        {result ? <><div className="generated-preview"><img src={result.image} alt={result.prompt} /></div><label className="field-label" htmlFor="generated-caption">Post caption</label><Input.TextArea id="generated-caption" value={caption} onChange={event => setCaption(event.target.value)} rows={2} maxLength={2000} disabled={published || publishing} />
          <div className="result-actions"><Button type="primary" loading={publishing} disabled={published || generating} icon={published ? <CheckOutlined aria-hidden="true" /> : <ArrowRightOutlined aria-hidden="true" />} onClick={publish}>{published ? "Published" : "Publish to collection"}</Button><a className="download-link" href={result.image} download="socialai-generated.png"><DownloadOutlined aria-hidden="true" /> Download</a></div>{published && <Link className="collection-link" to="/collection">View your collection →</Link>}</>
          : <div className={`studio-empty ${generating ? "is-generating" : ""}`} aria-live="polite"><span className="preview-icon"><PictureOutlined aria-hidden="true" /></span><h3>{generating ? "A new idea is taking shape." : "A fresh canvas for your ideas."}</h3><p>{generating ? "This can take a little while. Your image will appear here." : "Your generated image will appear here, ready for a caption and a place in the collection."}</p><span className="canvas-note">IMAGINE SOMETHING GOOD.</span></div>}
      </div>
    </section>
  </main>;
}
