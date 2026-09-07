import React from "react";
import { NavLink, Link } from "react-router-dom";
import { Button } from "antd";
import { LogoutOutlined } from "@ant-design/icons";
import { session } from "../lib/api";
export default function ResponsiveAppBar({ isLoggedIn, handleLogout, loggingOut }) {
  return <header className="site-header"><div className="nav-inner">
    <Link className="brand" to={isLoggedIn ? "/collection" : "/login"} aria-label="Social AI home"><span className="brand-mark">S<span>✳</span></span>Social AI<span className="brand-dot">.</span></Link>
    {isLoggedIn ? <>
      <nav aria-label="Main navigation"><NavLink to="/create">Create</NavLink><NavLink to="/collection">Collection</NavLink></nav>
      <div className="account-nav"><span className="account-name">@{session()?.username}</span><Button icon={<LogoutOutlined aria-hidden="true" />} onClick={handleLogout} loading={loggingOut}>Log out</Button></div>
    </> : <span className="nav-note">CREATE. SHARE. CONNECT.</span>}
  </div></header>;
}
