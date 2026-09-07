import React, { forwardRef } from "react";
import { Form, Upload, Input, message } from "antd";
import { InboxOutlined } from "@ant-design/icons";
export const ACCEPTED_MEDIA = ".jpg,.jpeg,.png,.gif,.webp,.mp4,.webm,.mov";
export function validateMedia(file) {
  if (!/\.(jpe?g|png|gif|webp|mp4|webm|mov)$/i.test(file.name)) return "Choose JPG, PNG, GIF, WebP, MP4, WebM or MOV.";
  if (file.size <= 0 || file.size > 10 * 1024 * 1024) return "Choose a file up to 10 MB.";
  return null;
}
export const PostForm = forwardRef((props, formRef) => <Form layout="vertical" ref={formRef}>
  <Form.Item name="description" label="Caption" rules={[{ required: true, whitespace: true, message: "Add a caption for your post." }]}>
    <Input.TextArea placeholder="Tell the story behind your post…" rows={3} maxLength={2000} showCount />
  </Form.Item>
  <Form.Item name="uploadPost" label="Image or video" valuePropName="fileList" getValueFromEvent={event => Array.isArray(event) ? event : event?.fileList} rules={[{ required: true, message: "Choose an image or video." }]}>
    <Upload.Dragger name="media" accept={ACCEPTED_MEDIA} maxCount={1} listType="picture" beforeUpload={file => { const error = validateMedia(file); if (error) { message.error(error); return Upload.LIST_IGNORE; } return false; }}>
      <p className="ant-upload-drag-icon"><InboxOutlined aria-hidden="true" /></p>
      <p className="ant-upload-text">Drop a moment here</p>
      <p className="ant-upload-hint">or click to choose · One file, up to 10 MB<br />Images, MP4, WebM or MOV</p>
    </Upload.Dragger>
  </Form.Item>
</Form>);
