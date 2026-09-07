import React, { useRef, useState } from "react";
import { Modal, Button, message } from "antd";
import { PlusOutlined } from "@ant-design/icons";
import { PostForm, validateMedia } from "./PostForm";
import { errorText, publishMedia } from "../lib/api";
export default function CreatePostButton({ onShowPost, label = "New post", type = "primary" }) {
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const formRef = useRef();
  async function publish() {
    let values;
    try { values = await formRef.current.validateFields(); } catch { return; }
    const file = values.uploadPost[0].originFileObj || values.uploadPost[0];
    const error = validateMedia(file);
    if (error) { message.error(error); return; }
    setBusy(true);
    try {
      const post = await publishMedia(file, values.description);
      const mediaType = /\.(mp4|webm|mov)$/i.test(file.name) ? "video" : "image";
      message.success("Your post is published."); formRef.current.resetFields(); setOpen(false);
      onShowPost?.(mediaType, post);
    } catch (error) { message.error(errorText(error, "Could not publish your post. Please retry.")); }
    finally { setBusy(false); }
  }
  return <><Button type={type} icon={<PlusOutlined aria-hidden="true" />} onClick={() => setOpen(true)}>{label}</Button>
    <Modal title="Share a new moment" open={open} onOk={publish} okText="Publish post" confirmLoading={busy} onCancel={() => !busy && setOpen(false)} cancelButtonProps={{ disabled: busy }} closable={!busy} maskClosable={!busy}>
      <p className="muted">Add an image or video and a few words.</p><PostForm ref={formRef} />
    </Modal></>;
}
