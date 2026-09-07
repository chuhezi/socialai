import React, { useState } from "react";
import { Input, Radio } from "antd";
import { SEARCH_KEY } from "../constants";
export default function SearchBar({ handleSearch }) {
  const [type, setType] = useState(SEARCH_KEY.all);
  const [value, setValue] = useState("");
  const [error, setError] = useState("");
  const submit = text => {
    if (type === SEARCH_KEY.user && !text.trim()) { setError("Enter a username to search."); return; }
    setError(""); handleSearch({ type, keyword: text.trim() });
  };
  return <section className="search-bar" aria-label="Find posts">
    <Input.Search aria-label="Search posts" placeholder={type === SEARCH_KEY.user ? "Search by username…" : "Search captions, ideas, moments…"} value={value} onChange={e => setValue(e.target.value)} onSearch={submit} allowClear enterButton="Search" size="large" maxLength={2000} />
    <div className="search-options"><Radio.Group aria-label="Search mode" value={type} onChange={e => { const next = e.target.value; setType(next); setError(""); setValue(""); handleSearch({ type: next, keyword: "" }); }}>
      <Radio value={SEARCH_KEY.all}>All posts</Radio><Radio value={SEARCH_KEY.keyword}>Keyword</Radio><Radio value={SEARCH_KEY.user}>User</Radio><Radio value={SEARCH_KEY.mine}>My posts</Radio>
    </Radio.Group><span className="sort-label" title="Newest dated posts first. Undated posts in the loaded results follow caption order, including numbered series.">Newest first · undated by caption</span></div>
    {error && <p className="error-msg" role="alert">{error}</p>}
  </section>;
}
