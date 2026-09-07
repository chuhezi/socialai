import React, { useState } from "react";
import { Form, Input, Button, message } from "antd";
import { Link } from "react-router-dom";
import { request, errorText } from "../lib/api";
export default function Login({ handleLoggedIn }) {
  const [busy, setBusy] = useState(false);
  async function onFinish(values) {
    setBusy(true);
    try { const { data } = await request("/signin", { method: "POST", data: { ...values, username: values.username.trim() }, auth: false }); handleLoggedIn(data); message.success("Welcome back."); }
    catch (error) { message.error(errorText(error, "Could not log in. Check your details and try again.")); }
    finally { setBusy(false); }
  }
  return <main className="auth-page page-width"><section className="auth-intro"><p className="eyebrow">A SMALL SPACE. A WORLD OF IDEAS.</p><h1>Make room<br />for <em>imagination.</em></h1><p>Create with AI, collect your moments,<br />and see the world through a new lens.</p><div className="auth-note">01 / CREATE <span>02 / SHARE</span> 03 / CONNECT</div></section>
    <section className="auth-card"><p className="eyebrow">GOOD TO SEE YOU</p><h2>Welcome back.</h2><p className="muted">Log in to your creative space.</p><Form name="login" layout="vertical" onFinish={onFinish} requiredMark={false}>
      <Form.Item name="username" label="Username" rules={[{ required: true, whitespace: true, message: "Enter your username." }]}><Input autoComplete="username" placeholder="Your username" /></Form.Item>
      <Form.Item name="password" label="Password" rules={[{ required: true, message: "Enter your password." }]}><Input.Password autoComplete="current-password" placeholder="Your password" /></Form.Item>
      <Button block type="primary" htmlType="submit" loading={busy}>Log in</Button>
    </Form><p className="auth-switch">New here? <Link to="/register">Create an account →</Link></p></section></main>;
}
