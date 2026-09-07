import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Form, Input, Button, message } from "antd";
import { request, errorText } from "../lib/api";
export default function Register() {
  const [busy, setBusy] = useState(false);
  const navigate = useNavigate();
  async function onFinish({ username, password }) {
    setBusy(true);
    try { await request("/signup", { method: "POST", data: { username, password }, auth: false }); message.success("Account created. You can log in now."); navigate("/login"); }
    catch (error) { message.error(errorText(error, "Could not create your account. Please retry.")); }
    finally { setBusy(false); }
  }
  return <main className="auth-page page-width"><section className="auth-intro"><p className="eyebrow">YOUR NEXT CHAPTER STARTS HERE</p><h1>A space<br />to be <em>yourself.</em></h1><p>A few details, a little curiosity.<br />You’re ready to create something new.</p><div className="auth-note">01 / CREATE <span>02 / SHARE</span> 03 / CONNECT</div></section>
    <section className="auth-card"><p className="eyebrow">JOIN THE COLLECTION</p><h2>Start creating.</h2><Form layout="vertical" onFinish={onFinish} requiredMark={false}>
      <Form.Item name="username" label="Username" rules={[{ required: true, message: "Choose a username." }, { pattern: /^[a-zA-Z0-9_]{3,32}$/, message: "Use 3–32 letters, numbers or underscores." }]}><Input autoComplete="username" placeholder="Choose a username" maxLength={32} /></Form.Item>
      <Form.Item name="password" label="Password" rules={[{ required: true, message: "Choose a password." }, { min: 8, message: "Use at least 8 characters." }]}><Input.Password autoComplete="new-password" placeholder="At least 8 characters" /></Form.Item>
      <Form.Item name="confirm" label="Confirm password" dependencies={["password"]} rules={[{ required: true, message: "Confirm your password." }, ({ getFieldValue }) => ({ validator: (_, value) => !value || getFieldValue("password") === value ? Promise.resolve() : Promise.reject(new Error("The passwords do not match.")) })]}><Input.Password autoComplete="new-password" placeholder="Enter it once more" /></Form.Item>
      <Button block type="primary" htmlType="submit" loading={busy}>Create account</Button>
    </Form><p className="auth-switch">Already a member? <Link to="/login">Log in →</Link></p></section></main>;
}
